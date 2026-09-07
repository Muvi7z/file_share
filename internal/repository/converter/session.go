package converter

import (
	"file_share/internal/entity"
	"strings"
	"time"
)

func SessionToRedisView(session entity.Session) entity.SessionRedisView {
	return entity.SessionRedisView{
		Token:     session.Token,
		Login:     session.Login,
		Role:      session.Role,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339Nano),
	}
}

func SessionFromRedisView(session entity.SessionRedisView) (entity.Session, error) {
	expiresAt, err := parseRedisTime(session.ExpiresAt)
	if err != nil {
		return entity.Session{}, err
	}

	return entity.Session{
		Token:     session.Token,
		Login:     session.Login,
		Role:      session.Role,
		ExpiresAt: expiresAt,
	}, nil
}

func parseRedisTime(value string) (time.Time, error) {
	expiresAt, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return expiresAt, nil
	}
	legacyValue, _, _ := strings.Cut(value, " m=")
	return time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", legacyValue)
}
