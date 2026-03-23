package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthenticateToken middleware that validates bearer tokens against the tokens table
func AuthenticateToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must start with Bearer"})
			c.Abort()
			return
		}

		// Extract token string (no UUID validation needed - VARCHAR(45))
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
			c.Abort()
			return
		}

		// Get user from custom header
		user := c.GetHeader("X-User")
		if user == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "X-User header required"})
			c.Abort()
			return
		}

		// Validate token-user combination
		if !isValidTokenForUser(db, tokenString, user) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token for user"})
			c.Abort()
			return
		}

		// Store in context
		c.Set("user", user)
		c.Set("token", tokenString)

		c.Next()
	}
}

// isValidTokenForUser validates if the token is valid for the given user using the tokens table
func isValidTokenForUser(db *gorm.DB, token string, user string) bool {
	var count int64

	// Simple validation query against tokens table only
	err := db.Raw(`
        SELECT COUNT(*) 
        FROM tokens 
        WHERE token = ? 
          AND username = ?
    `, token, user).Scan(&count).Error

	if err != nil {
		return false
	}

	return count > 0
}
