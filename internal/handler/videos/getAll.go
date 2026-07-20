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

	iLimit := 0
	iOffset := 0
	var err error

	ctx := c.Request.Context()
	if limit != "" {
		iLimit, err = strconv.Atoi(limit)
		if err != nil {
			h.logger.Error(ctx, err)
			c.JSON(http.StatusBadRequest, entity.ErrorResponse{Message: "invalid limit", Code: "400"})
			return
		}
	}

	if offset != "" {
		iOffset, err = strconv.Atoi(offset)
		if err != nil {
			c.JSON(http.StatusBadRequest, entity.ErrorResponse{Message: "invalid offset", Code: "400"})
			return
		}
	}

	videos, err := h.videoService.GetAll(ctx, query, rootFolderId, parentFolderId, iLimit, iOffset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorResponse{Message: err.Error(), Code: "500"})
		return
	}

	c.JSON(http.StatusOK, videos)
	return
}
