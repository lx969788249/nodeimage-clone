package jobs

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
)

type Scheduler struct {
	cron  *cron.Cron
	queue *redis.Client
	log   zerolog.Logger
}

func NewScheduler(queue *redis.Client, log zerolog.Logger) *Scheduler {
	c := cron.New(cron.WithSeconds())
	return &Scheduler{
		cron:  c,
		queue: queue,
		log:   log,
	}
}

func (s *Scheduler) Start() error {
	if s.queue == nil {
		return nil
	}

	if _, err := s.cron.AddFunc("0 0 0 * * *", s.enqueueCleanup); err != nil {
		return err
	}
	if _, err := s.cron.AddFunc("0 0 */1 * * *", s.enqueueReview); err != nil { // hourly recheck
		return err
	}

	s.cron.Start()
	return nil
}

func (s *Scheduler) Stop() context.CancelFunc {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	go func() {
		s.cron.Stop()
		cancel()
	}()
	return cancel
}

func (s *Scheduler) enqueueCleanup() {
	if err := s.enqueueTask(map[string]any{
		"type": "cleanup",
	}); err != nil {
		s.log.Error().Err(err).Msg("enqueue cleanup failed")
	}
}

func (s *Scheduler) enqueueReview() {
	if err := s.enqueueTask(map[string]any{
		"type": "nsfw",
		"scope": "review",
	}); err != nil {
		s.log.Error().Err(err).Msg("enqueue review failed")
	}
}

func (s *Scheduler) enqueueTask(payload map[string]any) error {
	if s.queue == nil {
		return nil
	}
	_, err := s.queue.XAdd(context.Background(), &redis.XAddArgs{
		Stream: "media:ingest",
		Values: payload,
	}).Result()
	return err
}
