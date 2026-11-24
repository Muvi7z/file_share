package auth

import (
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var dataRequest entity.LoginUser
	err := c.ShouldBindJSON(&dataRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.authService.Login(ctx, dataRequest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "error to login"})
		return
	}

	h.setAuthCookies(c, res.Token.AccessToken, res.Token.RefreshToken, h.accessTokenTTL, h.refreshTokenTTL)
	c.JSON(http.StatusOK, res.User)
}
