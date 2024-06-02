package middleware

import "github.com/gin-gonic/gin"

func Headers() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Content-Security-Policy", "require-sri-for style script")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Access-Control-Allow-Methods", "POST, GET")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("Strict-Transport-Security", "max-age=315360000, includeSubdomains; preload")
		c.Header("Cache-Control", "no-cache, no-store")
		c.Header("Pulic-Key-Pins", "pin-sha256=base64==; max-age=315360000")
		c.Header("X-Powered-By", "")
		c.Header("Server", "")
		c.Next()
	}
}
