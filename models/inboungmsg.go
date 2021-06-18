package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

type AppDetails struct {
	Bunit_id          int    `json:bunit_id`
	Store_type        string `json:store_type`
	App_name          string `json:app_name`
	App_key           string `json:app_key`
	Brand_name        string `json:brand_name`
	Source_number     int    `json:source_number`
	Store_loc_api_key string `json:store_loc_api_key`
	Category          int    `json:category`
}

type InboundBody struct {
	App       string `json:app`
	Timestamp int    `json:timestamp`
	Version   int    `json:version`
	Type      string `json:type`
	Payload   struct {
		ID      string `json:id`
		Source  string `json:source`
		Type    string `json:type`
		Payload struct {
			Text      string `json:text`
			Longitude string `json:longitude`
			Latitude  string `json:latitude`
		} `json:payload`
		Sender struct {
			Phone        string `json:phone`
			Name         string `json:name`
			Country_code string `json:country_code`
			Dial_code    string `json:dial_code`
		} `json:sender`
		Context struct {
			Id   string `json:id`
			GsId string `json:gsId`
		}
	} `json:payload`
}

type Item struct {
	Id           int    `json:id`
	Response_msg string `json:response_msg`
	Request_msg  string `json:request_msg`
}

type RespMsg struct {
	MessageId string `json:messageId`
	Status    string `json:status`
}

type NlpResp struct {
	Entity string `json:entity`
	Text   string `json:text`
}
type Productdetail struct {
	Id          int    `json:id`
	Name        string `json:name`
	Product_url string `json:product_url`
	Short_url   string `json:short_url`
}
type StoreDetail struct {
	// Lr         string `json:lr`
	SearchType string      `json:search_type`
	Count      int         `json:count`
	Stores     []Storelist `json:stores`
}

type Storelist struct {
	Id          int    `json:id`
	Dealer_id   string `json:dealer_id`
	Wp_name     string `json:wp_name`
	Category_id int    `json:category_id`
	Distance    string `json:distance`
	//distance      string `json:distance`
	Is_routeable  bool   `json:is_routeable`
	Wp_id         int    `json:wp_id`
	Phone_no      string `json:phone_no`
	Email_id      string `json:email_id`
	Address       string `json:address`
	City          string `json:city`
	State         string `json:state`
	Store_timings string `json:store_timings`
	Store_open    bool   `json:store_open`
	Geofeature    struct {
		Geoproperty string `json:geoproperty`
	} `json:geofeature`
}

type Shorturl struct {
	Short_url string `json:short_url`
}

type reversegeocodeforstore struct {
	Data revgeodat `json:data`
}

type revgeodat struct {
	//Poi      string           `json:poi`
	//Road     string           `json:road`
	Locality []revgeolocality `json:locality`
	Pincode  string           `json:pincode`
}

type revgeolocality struct {
	Id    int    `json:id`
	Order int    `json:order`
	Type  string `json:type`
	Name  string `json:name`
}

type chat_history struct {
	Id                   int    `json:id`
	Message_received     string `json:message_received`
	Message_sent         string `json:message_sent`
	Sender_mobile_number int    `json:sender_mobile_number`
	Bunit_id             int    `json:bunit_id`
}

