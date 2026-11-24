package middleware

import (
	"context"
	"errors"
	"file_share/internal/deps"
	"file_share/internal/entity"
	"file_share/internal/service/roles"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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

		token, _ := c.Cookie("access_token")

		if token == "" {
			if header := c.GetHeader("Authorization"); header != "" {
				parts := strings.SplitN(header, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					token = parts[1]
				}
			}
		}

		if token == "" {
			if _, ok := roleList[entity.RoleViewer]; ok {
				ctx = context.WithValue(ctx, "role", entity.RoleViewer)
				c.Request = c.Request.WithContext(ctx)
				c.Next()
				return
			}
			// 401 (а не 403), чтобы фронт понимал: нужно авторизоваться / обновить токен
			c.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrorResponse{
				Message: "Unauthorized",
				Code:    "401",
			})
			return
		}

		user, err := m.authService.Me(ctx, token)
		if err != nil {
			if errors.Is(err, entity.ErrExpiredToken) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, entity.ErrorResponse{
					Message: "Unauthorized",
					Code:    "401",
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, entity.ErrorResponse{
				Message: "Forbidden",
				Code:    "403",
			})
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
