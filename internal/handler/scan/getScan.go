package scan

import (
	"errors"
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetScanJob(c *gin.Context) {
	ctx := c.Request.Context()
	scanJobId := c.Param("jobId")

	job, err := h.scanService.GetScanJob(ctx, scanJobId)
	if err != nil {
		h.logger.Error(ctx, err)
		if errors.Is(err, entity.ErrorNotFoundScan) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, job)
}