func (u *InboundBody) Process(conn *pgx.Conn) error {
	var appdetail AppDetails
	var wherequery string
	var pincode_available bool
	destination_number := u.Payload.Sender.Phone
	sender_name := u.Payload.Sender.Name
	payload_type := u.Payload.Type
	rows, err := conn.Query(context.Background(), "select bunit_id,store_type,app_name,app_key,brand_name,source_number,store_loc_api_key,category from bunit_config where app_name = $1", u.App)
	fmt.Println(rows)
	if err != nil {
		log.Println("--------------***Error in method-Process > finding app query-----*******", rows, err)
		return nil
	}

	for rows.Next() {
		err = rows.Scan(&appdetail.Bunit_id, &appdetail.Store_type, &appdetail.App_name, &appdetail.App_key, &appdetail.Brand_name, &appdetail.Source_number, &appdetail.Store_loc_api_key, &appdetail.Category)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("-------------------------in loop-----------", appdetail)
	}

	fmt.Println("------------------------------------", appdetail)

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")

	processedtext := reg.ReplaceAllString(u.Payload.Payload.Text, " ")

	if u.Payload.Type == "text" {
		welcomenotefinder(processedtext, u.Payload.Payload.Text, destination_number, sender_name, payload_type, conn, appdetail.Bunit_id, appdetail.Brand_name, appdetail.App_name, appdetail.App_key, appdetail.Source_number, appdetail.Store_loc_api_key)
		res := Latlongnlp(processedtext, strconv.Itoa(appdetail.Category), destination_number)
		for _, v := range res {
			if v.Entity == "product" {
				wherequery = "name ilike '%" + v.Text + "%' and bunit_id = " + strconv.Itoa(appdetail.Bunit_id)
				fmt.Println("Value: ", v.Text)
				for _, v := range res {
					if v.Entity == "category" {
						wherequery = wherequery + " and category ilike '%" + v.Text + "%' limit 1"
					}
				}
				findproduct(wherequery, u.Payload.Payload.Text, destination_number, sender_name, payload_type, conn, appdetail.Bunit_id, appdetail.Brand_name, appdetail.App_name, appdetail.App_key, appdetail.Source_number, appdetail.Store_loc_api_key)
			} else if v.Entity == "category" {
				wherequery = " bunit_id = " + strconv.Itoa(appdetail.Bunit_id)
				for _, v := range res {
					if v.Entity == "product" {
						return nil
					} else if v.Entity == "category" {
						wherequery = wherequery + " and category ilike '%" + v.Text + "%' limit 1"
						findproduct(wherequery, u.Payload.Payload.Text, destination_number, sender_name, payload_type, conn, appdetail.Bunit_id, appdetail.Brand_name, appdetail.App_name, appdetail.App_key, appdetail.Source_number, appdetail.Store_loc_api_key)
					}
				}
			} else if v.Entity == "location" || v.Entity == "pincode" {
				for _, v := range res {
					if v.Entity == "pincode" {
						pincode_available = true
					}
				}
				if strings.Contains(v.Text, "me") || strings.Contains(v.Text, "store") || strings.Contains(v.Text, "location") || strings.Contains(v.Text, "near") || strings.Contains(v.Text, "which") || strings.Contains(v.Text, "about") {
					if strings.Contains(v.Text, "me") && !strings.Contains(v.Text, "merchant") {
						message := "Am unable to find nearest " + appdetail.Store_type + " using area name, Please share your location"
						gubshubcall(conn, u.Payload.Payload.Text, appdetail.Bunit_id, appdetail.Brand_name, message, destination_number, appdetail.App_key, appdetail.App_name, sender_name, payload_type, appdetail.Source_number)
					}
				} else {
					if v.Entity == "location_prefix" {
						//message := strings.Replace(v.Text, "near", "", -1)
					}
					var message []string
					if pincode_available != true {
						if v.Entity == "location" {
							message = neareststore(conn, v.Text, appdetail.Brand_name, appdetail.Bunit_id, appdetail.Store_loc_api_key, destination_number, appdetail.Store_type)
							fmt.Println("-----------1--------1---------1")
						}
					} else {
						if v.Entity == "pincode" {
							fmt.Println("-----------2--------2---------2")
							message = neareststore(conn, v.Text, appdetail.Brand_name, appdetail.Bunit_id, appdetail.Store_loc_api_key, destination_number, appdetail.Store_type)
						}
					}

					if len(message) > 0 {
						messagefirst := "Please wait am finding " + appdetail.Store_type + " for you :-)"
						gubshubcall(conn, u.Payload.Payload.Text, appdetail.Bunit_id, appdetail.Brand_name, messagefirst, destination_number, appdetail.App_key, appdetail.App_name, sender_name, payload_type, appdetail.Source_number)
					} else if pincode_available != true {
						message = append(message, "Am unable to find nearest "+appdetail.Store_type+" using area name, Please share your location")
					} else {
						message = append(message, fmt.Sprintf("Am unable to find you the nearest "+appdetail.Store_type+" using the area/city name. I request you to share you location or send me the pincode."))
					}
					for i := 0; i < len(message); i++ {
						gubshubcall(conn, u.Payload.Payload.Text, appdetail.Bunit_id, appdetail.Brand_name, message[i], destination_number, appdetail.App_key, appdetail.App_name, sender_name, payload_type, appdetail.Source_number)
						fmt.Println(message)
					}
				}
			} else {
				err_msg_required := true
				for _, v := range res {
					if v.Entity == "product" || v.Entity == "category" || v.Entity == "location" || v.Entity == "pincode" {
						err_msg_required = false
					}
				}
				if v.Entity == "random" && !strings.Contains(v.Text, "hello") && err_msg_required == true {
					errormessage(conn, u.Payload.Payload.Text, appdetail.Bunit_id, appdetail.Brand_name, destination_number, appdetail.App_key, appdetail.App_name, sender_name, payload_type, appdetail.Source_number)
				}
			}
		}
	} else if u.Payload.Type == "location" {
		fmt.Println("--------------location-------------------")
		message := "Please wait am finding nearest " + appdetail.Store_type + " to you :-)"
		gubshubcall(conn, u.Payload.Payload.Text, appdetail.Bunit_id, appdetail.Brand_name, message, destination_number, appdetail.App_key, appdetail.App_name, sender_name, payload_type, appdetail.Source_number)
		Payloadtextlocation(u, conn, appdetail.Bunit_id, appdetail.Brand_name, appdetail.App_name, appdetail.App_key, appdetail.Source_number, appdetail.Store_loc_api_key, appdetail.Store_type)
	}
	return nil
}

