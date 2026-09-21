package videos

import (
	"file_share/internal/entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) DeleteVideo(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	err := h.videoService.DeleteVideo(ctx, id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Message: "delete video failed",
			Code:    "500",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}
