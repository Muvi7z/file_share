package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	AccessCookieName  = "access_token"
	RefreshCookieName = "refresh_token"
)

func (h *Handler) setAuthCookies(c *gin.Context, accessToken string, refreshToken string, _, refreshTokenTTL time.Duration) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AccessCookieName,
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(refreshTokenTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cookieSecure,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    refreshToken,
		Path:     "/api/auth",
		MaxAge:   int(refreshTokenTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cookieSecure,
	})
}

func (h *Handler) clearAuthCookies(c *gin.Context) {
	// Path обязан совпадать с тем, что был при установке, иначе браузер не удалит куку
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AccessCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}
