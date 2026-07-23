package videos

import (
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Stream(c *gin.Context) {
	ctx := c.Request.Context()
	videoId := c.Param("videoId")

	stream, err := h.videoService.Stream(ctx, videoId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Message: err.Error(),
			Code:    "500",
		})
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Disposition", `inline; filename="`+stream.FileName+`"`)

	http.ServeContent(
		c.Writer,
		c.Request,
		stream.FileName,
		stream.ModTime,
		stream.Reader,
	)
}
