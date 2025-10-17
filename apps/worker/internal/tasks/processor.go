package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Processor struct {
	logger zerolog.Logger
}

type TaskPayload struct {
	Type    string                 `json:"type"`
	ImageID string                 `json:"imageId"`
	Data    map[string]interface{} `json:"data"`
}

func NewProcessor(logger zerolog.Logger) *Processor {
	return &Processor{
		logger: logger,
	}
}

func (p *Processor) Handle(ctx context.Context, msg redis.XMessage) error {
	var payload TaskPayload
	if err := decodePayload(msg.Values, &payload); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	switch payload.Type {
	case "ingest":
		return p.handleIngest(ctx, payload)
	case "thumbnail":
		return p.handleThumbnail(ctx, payload)
	case "nsfw":
		return p.handleNSFW(ctx, payload)
	case "cleanup":
		return p.handleCleanup(ctx, payload)
	default:
		p.logger.Warn().Str("type", payload.Type).Msg("unknown task type")
		return nil
	}
}

func decodePayload(values map[string]interface{}, out *TaskPayload) error {
	bytes, err := json.Marshal(values)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, out)
}

func (p *Processor) handleIngest(ctx context.Context, payload TaskPayload) error {
	p.logger.Info().Str("image_id", payload.ImageID).Msg("ingest task received (stub)")
	return nil
}

func (p *Processor) handleThumbnail(ctx context.Context, payload TaskPayload) error {
	p.logger.Info().Str("image_id", payload.ImageID).Msg("thumbnail task stub")
	return nil
}

func (p *Processor) handleNSFW(ctx context.Context, payload TaskPayload) error {
	p.logger.Info().Str("image_id", payload.ImageID).Msg("nsfw task stub")
	return nil
}

func (p *Processor) handleCleanup(ctx context.Context, payload TaskPayload) error {
	p.logger.Info().Msg("cleanup task stub")
	return nil
}
