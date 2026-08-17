package converter

import (
	"file_share/internal/entity"
)

func SessionToRedisView(session entity.Session) entity.SessionRedisView {
	return entity.SessionRedisView{
		Token:     session.Token,
		Login:     session.Login,
		Role:      session.Role,
		ExpiresAt: session.ExpiresAt,
	}
}

func SessionFromRedisView(session entity.SessionRedisView) entity.Session {
	return entity.Session{
		Token:     session.Token,
		Login:     session.Login,
		Role:      session.Role,
		ExpiresAt: session.ExpiresAt,
	}
}
