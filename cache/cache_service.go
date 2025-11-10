package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// Cache is a pointer to the go-cache instance used throughout the application.
var Cache *cache.Cache

// InitCache initializes the in-memory cache.
func InitCache() {
	// Initialize cache: 5 minutes default expiration, 10 minutes cleanup interval
	Cache = cache.New(5*time.Minute, 10*time.Minute)
}
