package videos

import (
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetPoster(c *gin.Context) {
	videoID := c.Param("videoId")
	ctx := c.Request.Context()
	poster, err := h.videoService.GetPoster(ctx, videoID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, entity.ErrorResponse{
			Message: err.Error(),
			Code:    "404",
		})
		return
	}
	defer poster.Reader.Close()

	c.Header("Content-Type", poster.ContentType)
	c.Header("Cache-Control", "public, max-age=86400")

	http.ServeContent(
		c.Writer,
		c.Request,
		poster.FileName,
		poster.ModTime,
		poster.Reader,
	)
}
