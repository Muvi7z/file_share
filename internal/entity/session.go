package entity

import "time"

type Session struct {
	Token     string    `json:"token"`
	Login     string    `json:"login"`
	Role      Role      `json:"role"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type SessionRedisView struct {
	Token     string    `redis:"token"`
	Login     string    `redis:"login"`
	Role      Role      `redis:"role"`
	ExpiresAt time.Time `redis:"expires_at"`
}
