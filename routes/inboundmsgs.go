package routes

import (
	"fmt"
	"latlongwatsapp/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

func IncomingdMsg(c *gin.Context) {
	//app := c.GetString("app")
	inboundmsg := models.InboundBody{}
	err := c.ShouldBindJSON(&inboundmsg)
	db, _ := c.Get("db")
	fmt.Println(err)
	conn := db.(pgx.Conn)
	err = inboundmsg.Process(&conn)
	if err != nil {
		//c.JSON(http.StatusOK, gin.H{"status": err})
	} else {
		//c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
