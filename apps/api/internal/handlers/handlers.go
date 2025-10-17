package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"nodeimage/api/internal/config"
	"nodeimage/api/internal/middleware"
	"nodeimage/api/internal/models"
	"nodeimage/api/internal/repository"
	"nodeimage/api/internal/service"
	"nodeimage/api/internal/storage"
)

type HandlerSet struct {
	log         zerolog.Logger
	cfg         *config.AppConfig
	authService *service.AuthService
	uploadService *service.UploadService
	db          *pgxpool.Pool
	cache       *redis.Client
	store       *storage.ObjectStore
	users       *repository.UserRepository
	sessions    *repository.SessionRepository
	images      *repository.ImageRepository
}

func NewHandlerSet(log zerolog.Logger, db *pgxpool.Pool, cache *redis.Client, store *storage.ObjectStore, cfg *config.AppConfig) HandlerSet {
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	imageRepo := repository.NewImageRepository(db)
	auth := service.NewAuthService(userRepo, sessionRepo, cache, cfg, log)
	upload := service.NewUploadService(imageRepo, store, cache, cfg, log)

	return HandlerSet{
		log:         log,
		cfg:         cfg,
		authService: auth,
		uploadService: upload,
		db:          db,
		cache:       cache,
		store:       store,
		users:       userRepo,
		sessions:    sessionRepo,
		images:      imageRepo,
	}
}

func (h HandlerSet) Register(router *gin.RouterGroup) {
	router.GET("/healthz", h.Health)

	v1 := router.Group("/v1")
	{
		auth := v1.Group("/auth")
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)

		protected := v1.Group("/auth")
		protected.Use(
			middleware.Auth(h.cfg, h.users, h.sessions),
			middleware.Signature(h.cfg, h.cache),
		)
		protected.GET("/me", h.Me)
		protected.GET("/sessions", h.ListSessions)
		protected.DELETE("/sessions/:deviceId", h.RevokeSession)
	}

	media := v1.Group("/media")
	media.Use(
		middleware.Auth(h.cfg, h.users, h.sessions),
		middleware.Signature(h.cfg, h.cache),
	)
	media.POST("/upload", h.UploadMedia)

	admin := v1.Group("/admin")
	admin.Use(
		middleware.Auth(h.cfg, h.users, h.sessions),
		middleware.Signature(h.cfg, h.cache),
		middleware.RequireRoles(models.UserRoleAdmin, models.UserRoleSuperAdmin),
	)
	admin.GET("/images", h.AdminListImages)
}
