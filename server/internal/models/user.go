package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
