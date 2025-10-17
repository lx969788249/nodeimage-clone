package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"nodeimage/api/internal/cache"
	"nodeimage/api/internal/config"
	"nodeimage/api/internal/database"
	"nodeimage/api/internal/handlers"
	"nodeimage/api/internal/jobs"
	"nodeimage/api/internal/log"
	"nodeimage/api/internal/server"
	"nodeimage/api/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := log.New(cfg.Environment)

	ctx := context.Background()

	dbPool, err := database.NewPostgresPool(ctx, cfg.Postgres)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect postgres")
	}

	redisClient, err := cache.NewRedisClient(ctx, cfg.Redis)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect redis")
	}

	objectStore, err := storage.NewObjectStore(cfg.Storage)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init object store")
	}
	if err := objectStore.EnsureBuckets(ctx); err != nil {
		logger.Warn().Err(err).Msg("ensure buckets failed")
	}

	handlerSet := handlers.NewHandlerSet(logger, dbPool, redisClient, objectStore, cfg)
	httpServer := server.NewHTTPServer(cfg, logger, handlerSet)

	scheduler := jobs.NewScheduler(redisClient, logger)
	if err := scheduler.Start(); err != nil {
		logger.Error().Err(err).Msg("scheduler start failed")
	}

	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Fatal().Err(err).Msg("http server failed")
		}
	}()

	waitForShutdown(logger, httpServer, scheduler, dbPool, redisClient)
}

func waitForShutdown(logger zerolog.Logger, srv *server.HTTPServer, scheduler *jobs.Scheduler, db *pgxpool.Pool, redisClient *redis.Client) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	logger.Info().Msg("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("graceful shutdown failed")
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error().Err(err).Msg("forced shutdown failed")
		}
	}

	if scheduler != nil {
		cancel := scheduler.Stop()
		cancel()
	}

	db.Close()
	if err := redisClient.Close(); err != nil {
		logger.Error().Err(err).Msg("redis close error")
	}

	logger.Info().Msg("server exited cleanly")
}
