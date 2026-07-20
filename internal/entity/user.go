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

type LoginUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