func Payloadtextlocation(u *InboundBody, conn *pgx.Conn, bunit_id int, brand_name string, app_name string, app_key string, source_number int, store_api_key string, store_type string) error {
	destination_number := u.Payload.Sender.Phone
	lat := u.Payload.Payload.Latitude
	long := u.Payload.Payload.Longitude
	var message []string
	var store StoreDetail
	var messageadd string
	rev := reversegeocodeforstore{}
	res, err := http.Get("https://api.latlong.in/v2/brands/" + strconv.Itoa(bunit_id) + "/stores_around.json?lat=" + lat + "&long=" + long + "&access_token=" + store_api_key)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(body, &store)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("------*********", store)
	res1, err1 := http.Get("https://api.latlong.in/v2/reverse_geocode.json?access_token=" + store_api_key + "&latitude=" + lat + "&longitude=" + long + "&more_info=false")
	defer res1.Body.Close()
	body1, err1 := ioutil.ReadAll(res1.Body)
	if err != nil {
		log.Println(err1)
	}

	fmt.Println("----------------------", res1, body1)

	err2 := json.Unmarshal(body1, &rev)
	if err2 != nil {
		fmt.Println(err2)
	}
	for i := 0; i < len(rev.Data.Locality); i++ {
		if rev.Data.Locality[i].Order <= 9 {
			previousmsg := messageadd
			if len(previousmsg) <= 0 {
				messageadd = rev.Data.Locality[i].Name
			} else {
				messageadd = rev.Data.Locality[i].Name
				messageadd = previousmsg + "," + messageadd
			}
			//fmt.Println(messageadd)
		}
	}
	messageadd = "*" + messageadd + " - " + rev.Data.Pincode + "*"

	if store.Stores != nil {
		//fmt.Println(store.Stores[0].Geofeature.Geoproperty)
		point := store.Stores[0].Geofeature.Geoproperty
		t := strings.Replace(point, "(", "", -1)
		t = strings.Replace(t, ")", "", -1)
		t = strings.Replace(t, "POINT", "", -1)
		f := strings.Fields(t)
		var s Shorturl
		res2, err2 := http.Get("https://api.latlong.in/v2/short_urls/create_short_url.json?url_string=https://www.dellretailstores.in/single_store/landing_page/" + strconv.Itoa(store.Stores[0].Id) + "&access_token=" + store_api_key + "&latitude")
		body1, err1 := ioutil.ReadAll(res2.Body)
		if err1 != nil {
			log.Println(err1)
		}
		err2 = json.Unmarshal(body1, &s)
		if err2 != nil {
			fmt.Println(err2)
		}
		message = append(message, store_type+" near - "+messageadd+".\r\n*"+store.Stores[0].Wp_name+"*\r\n"+store.Stores[0].Address+"\r\n*Ph* : "+store.Stores[0].Phone_no+"\r\nStore is just *"+store.Stores[0].Distance+" Km* from your Place.")
		if bunit_id == 242 {
			message = append(message, "\r\n*For more Details*\r\nhttps://llmap.in/"+s.Short_url)
		}
		message = append(message, "For detailed direction click below,")
		message = append(message, "{type: location,longitude: "+f[0]+",latitude: "+f[1]+",name: "+store.Stores[0].Wp_name+",address: '"+store.Stores[0].Address+"'}")
		row := conn.QueryRow(context.Background(), "select id,message_received,message_sent,sender_mobile_number,bunit_id from chat_history where sender_mobile_number = '"+destination_number+"' and created_at >= now()-INTERVAL '5 MINUTE' and message_sent in ('https://www.prestigexclusive.in/cookware','https://www.prestigexclusive.in/mixer-grinders','https://www.prestigexclusive.in/SearchResults.aspx?search=MIXER%20GRINDERS')")
		var chat_history chat_history
		fmt.Println("***************************************************")
		fmt.Println(row)
		err := row.Scan(&chat_history.Id, &chat_history.Message_received, &chat_history.Message_sent, &chat_history.Sender_mobile_number, &chat_history.Bunit_id)
		if err != nil {
			fmt.Println(err)
		} else {
			if chat_history.Bunit_id == 271 {
				link_id := "1"
				if chat_history.Message_sent == "https://www.prestigexclusive.in/cookware" {
					link_id = "1"
				} else {
					link_id = "2"
				}
				fmt.Println(chat_history)
				message = append(message, "Would you rather buy the product right now from the above store, The nearest store will deliver it at your doorstep within the next few hours. click below\r\nhttps://prestigedemo.latlong.in/redirect/"+link_id+"/123")
			}
		}
	} else {
		message = append(message, fmt.Sprintf("Thank you for contacting %s! \r\nOur speacialist will call you Shortly", brand_name))
	}
	for i := 0; i < len(message); i++ {
		gubshubcall(conn, u.Payload.Payload.Text, bunit_id, brand_name, message[i], destination_number, app_key, app_name, u.Payload.Sender.Name, u.Payload.Type, source_number)
	}
	return nil
}

