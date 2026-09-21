package file

import (
	"file_share/internal/entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) DeleteFile(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()

	err := h.fileService.DeleteFile(ctx, id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Code:    "500",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}
