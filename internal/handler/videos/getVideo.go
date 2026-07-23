package videos

import (
	"errors"
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetVideo(c *gin.Context) {
	ctx := c.Request.Context()
	videoId := c.Param("videoId")

	video, err := h.videoService.GetVideoById(ctx, videoId)
	if err != nil {
		h.logger.Error(ctx, err)
		if errors.Is(err, entity.ErrorNotFoundVideo) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{Message: err.Error(), Code: "500"})
		return
	}

	c.JSON(http.StatusOK, video)
}
