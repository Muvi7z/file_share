package server

import (
	middleware "file_share/internal/handler/middlewares/auth"
	videosHandler "file_share/internal/handler/videos"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router         *gin.Engine
	videosHandler  *videosHandler.Handler
	authMiddleware *middleware.Middleware
}

func NewServer(router *gin.Engine, videosHandler *videosHandler.Handler) *Server {

	return &Server{
		videosHandler: videosHandler,
		router:        router,
	}
}

func LiberalCORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")

	if c.Request.Method == "OPTIONS" {
		if len(c.Request.Header["Access-Control-Request-Headers"]) > 0 {
			c.Header("Access-Control-Allow-Headers",
				c.Request.Header["Access-Control-Request-Headers"][0])
		}
		c.AbortWithStatus(http.StatusOK)
	}
}

func (s *Server) Register() {
	s.router.Use(LiberalCORS)
	s.router.Use(s.authMiddleware.Apply)
	api := s.router.Group("/api")
	{
		api.GET("/health")

		auth := api.Group("/auth")
		{
			auth.POST("/login")
			auth.POST("/me")
		}

		videos := api.Group("/videos")
		{
			videos.GET("/", s.videosHandler.GetAll)
			videos.GET("/:videoId")
			videos.GET("/:videoId/stream", s.videosHandler.VideoStream)
			videos.GET("/:videoId/poster")

		}

		files := api.Group("/files")
		{
			files.GET("/")
		}

		folders := api.Group("/folders")
		{
			folders.GET("/")
			folders.POST("/")
			folders.GET("/:folderId")
			folders.GET("/root/entries")
			folders.PATCH("/:folderId")
			folders.DELETE("/:folderId")
			folders.GET("/:folderId/entries")
			folders.GET("/:folderId/rescan")
		}

		api.GET("/scan-jobs/:jobId")
	}
}
