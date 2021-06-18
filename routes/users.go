package routes

import (
	"fmt"
	"latlongwatsapp/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
)

func UsersRegister(c *gin.Context) {
	user := models.User{}
	fmt.Println(user)
	err := c.ShouldBindJSON(&user)
	fmt.Println(err)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, _ := c.Get("db")
	conn := db.(pgx.Conn)
	err = user.Register(&conn)

	if err != nil {
		fmt.Println("error in registeration")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": '1'})
}
