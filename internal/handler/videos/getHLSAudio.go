package videos

import (
	"file_share/internal/entity"
	"file_share/pkg/utils/video"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetHLSAudio(c *gin.Context) {
	ctx := c.Request.Context()
	audioId := c.Param("audioId")
	fileName := c.Param("fileName")

	stream, err := h.playbackService.StreamHLS(ctx, fileName, audioId, entity.FileTypeAudio)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Message: err.Error(),
			Code:    "500",
		})
		return
	}
	defer stream.Reader.Close()

	contentType := video.ContentType(stream.FileName)

	c.Header("Content-Type", contentType)

	http.ServeContent(
		c.Writer,
		c.Request,
		stream.FileName,
		stream.ModTime,
		stream.Reader,
	)
}
