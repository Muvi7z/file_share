package folder

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DeleteFolder(c *gin.Context) {
	folderId := c.Param("folderId")

	ctx := c.Request.Context()

	err := h.folderService.DeleteFolder(ctx, folderId)
	if err != nil {
		h.logger.Error(ctx, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, true)
}
