package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h HandlerSet) AdminListImages(c *gin.Context) {
	limit := 50
	offset := 0

	if perPage := c.Query("perPage"); perPage != "" {
		if v, err := strconv.Atoi(perPage); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	if page := c.Query("page"); page != "" {
		if v, err := strconv.Atoi(page); err == nil && v > 1 {
			offset = (v - 1) * limit
		}
	}

	images, err := h.images.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]map[string]interface{}, 0, len(images))
	for _, img := range images {
		items = append(items, map[string]interface{}{
			"id":         img.ID,
			"userId":     img.UserID,
			"format":     img.Format,
			"status":     img.Status,
			"sizeBytes":  img.SizeBytes,
			"visibility": img.Visibility,
			"nsfwScore":  img.NSFWScore,
			"createdAt":  img.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
	})
}
