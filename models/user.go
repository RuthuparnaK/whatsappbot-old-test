package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

type User struct {
	ID           int       `json:"id"`
	CreatedAt    time.Time `json:"_"`
	UpdatedAt    time.Time `json:"_"`
	Name         string    `json:"name"`
	MobileNumber int       `json:"mobile_number"`
	LastPing     time.Time `json:"last_ping"`
	IsActive     bool      `json:"is_active"`
}

func (u *User) Register(conn *pgx.Conn) error {
	// if len(u.MobileNumber) < 10 {
	// 	return fmt.Errorf("Mobile Number is less than 10 Digit or invalid mobile number")
	// }
	row := conn.QueryRow(context.Background(), "SELECT mobile_number from user_account WHERE mobile_number = $1", u.MobileNumber)
	userLookup := User{}
	// fmt.Println(userLookup)
	err := row.Scan(&userLookup.MobileNumber)
	fmt.Println(userLookup)
	fmt.Println(err)
	// now := time.Now()
	// _, err = conn.Exec(context.Background(), "insert into user_account (created_at,updated_at,name,last_ping,mobile_number) values ($1,$2,$3,$4,$5)", now, now, u.Name, now, u.MobileNumber)

	return nil
}
