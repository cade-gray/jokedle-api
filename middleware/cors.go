package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var allowedOrigins = map[string]struct{}{
	"http://localhost:5173":      {},
	"https://admin.cadegray.dev": {},
}

// CORS middleware that supports admin app origins and handles preflight requests.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if _, ok := allowedOrigins[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-User")
			c.Header("Access-Control-Max-Age", "600")
			c.Header("Vary", "Origin")
			c.Header("Vary", "Access-Control-Request-Method")
			c.Header("Vary", "Access-Control-Request-Headers")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
