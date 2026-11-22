package ws

import (
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) OpenOnTV(c *gin.Context) {
	var request entity.CmdOpenOnTV
	ctx := c.Request.Context()

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.tvService.OpenOnTV(ctx, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
