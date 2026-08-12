package server

import (
	"context"
	"errors"
	"file_share/internal/deps"
	"file_share/internal/handler/folder"
	middleware "file_share/internal/handler/middlewares/auth"
	"file_share/internal/handler/scan"
	videosHandler "file_share/internal/handler/videos"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	videosHandler  *videosHandler.Handler
	folderHandler  *folder.Handler
	scanHandler    *scan.Handler
	logger         deps.Logger
	addr           string
	server         *http.Server
	authMiddleware *middleware.Middleware
}

func NewServer(videosHandler *videosHandler.Handler, folderHandler *folder.Handler, scanHandler *scan.Handler, logger deps.Logger, addr string) *Server {
	s := &Server{
		videosHandler: videosHandler,
		folderHandler: folderHandler,
		scanHandler:   scanHandler,
		logger:        logger,
		addr:          addr,
	}

	router := s.Register(gin.Default())

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: router,
	}

	return s
}

func (s *Server) Run(ctx context.Context) {
	s.logger.Info(ctx, "server starting", "address", s.addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error(ctx, err)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
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

func (s *Server) Register(router *gin.Engine) *gin.Engine {
	router.Use(LiberalCORS)
	//router.Use(s.authMiddleware.Apply)

	api := router.Group("/api")
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
			videos.GET("/:videoId", s.videosHandler.GetVideo)
			videos.GET("/:videoId/stream", s.videosHandler.Stream)
			videos.HEAD("/:videoId/stream", s.videosHandler.Stream)
			videos.GET("/:videoId/poster", s.videosHandler.GetPoster)

		}

		files := api.Group("/files")
		{
			files.GET("/")
		}

		folders := api.Group("/folders")
		{
			folders.GET("/", s.folderHandler.GetAll)
			folders.POST("/", s.folderHandler.CreateRootFolder)
			folders.GET("/:folderId", s.folderHandler.GetFolder)
			folders.GET("/root/entries", s.folderHandler.GetRootFolderEntries)
			folders.PATCH("/:folderId", s.folderHandler.UpdateFolder)
			folders.DELETE("/:folderId", s.folderHandler.DeleteFolder)
			folders.GET("/:folderId/entries", s.folderHandler.GetFoldersEntries)
			folders.GET("/:folderId/rescan", s.folderHandler.FolderRescan)
		}

		api.GET("/scan-jobs/:jobId", s.scanHandler.GetScanJob)
	}

	return router
}
