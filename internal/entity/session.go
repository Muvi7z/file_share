package entity

import "time"

type Session struct {
	Id        string    `json:"id"`
	Token     string    `json:"token"`
	Login     string    `json:"login"`
	Role      Role      `json:"role"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type SessionRedisView struct {
	Id        string `redis:"id"`
	Token     string `redis:"token"`
	Login     string `redis:"login"`
	Role      Role   `redis:"role"`
	ExpiresAt string `redis:"expires_at"`
}
