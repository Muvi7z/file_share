package auth

import (
	"file_share/internal/entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var dataRequest entity.LoginUser
	err := c.ShouldBindJSON(&dataRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.authService.Login(ctx, dataRequest)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.JSON(http.StatusOK, session)
}
