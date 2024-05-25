package model

import "time"

type User struct {
	ID          string
	Username    string
	Password    string
	FirstName   string
	LastName    string
	Email       string
	PhoneNumber string
	CreatedAt   time.Time
	LastUpdated time.Time
	Avatar      *Avatar
}

type Avatar string

type CreateUserParams struct {
	ID           string
	Username     string
	HashPassword string
	FirstName    string
	LastName     string
	Email        string
	PhoneNumber  string
	Avatar       *Avatar
}

type UpdateUserParams struct {
	Password    string
	FirstName   string
	LastName    string
	Email       string
	PhoneNumber string
	Avatar      *Avatar
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Email       string  `json:"email"`
	PhoneNumber string  `json:"phone_number"`
	Avatar      *Avatar `json:"avatar"`
}

type UpdateUserRequest struct {
	Password    string  `json:"password"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Email       string  `json:"email"`
	PhoneNumber string  `json:"phone_number"`
	Avatar      *Avatar `json:"avatar"`
}

type GetUserResponse struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	Avatar      *Avatar   `json:"avatar"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
}
