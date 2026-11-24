package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Me(c *gin.Context) {
	ctx := c.Request.Context()

	token, _ := c.Cookie(AccessCookieName)
	if token == "" {
		if header := c.GetHeader("Authorization"); header != "" {
			parts := strings.SplitN(header, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				token = parts[1]
			}
		}
	}

	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token not found"})
		return
	}

	user, err := h.authService.Me(ctx, token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
