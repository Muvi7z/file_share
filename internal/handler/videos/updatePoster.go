package videos

import (
	"errors"
	"file_share/internal/entity"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const maxPosterSize = 60 << 20 // 60 MiB

func (h *Handler) UpdatePoster(c *gin.Context) {
	videoID := c.Param("videoId")
	ctx := c.Request.Context()

	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxPosterSize+(1<<20),
	)

	header, err := c.FormFile("poster")
	if err != nil {
		var sizeError *http.MaxBytesError
		if errors.As(err, &sizeError) {
			h.logger.Error(ctx, err, errors.New("постер слишком больших размеров"))
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge,
				gin.H{"error": "poster is too large"},
			)
			return
		}
		h.logger.Error(ctx, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "invalid poster"})
		return
	}

	if header.Size <= 0 {
		h.logger.Error(ctx, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "invalid poster"})
		return
	}

	if header.Size > maxPosterSize {
		h.logger.Error(ctx, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "Размер постера не должен превышать 10 МБ"})
		return
	}

	file, err := header.Open()
	if err != nil {
		h.logger.Error(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "invalid poster"},
		)
		return
	}
	defer file.Close()

	//config, _, err := image.DecodeConfig(file)
	//if err != nil {
	//	h.logger.Error(ctx, err)
	//	c.AbortWithStatusJSON(http.StatusBadRequest,
	//		gin.H{"error": "poster isn't a valid image"},
	//	)
	//}
	//
	//if config.Height <= 0 || config.Width <= 0 || int64(config.Width)*int64(config.Height) > 40_000_000 {
	//	h.logger.Error(ctx, err)
	//	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid size poster"})
	//}

	posterFile := entity.PosterFile{
		FileName:    header.Filename,
		ContentType: "",
		ModTime:     time.Time{},
		Reader:      file,
		Path:        "",
		VideoId:     videoID,
	}
	url, err := h.videoService.UpdatePoster(ctx, posterFile)
	if err != nil {
		h.logger.Error(ctx, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error to update poster"})
	}

	c.JSON(http.StatusOK, url)
}
