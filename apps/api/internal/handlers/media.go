package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"nodeimage/api/internal/models"
	"nodeimage/api/internal/security"
	"nodeimage/api/internal/service"
)

type uploadResponse struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	Format    string    `json:"format"`
	SizeBytes int64     `json:"sizeBytes"`
	CreatedAt time.Time `json:"createdAt"`
}

func (h HandlerSet) UploadMedia(c *gin.Context) {
	userVal, exists := c.Get("current_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, ok := userVal.(models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_user"})
		return
	}

	claimsVal, exists := c.Get("access_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing_claims"})
		return
	}
	claims, ok := claimsVal.(security.AccessClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_claims"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_required"})
		return
	}
	defer file.Close()

	var expireAt *time.Time
	if expires := c.PostForm("expireAt"); expires != "" {
		if parsed, err := time.Parse(time.RFC3339, expires); err == nil {
			expireAt = &parsed
		}
	}

	result, err := h.uploadService.Upload(c.Request.Context(), service.UploadInput{
		User:       user,
		DeviceID:   claims.DeviceID,
		File:       file,
		Header:     header,
		Visibility: c.PostForm("visibility"),
		ExpireAt:   expireAt,
	})
	if err != nil {
		h.log.Error().Err(err).Str("user_id", user.ID).Msg("upload failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := uploadResponse{
		ID:        result.Image.ID,
		URL:       result.URL,
		Status:    string(result.Image.Status),
		Format:    result.Image.Format,
		SizeBytes: result.Image.SizeBytes,
		CreatedAt: result.Image.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"image": resp,
	})
}
