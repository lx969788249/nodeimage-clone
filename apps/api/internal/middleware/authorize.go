package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nodeimage/api/internal/models"
)

func RequireRoles(roles ...models.UserRole) gin.HandlerFunc {
	roleSet := make(map[models.UserRole]struct{}, len(roles))
	for _, role := range roles {
		roleSet[role] = struct{}{}
	}

	return func(c *gin.Context) {
		userVal, exists := c.Get("current_user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		user, ok := userVal.(models.User)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_user"})
			return
		}

		if _, ok := roleSet[user.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}