func welcomenotefinder(processedtext string, rawtext string, destination_number string, sender_name string, payload_type string, conn *pgx.Conn, bunit_id int, brand_name string, app_name string, app_key string, source_number int, store_api_key string) {
	var items Item
	row := conn.QueryRow(context.Background(), "select response_msg from msg_response where request_msg ilike '%' || $1 || '%'", strings.ToLower(processedtext))
	fmt.Println(row)
	err := row.Scan(&items.Response_msg)
	if err != nil {
		log.Println("--------------***Error in method-welcomenotefinder > finding welcome note query-----*******")
		//errormessage(conn, rawtext, bunit_id, brand_name, destination_number, app_key, app_name, sender_name, payload_type, source_number)
		return
	} else {
		message := fmt.Sprintf(items.Response_msg, sender_name, brand_name, brand_name)
		gubshubcall(conn, rawtext, bunit_id, brand_name, message, destination_number, app_key, app_name, sender_name, payload_type, source_number)
		fmt.Println(message)
		return
	}
}

func Latlongnlp(processedtext string, category string, mob_number string) []NlpResp {
	log.Println("---------inside nlp---------", category, processedtext)
	var res []NlpResp
	// mobile_number, err := strconv.Atoi(mob_number)
	// if err != nil {
	// 	fmt.Println("-----------------error in sting to int conversion---------------", err)
	// }
	endpoint := "http://101.53.137.15:10000/entity_tag"
	values := map[string]string{"string_in": processedtext, "category": category, "mob_numb": mob_number}
	jsonValue, _ := json.Marshal(values)
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonValue))
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &res)
	if err != nil {
		log.Println("--------------***Error in method-latlongnlp > ajax call-----*******")
	}
	log.Println("---------response of NLP return nlp---------", res)
	return res
}

