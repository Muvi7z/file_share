package folder

import (
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetRootFolderEntries(c *gin.Context) {
	ctx := c.Request.Context()
	query := c.Query("query")

	errorRes := entity.ErrorResponse{}
	var isRoot bool

	isRoot = true
	folders, err := h.folderService.GetFoldersEntries(ctx, query, "", "", &isRoot)
	if err != nil {
		h.logger.Error(ctx, err)
		errorRes.Code = "500"
		errorRes.Message = err.Error()
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorRes)
		return
	}

	c.JSON(http.StatusOK, folders)
}
