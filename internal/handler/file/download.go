package file

import (
	"file_share/internal/entity"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Download(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	errorRes := entity.ErrorResponse{}
	file, err := h.fileService.GetFileById(ctx, id)
	if err != nil {
		h.logger.Error(c.Request.Context(), err)
		errorRes.Code = "500"
		errorRes.Message = err.Error()
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorRes)
		return
	}

	openFile, err := os.Open(file.Path)
	if err != nil {
		h.logger.Error(c.Request.Context(), err)
		errorRes.Code = "500"
		errorRes.Message = err.Error()
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorRes)
		return
	}
	defer openFile.Close()

	fileInfo, err := openFile.Stat()
	if err != nil {

		errorRes.Code = "500"
		errorRes.Message = "failed to read file information"
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorRes)
		return
	}

	if fileInfo.IsDir() {
		c.AbortWithStatusJSON(http.StatusBadRequest, entity.ErrorResponse{
			Message: "requested path is a directory",
			Code:    "400",
		})
		return
	}

	extension := strings.ToLower(strings.TrimSpace(file.Extension))
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	fileName := filepath.Base(strings.TrimSpace(file.Name))
	if fileName == "." || fileName == "" {
		fileName = filepath.Base(file.Path)
	}
	if extension != "" && !strings.EqualFold(filepath.Ext(fileName), extension) {
		fileName += extension
	}

	contentType := mime.TypeByExtension(extension)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	mimeImg, ok := entity.AllowedImageExts[extension]
	if ok {
		contentType = mimeImg
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", mime.FormatMediaType(
		"attachment",
		map[string]string{"filename": fileName},
	))
	c.Header("Cache-Control", "private, max-age=3600")
	c.Header("Accept-Ranges", "bytes")
	c.Header("X-Content-Type-Options", "nosniff")

	http.ServeContent(
		c.Writer,
		c.Request,
		fileName,
		fileInfo.ModTime(),
		openFile,
	)
}
