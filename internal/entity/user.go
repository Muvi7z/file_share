package entity

import "time"

type User struct {
	Id           string    `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"passwordHash"`
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
