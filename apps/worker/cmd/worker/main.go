package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"nodeimage/worker/internal/config"
	"nodeimage/worker/internal/log"
	"nodeimage/worker/internal/queue"
	"nodeimage/worker/internal/tasks"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := log.New(cfg.Logging.Level)

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Fatal().Err(err).Msg("redis connection failed")
	}
	defer client.Close()

	processor := tasks.NewProcessor(logger)
	consumer := queue.NewConsumer(
		client,
		cfg.Redis.Stream,
		cfg.Redis.Group,
		cfg.Redis.Consumer,
		cfg.Queues.ClaimInterval,
		logger,
		processor,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := consumer.Start(ctx); err != nil && err != context.Canceled {
			logger.Fatal().Err(err).Msg("consumer stopped unexpectedly")
		}
	}()

	<-ctx.Done()
	logger.Info().Msg("shutdown signal received")
	time.Sleep(500 * time.Millisecond)
}
