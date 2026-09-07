package middleware

import (
	"context"
	"file_share/internal/deps"
	"file_share/internal/entity"
	"file_share/internal/service/roles"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

type Middleware struct {
	rolesProvider roles.RolesProvider
	logger        deps.Logger
	excludedPaths map[string][]string
	authService   authService
}

type authService interface {
	Me(ctx context.Context, token string) (entity.MeUser, error)
}

func NewMiddleware(logger deps.Logger, authService authService) *Middleware {
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
		authService:   authService,
	}
}

func (m *Middleware) Apply(allowed ...entity.Role) gin.HandlerFunc {
	roleList := make(map[entity.Role]bool)

	for _, role := range allowed {
		roleList[role] = true
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		header := c.GetHeader("Authorization")

		if len(header) == 0 {
			_, ok := roleList[entity.RoleViewer]
			if !ok {
				c.AbortWithStatusJSON(403, gin.H{
					"error": "forbidden",
				})
				return
			}
		}

		parts := strings.Split(header, " ")

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "error not found bearer"})
			return
		}

		token := parts[1]

		user, err := m.authService.Me(ctx, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, entity.ErrorResponse{
				Message: "Forbidden",
				Code:    "403",
			})
			return
		}

		if user.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has expired"})
			return
		}

		_, ok := roleList[user.Role]
		if !ok {
			c.AbortWithStatusJSON(403, gin.H{
				"error": "forbidden",
			})
			return
		}

		c.Next()
	}
}
