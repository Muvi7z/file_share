package folder

import (
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateRootFolder(c *gin.Context) {
	var dataRequest entity.CreateRootFolderRequest

	ctx := c.Request.Context()

	if err := c.ShouldBindJSON(&dataRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.folderService.CreateFolderRootFolder(ctx, dataRequest.Path)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, folder)

}
