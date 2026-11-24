package ws

import (
	"context"
	"errors"
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) TvNowPlaying(c *gin.Context) {
	var request entity.CmdPlayNowOnTV

	ctx := c.Request.Context()

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.tvService.TvNowPlaying(ctx, request)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			c.JSON(http.StatusNoContent, gin.H{
				"status": "play now off",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
