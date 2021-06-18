package main

import (
	"context"
	"io"
	"latlongwatsapp/config"
	"latlongwatsapp/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

func main() {
	conn, err := connectDB()
	if err != nil {
		return
	}
	//gin.SetMode(gin.ReleaseMode)
	//router := gin.New()
	root_path, err := os.Getwd()
	if err != nil {
		log.Println("error at root path")
	} else {
		log.Println(root_path)
	}
	logFile, err := os.Create(root_path + "/log/production.log")
	if err != nil {
		log.Println(err)
	}
	if err != nil {
		log.Println(err)
	}
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(gin.DefaultWriter)
	router := gin.Default()

	router.Use(dbMiddleware(*conn))

	usersGroup := router.Group("users")

	{
		usersGroup.POST("/register", routes.UsersRegister)
	}

	inboundmsgsGroup := router.Group("latlongwatsappapi")
	{
		inboundmsgsGroup.POST("inboundmsg", routes.IncomingdMsg)
	}
	server := config.Severconfig{}
	port := server.Assignserver()
	router.RunTLS(":"+port, "/etc/nginx/latlong_trust.crt", "/etc/nginx/latlong_trust.key")
	router.Run(":" + port)
}

func connectDB() (c *pgx.Conn, err error) {
	db := config.Databaseconfig{}
	username, password, host, port := db.Assigndb()
	conn, err := pgx.Connect(context.Background(), "postgresql://"+username+":"+password+"@"+host+":"+port+"/latlongwatsapp")
	if err != nil {
		log.Println("error connecting to db below is the error")
		log.Println(err.Error())
	}
	_ = conn.Ping(context.Background())
	return conn, err
}

func dbMiddleware(conn pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", conn)
		//for continues procesing it
		c.Next()
	}
}
