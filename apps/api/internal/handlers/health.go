package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type healthResponse struct {
	Status     string `json:"status"`
	Database   string `json:"database"`
	Cache      string `json:"cache"`
	Environment string `json:"environment"`
}

func (h HandlerSet) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "ok"
	if err := h.db.Ping(ctx); err != nil {
		dbStatus = "error"
		h.log.Error().Err(err).Msg("database ping failed")
	}

	cacheStatus := "ok"
	if err := h.cache.Ping(ctx).Err(); err != nil {
		cacheStatus = "error"
		h.log.Error().Err(err).Msg("redis ping failed")
	}

	c.JSON(http.StatusOK, healthResponse{
		Status:      "ok",
		Database:    dbStatus,
		Cache:       cacheStatus,
		Environment: h.cfg.Environment,
	})
}
