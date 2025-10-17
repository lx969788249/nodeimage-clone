package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"nodeimage/api/internal/config"
	"nodeimage/api/internal/handlers"
	"nodeimage/api/internal/middleware"
)

type HTTPServer struct {
	engine *gin.Engine
	server *http.Server
	log    zerolog.Logger
	cfg    *config.AppConfig
}

func NewHTTPServer(cfg *config.AppConfig, log zerolog.Logger, handlerSet handlers.HandlerSet) *HTTPServer {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true

	engine.Use(
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.Recovery(log),
		middleware.CORS(cfg.AllowCORSOrigins),
	)

	handlerSet.Register(engine.Group("/api"))

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      engine,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &HTTPServer{
		engine: engine,
		server: srv,
		log:    log,
		cfg:    cfg,
	}
}

func (s *HTTPServer) Start() error {
	s.log.Info().
		Str("addr", s.server.Addr).
		Msg("http server starting")

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.log.Info().Msg("http server shutting down")
	return s.server.Shutdown(ctx)
}
