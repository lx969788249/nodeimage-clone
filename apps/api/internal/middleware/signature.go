package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"nodeimage/api/internal/config"
	"nodeimage/api/internal/security"
)

func NewReadCloser(b []byte) http.ReadCloser {
	return io.NopCloser(bytes.NewReader(b))
}

func Signature(cfg *config.AppConfig, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		date, nonce, signature, err := security.ExtractSignatureHeaders(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "signature_required"})
			return
		}

		requestTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_date"})
			return
		}

		if time.Since(requestTime) > 5*time.Minute || time.Until(requestTime) > 2*time.Minute {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "request_expired"})
			return
		}

		rawBody, err := c.GetRawData()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
			return
		}
		c.Request.Body = NewReadCloser(rawBody)

		claims, ok := c.Get("access_claims")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_access_claims"})
			return
		}

		accessClaims, ok := claims.(security.AccessClaims)
		if !ok {
			if ptr, ok2 := claims.(*security.AccessClaims); ok2 && ptr != nil {
				accessClaims = *ptr
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_access_claims"})
				return
			}
		}

		path, query := security.CanonicalPath(c.Request)
		valid := security.ValidateSignature(
			cfg.Security.SignatureSecret,
			accessClaims.DeviceID,
			signature,
			c.Request.Method,
			path,
			query,
			rawBody,
			date,
			nonce,
		)
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_signature"})
			return
		}

		nonceKey := fmt.Sprintf("sig:%s:%s", accessClaims.DeviceID, nonce)
		if ok := redisClient.SetNX(c, nonceKey, "1", 5*time.Minute); !ok.Val() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "replay_detected"})
			return
		}

		c.Next()
	}
}
