package report

import (
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	var dataRequest entity.CreateReportVideo

	if err := c.ShouldBindJSON(&dataRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.reportVideoService.CreateReportVideo(ctx, dataRequest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, entity.ErrorResponse{
			Message: "error create report",
			Code:    "500",
		})
		return
	}

	c.JSON(http.StatusOK, folder)
}
