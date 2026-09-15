package entity

import "time"

const (
	StatusActive   = "active"
	StatusPending  = "pending"
	StatusBlocked  = "blocked"
	StatusRejected = "rejected"
)

type User struct {
	Id           string    `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"passwordHash"`
	Status       string    `json:"status"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type MeUser struct {
	Token     string    `json:"token"`
	Login     string    `json:"login"`
	Role      Role      `json:"role"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type LoginUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Status string `json:"status"`
}
