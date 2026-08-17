package middleware

import (
	"file_share/internal/deps"
	"file_share/internal/entity"
	"file_share/internal/service/auth"
	"file_share/internal/service/roles"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	rolesProvider roles.RolesProvider
	logger        deps.Logger
	excludedPaths map[string][]string
	authService   *auth.Service
}

func NewMiddleware(logger deps.Logger) *Middleware {
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
		logger:        logger,
		excludedPaths: excludedPaths,
	}
}

func (m *Middleware) Apply(allowed ...entity.Role) gin.HandlerFunc {
	roles := make(map[entity.Role]bool)

	for _, role := range allowed {
		roles[role] = true
	}

	return func(c *gin.Context) {
		//header := c.GetHeader("Authorization")
		//ctx := c.Request.Context()
		//parts := strings.Split(header, " ")
		//
		//if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "error not found bearer"})
		//	return
		//}
		//
		//token := parts[1]

	}
}
