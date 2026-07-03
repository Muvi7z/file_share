package videos

import (
	"github.com/gin-gonic/gin"
	"path/filepath"
)

func (h *Handler) VideoStream(c *gin.Context) {
	filename := c.Param("filename")

	filePath := filepath.Join("./hls_files", filename)

	_ = filePath
}
