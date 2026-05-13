package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestLogger middleware
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if status >= 400 {
			gin.DefaultWriter.Write([]byte(
				time.Now().Format("2006/01/02 - 15:04:05") + " | " + path + " | " + latency.String() + " | " + string(rune(status)) + "\n",
			))
		}
	}
}

// RateLimiter middleware (simple implementation)
func RateLimiter() gin.HandlerFunc {
	// Simple in-memory rate limiting
	type client struct {
		count     int
		resetTime time.Time
	}

	clients := make(map[string]*client)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		now := time.Now()

		if c, exists := clients[ip]; exists {
			if now.After(c.resetTime) {
				c.count = 0
				c.resetTime = now.Add(1 * time.Minute)
			}

			if c.count >= 60 { // 60 requests per minute
				c.AbortWithStatus(429)
				return
			}

			c.count++
		} else {
			clients[ip] = &client{
				count:     1,
				resetTime: now.Add(1 * time.Minute),
			}
		}

		c.Next()
	}
}