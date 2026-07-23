package middleware

import (
	"file_share/internal/deps"
	"file_share/internal/service/roles"
	"fmt"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	rolesProvider roles.RolesProvider
	logger        deps.Logger
	excludedPaths map[string][]string
}

func NewMiddleware(rolesProvider roles.RolesProvider, logger deps.Logger) *Middleware {
	excludedPaths := map[string][]string{
		"GET:/api/videos/":                {"admin"},
		"GET:/api/videos/:videoId/stream": {"admin"},
		"GET:/api/videos/:videoId/poster": {"admin"},
		"GET:/api/folders/":               {"admin"},
		"POST:/api/folders/:folderId":     {"admin"},
		"/api/folders/root/entries":       {"admin"},
		"/api/folders/:folderId":          {"admin"},
		"/api/folders/:folderId/entries":  {"admin"},
		"/api/folders/:folderId/rescan":   {"admin"},
	}

	return &Middleware{
		rolesProvider: rolesProvider,
		logger:        logger,
		excludedPaths: excludedPaths,
	}
}

func (m *Middleware) Apply(c *gin.Context) {
	path := fmt.Sprintf("%s:%s", c.Request.Method, c.FullPath())
	if m.excludedPaths[path] != nil {

	}
}
