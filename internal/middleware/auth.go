package middleware

import (
	"net/http"
	"strings"

	"github.com/attendance/backend/internal/config"
	"github.com/attendance/backend/internal/utils"
	"github.com/attendance/backend/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header required", nil)
			c.Abort()
			return
		}

		// Extract token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format", nil)
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Validate token
		claims, err := jwt.ValidateToken(token, cfg.JWT.Secret)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token", err.Error())
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// AdminMiddleware checks if user is admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
			c.Abort()
			return
		}

		if role != "admin" {
			utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
