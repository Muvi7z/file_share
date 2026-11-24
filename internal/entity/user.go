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
	Login string `json:"login"`
	Role  Role   `json:"role"`
}

type LoginUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type LoginResponse struct {
	Token Token  `json:"token"`
	User  MeUser `json:"user"`
}

type RegisterUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Status string `json:"status"`
}
