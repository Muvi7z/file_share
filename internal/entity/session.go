package entity

import "time"

type Session struct {
	Token     string    `json:"token"`
	Login     string    `json:"login"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expiresAt"`
}