func findproduct(wherequery string, rawtext string, destination_number string, sender_name string, payload_type string, conn *pgx.Conn, bunit_id int, brand_name string, app_name string, app_key string, source_number int, store_api_key string) {
	fmt.Println("----------came to product method-----------", wherequery)

	rows, err := conn.Query(context.Background(), "select id,name,product_url,short_url from product_details where "+wherequery)
	fmt.Println(rows)
	var productdetails []Productdetail

	for rows.Next() {
		product := Productdetail{}
		err = rows.Scan(&product.Id, &product.Name, &product.Product_url, &product.Short_url)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("--------------------------------")
		fmt.Println(product)
		productdetails = append(productdetails, product)
	}

	if len(productdetails) == 0 {
		message := "Am unable to find a Product to you, I will aranage a call back. request you to share preferred time"
		gubshubcall(conn, rawtext, bunit_id, brand_name, message, destination_number, app_key, app_name, sender_name, payload_type, source_number)
	} else {
		messagefirst := "Here are some details about the range of products you might be looking for,"
		gubshubcall(conn, rawtext, bunit_id, brand_name, messagefirst, destination_number, app_key, app_name, sender_name, payload_type, source_number)

		for _, v := range productdetails {
			message := v.Product_url
			gubshubcall(conn, rawtext, bunit_id, brand_name, message, destination_number, app_key, app_name, sender_name, payload_type, source_number)
		}
	}

	message := "If you would like to visit the nearest store, kindly share your location"
	gubshubcall(conn, rawtext, bunit_id, brand_name, message, destination_number, app_key, app_name, sender_name, payload_type, source_number)
}

