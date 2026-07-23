package folder

import (
	"errors"
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetFoldersEntries(c *gin.Context) {
	query := c.Query("query")
	folderId := c.Param("folderId")
	ctx := c.Request.Context()

	folders, err := h.folderService.GetFoldersEntries(ctx, query, "", folderId, nil)
	if err != nil {
		if errors.Is(err, entity.ErrorNotFoundFolder) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		h.logger.Error(ctx, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	videos, err := h.videoService.GetEntries(ctx, query, "", folderId, 0, 0)
	if err != nil {
		h.logger.Error(ctx, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var result []entity.FileBrowserEntry

	result = append(result, folders...)
	result = append(result, videos...)

	c.JSON(http.StatusOK, result)
}
