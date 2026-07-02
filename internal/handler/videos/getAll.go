package videos

import (
	"file_share/internal/entity"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAll(c *gin.Context) {
	rootFolderId := c.Query("rootFolderId")
	parentFolderId := c.Query("parentFolderId")
	query := c.Query("query")
	limit := c.Query("limit")
	offset := c.Query("offset")

	ctx := c.Request.Context()

	iLimit, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{Message: "invalid limit"})
		return
	}
	iOffset, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{Message: "invalid offset"})
		return
	}

	videos, err := h.videoService.GetAll(ctx, query, rootFolderId, parentFolderId, iLimit, iOffset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, videos)
	return
}