func neareststore(conn *pgx.Conn, text string, brand_name string, bunit_id int, store_api_key string, destination_number string, store_type string) (message []string) {
	var store StoreDetail
	fmt.Println("https://api.latlong.in/v2/brands/" + strconv.Itoa(bunit_id) + "/find.json?query=" + text + "&sure_search=true&access_token=" + store_api_key)
	res, err := http.Get("https://api.latlong.in/v2/brands/" + strconv.Itoa(bunit_id) + "/find.json?query=" + text + "&sure_search=true&access_token=" + store_api_key)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(body, &store)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("--------------------inside nearest store---------------")
	fmt.Println(err)
	if store.Stores != nil && store.Count != 0 {
		var s Shorturl
		point := store.Stores[0].Geofeature.Geoproperty
		t := strings.Replace(point, "(", "", -1)
		t = strings.Replace(t, ")", "", -1)
		t = strings.Replace(t, "POINT", "", -1)
		f := strings.Fields(t)
		res2, err2 := http.Get("https://api.latlong.in/v2/short_urls/create_short_url.json?url_string=https://www.dellretailstores.in/single_store/landing_page/" + strconv.Itoa(store.Stores[0].Id) + "&access_token=d3f1af9bda7ef01f770bfe432f74849a848becbdeeb82e4a015dcad3fb46eaa6&latitude")
		body1, err1 := ioutil.ReadAll(res2.Body)
		if err1 != nil {
			log.Println(err1)
		}
		err2 = json.Unmarshal(body1, &s)
		if err2 != nil {
			fmt.Println(err2)
		}
		message = append(message, store_type+" near - *"+text+"*,\r\n*"+store.Stores[0].Wp_name+"*\r\n"+store.Stores[0].Address+"\r\n*Ph:* "+store.Stores[0].Phone_no)
		if bunit_id == 242 {
			message = append(message, "\r\n*For more Details*\r\nhttps://llmap.in/"+s.Short_url)
		}
		message = append(message, "For detailed direction click below,")
		message = append(message, "{type: location,longitude: "+f[0]+",latitude: "+f[1]+",name: '"+store.Stores[0].Wp_name+"',address: '"+strings.Replace(strings.Replace(store.Stores[0].Address, "\r\n", "", -1), "\\", "\\", -1)+"'}")
		row := conn.QueryRow(context.Background(), "select id,message_received,message_sent,sender_mobile_number,bunit_id from chat_history where sender_mobile_number = '"+destination_number+"' and created_at >= now()-INTERVAL '5 MINUTE' and message_sent in ('https://www.prestigexclusive.in/cookware','https://www.prestigexclusive.in/mixer-grinders','https://www.prestigexclusive.in/SearchResults.aspx?search=MIXER%20GRINDERS')")
		var chat_history chat_history
		fmt.Println("***************************************************")
		fmt.Println(row)
		err := row.Scan(&chat_history.Id, &chat_history.Message_received, &chat_history.Message_sent, &chat_history.Sender_mobile_number, &chat_history.Bunit_id)
		if err != nil {
			fmt.Println(err)
		} else {
			if chat_history.Bunit_id == 271 {
				link_id := "1"
				if chat_history.Message_sent == "https://www.prestigexclusive.in/cookware" {
					link_id = "1"
				} else {
					link_id = "2"
				}
				fmt.Println(chat_history)
				message = append(message, "Would you rather buy the product right now from the above store, The nearest store will deliver it at your doorstep within the next few hours. click below\r\nhttps://prestigedemo.latlong.in/redirect/"+link_id+"/123")
			}
		}
	}
	return message
}

func errormessage(conn *pgx.Conn, msg_received string, bunit_id int, brand_name string, destination_number string, app_key string, app_name string, sender_name string, message_type string, source_number int) {
	//row := conn.QueryRow(context.Background(), "select count(*) from chat_history where sender_mobile_number = $1 and date(created_at) = CURRENT_DATE", destination_number)

	message := "Thank You for contacting us. Our Specialist will call on the query"

	gubshubcall(conn, msg_received, bunit_id, brand_name, message, destination_number, app_key, app_name, sender_name, message_type, source_number)
}

func gubshubcall(conn *pgx.Conn, msg_received string, bunit_id int, brand_name string, message string, destination_number string, app_key string, app_name string, sender_name string, message_type string, source_number int) {
	log.Println("Sending message for app:" + app_name + ", to mnumber:" + destination_number + " , Message is :" + message)
	endpoint := "https://api.gupshup.io/sm/api/v1/msg"
	data := url.Values{}
	data.Set("channel", "whatsapp")
	data.Set("source", strconv.Itoa(source_number))
	data.Set("destination", destination_number)
	data.Set("src.name", app_name)
	data.Set("message", message)
	log.Println("--////// gubshup input-------//////////////")
	log.Println(data)
	log.Println("--------app-----key-------")
	log.Println(app_key)
	client := &http.Client{}
	r, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		log.Fatal(err)
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	r.Header.Add("apikey", app_key)

	res, err := client.Do(r)
	if err != nil {
		log.Println("Error at making gubshub call", err)
	}
	log.Println(res.Status)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Error at reading Response", err)
	}
	response_message := RespMsg{}
	err = json.Unmarshal(body, &response_message)
	log.Println(response_message.MessageId)
	now := time.Now()
	_, err = conn.Exec(context.Background(), "insert into chat_history (created_at,updated_at,message_received,message_sent,message_id,message_status,sender_mobile_number,sender_name,message_type,bunit_id,app_name,brand_name) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)", now, now, msg_received, message, response_message.MessageId, response_message.Status, destination_number, sender_name, message_type, bunit_id, app_name, brand_name)

	log.Println(string(body))
}
