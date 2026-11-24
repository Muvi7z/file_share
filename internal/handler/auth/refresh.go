package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Refresh(c *gin.Context) {
	ctx := c.Request.Context()

	refreshToken, err := c.Cookie(RefreshCookieName)
	if err != nil || refreshToken == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid refresh token",
		})
		return
	}

	sessionId, secret, ok := strings.Cut(refreshToken, ".")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid refresh token",
		})
		return
	}

	res, err := h.authService.RefreshTokens(ctx, sessionId, secret)
	if err != nil {
		h.clearAuthCookies(c)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "invalid refresh token",
		})
		return
	}
	h.setAuthCookies(c, res.AccessToken, res.RefreshToken, h.accessTokenTTL, h.refreshTokenTTL)
	c.Status(http.StatusNoContent)
}
