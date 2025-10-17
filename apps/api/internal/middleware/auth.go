package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nodeimage/api/internal/config"
	"nodeimage/api/internal/repository"
	"nodeimage/api/internal/security"
)

func Auth(cfg *config.AppConfig, users *repository.UserRepository, sessions *repository.SessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_token"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := security.ParseAccessToken(tokenStr, cfg.Security.JWTAccessSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}

		session, err := sessions.GetByID(c.Request.Context(), claims.SessionID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session_not_found"})
			return
		}

		if session.UserID != claims.UserID || session.DeviceID != claims.DeviceID {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session_mismatch"})
			return
		}

		user, err := users.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user_not_found"})
			return
		}

		if user.Status != "active" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user_inactive"})
			return
		}

		_ = sessions.Touch(c.Request.Context(), session.ID, c.ClientIP(), c.GetHeader("User-Agent"))

		c.Set("access_token", tokenStr)
		c.Set("access_claims", *claims)
		c.Set("current_user", user)

		c.Next()
	}
}
