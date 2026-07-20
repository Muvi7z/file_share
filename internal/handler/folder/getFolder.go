package folder

import (
	"errors"
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetFolder(c *gin.Context) {
	folderId := c.Param("folderId")

	ctx := c.Request.Context()

	folder, err := h.folderService.GetFolder(ctx, folderId)
	if err != nil {
		h.logger.Error(ctx, err)
		if errors.Is(err, entity.ErrorNotFoundFolder) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, folder)
}
