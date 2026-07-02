package entity

import "time"

type User struct {
	Id           string    `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"passwordHash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
