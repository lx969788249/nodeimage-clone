package queue

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type MessageHandler interface {
	Handle(ctx context.Context, msg redis.XMessage) error
}

type Consumer struct {
	client  *redis.Client
	stream  string
	group   string
	consumer string
	claimInterval time.Duration
	logger  zerolog.Logger
	handler MessageHandler
}

func NewConsumer(client *redis.Client, stream, group, consumer string, claimInterval time.Duration, logger zerolog.Logger, handler MessageHandler) *Consumer {
	return &Consumer{
		client:  client,
		stream:  stream,
		group:   group,
		consumer: consumer,
		claimInterval: claimInterval,
		logger:  logger,
		handler: handler,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	ticker := time.NewTicker(c.claimInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := c.read(ctx); err != nil {
				c.logger.Error().Err(err).Msg("stream read error")
				time.Sleep(2 * time.Second)
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_ = c.claimStalled(ctx)
		default:
		}
	}
}

func (c *Consumer) read(ctx context.Context) error {
	result, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    c.group,
		Consumer: c.consumer,
		Streams:  []string{c.stream, ">"},
		Count:    10,
		Block:    5 * time.Second,
	}).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	for _, stream := range result {
		for _, msg := range stream.Messages {
			if err := c.handleMessage(ctx, msg); err != nil {
				c.logger.Error().
					Err(err).
					Str("message_id", msg.ID).
					Msg("handle message failed")
				continue
			}
			if err := c.client.XAck(ctx, c.stream, c.group, msg.ID).Err(); err != nil {
				c.logger.Error().Err(err).Str("message_id", msg.ID).Msg("ack failed")
			}
		}
	}
	return nil
}

func (c *Consumer) handleMessage(ctx context.Context, msg redis.XMessage) error {
	return c.handler.Handle(ctx, msg)
}

func (c *Consumer) claimStalled(ctx context.Context) error {
	pending, err := c.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream:   c.stream,
		Group:    c.group,
		Start:    "-",
		End:      "+",
		Count:    10,
		Consumer: "",
	}).Result()
	if err != nil {
		return err
	}

	for _, entry := range pending {
		if entry.Idle < c.claimInterval {
			continue
		}
		msgs, err := c.client.XClaim(ctx, &redis.XClaimArgs{
			Stream:   c.stream,
			Group:    c.group,
			Consumer: c.consumer,
			MinIdle:  c.claimInterval,
			Messages: []string{entry.ID},
		}).Result()
		if err != nil {
			c.logger.Error().Err(err).Msg("claim error")
			continue
		}
		for _, msg := range msgs {
			if err := c.handleMessage(ctx, msg); err != nil {
				c.logger.Error().Err(err).Str("message_id", msg.ID).Msg("handle claimed message failed")
				continue
			}
			if err := c.client.XAck(ctx, c.stream, c.group, msg.ID).Err(); err != nil {
				c.logger.Error().Err(err).Str("message_id", msg.ID).Msg("ack claimed failed")
			}
		}
	}
	return nil
}
