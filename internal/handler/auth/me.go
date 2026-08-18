package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (h *Handler) Me(c *gin.Context) {
	ctx := c.Request.Context()
	header := c.GetHeader("Authorization")
	parts := strings.Split(header, " ")

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "error not found bearer"})
		return
	}

	token := parts[1]

	user, err := h.authService.Me(ctx, token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
