package folder

import (
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateFolder(c *gin.Context) {
	ctx := c.Request.Context()
	folderId := c.Param("folderId")
	var dataRequest entity.UpdateFolderRequest

	if err := c.ShouldBindJSON(&dataRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	errorRes := entity.ErrorResponse{}

	folders, err := h.folderService.UpdateFolder(ctx, dataRequest, folderId)
	if err != nil {
		h.logger.Error(ctx, err)
		errorRes.Code = "500"
		errorRes.Message = err.Error()
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorRes)
		return
	}

	c.JSON(http.StatusOK, folders)
}
