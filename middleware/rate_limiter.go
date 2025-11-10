package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

// RateLimitEntry stores the timestamp of the last request by an IP.
type RateLimitEntry struct {
	LastRequest time.Time
	Count       int
}

// RateLimitMiddleware creates a Gin middleware to limit requests based on IP address.
// This example limits to 5 requests every 10 seconds.
func RateLimitMiddleware(c *cache.Cache, limit int, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()
		cacheKey := "rate_limit:" + clientIP

		// Load existing entry or create a new one
		var entry RateLimitEntry
		if cached, found := c.Get(cacheKey); found {
			entry = cached.(RateLimitEntry)
		} else {
			// First request in the window
			entry = RateLimitEntry{
				LastRequest: time.Now(),
				Count:       0, // Initialize count to 0, will increment to 1 below
			}
		}

		currentTime := time.Now()

		// Check if the current window has passed
		if currentTime.Sub(entry.LastRequest) >= window {
			// New window: Reset count and timestamp
			entry.Count = 1
			entry.LastRequest = currentTime
		} else {
			// Same window: Increment count
			entry.Count++
		}

		// Check the limit
		if entry.Count > limit {
			// Too many requests
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "Too many requests. Please try again later.",
			})
			return
		}

		// Save the updated entry (set to expire slightly after the window)
		c.Set(cacheKey, entry, window+1*time.Second)

		ctx.Next()
	}
}
