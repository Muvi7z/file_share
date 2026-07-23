package folder

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAll(c *gin.Context) {
	rootFolderId := c.Query("rootFolderId")
	parentFolderId := c.Query("parentFolderId")
	query := c.Query("query")

	ctx := c.Request.Context()

	folders, err := h.folderService.GetFolders(ctx, rootFolderId, parentFolderId, query, nil)
	if err != nil {
		h.logger.Error(ctx, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, folders)
}
