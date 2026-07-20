package folder

import (
	"errors"
	"file_share/internal/entity"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) FolderRescan(c *gin.Context) {
	ctx := c.Request.Context()
	folderId := c.Param("folderId")

	scanJobID := uuid.New().String()

	scanJob := entity.ScanJob{
		Id:               scanJobID,
		FolderId:         folderId,
		Status:           "queued",
		ProcessedVideos:  0,
		ProcessedFolders: 0,
		StartedAt:        time.Now(),
		Error:            "",
	}
	_, err := h.folderService.GetFolder(ctx, folderId)
	if err != nil {
		h.logger.Error(ctx, err)
		if errors.Is(err, entity.ErrorNotFoundFolder) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	job, err := h.scanService.CreateScanJob(ctx, scanJob)
	if err != nil {
		h.logger.Error(ctx, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusAccepted, job)
}
