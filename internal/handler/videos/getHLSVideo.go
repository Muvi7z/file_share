package videos

import (
	"file_share/internal/entity"
	"file_share/pkg/utils/video"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetHLSVideo(c *gin.Context) {
	ctx := c.Request.Context()
	videoId := c.Param("videoId")
	fileName := c.Param("fileName")
	typeFile := entity.FileTypeVideo
	if strings.EqualFold(fileName, "master.m3u8") {
		typeFile = entity.FileTypePlaylistHLS
	}

	stream, err := h.playbackService.StreamHLS(ctx, fileName, videoId, typeFile)
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
