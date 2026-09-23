package report

import (
	"file_share/internal/entity"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	var dataRequest entity.ReportVideo
	if err := c.ShouldBindJSON(&dataRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	errorRes := entity.ErrorResponse{}

	report, err := h.reportVideoService.UpdateReportVideo(ctx, dataRequest)
	if err != nil {
		errorRes.Message = err.Error()
		errorRes.Code = "500"
		c.JSON(http.StatusInternalServerError, errorRes)
		return
	}

	c.JSON(http.StatusOK, report)
}
