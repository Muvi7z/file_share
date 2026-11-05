package videos

import (
	"errors"
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetVideoAudioTracks(c *gin.Context) {
	ctx := c.Request.Context()
	videoId := c.Param("videoId")

	audio, err := h.audioTrackService.GetAll(ctx, videoId, 0, 0)
	if err != nil {
		h.logger.Error(ctx, err)
		if errors.Is(err, entity.ErrorNotFoundAudioTrack) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{Message: err.Error(), Code: "500"})
		return
	}

	c.JSON(http.StatusOK, audio)
}
