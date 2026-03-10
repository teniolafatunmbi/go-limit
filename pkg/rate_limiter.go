package pkg

import (
	"net/http"
	"sync"
	"teniolafatunmbi/go-limit/pkg/cache"
	"teniolafatunmbi/go-limit/pkg/handlers"
	"teniolafatunmbi/go-limit/pkg/logging"

	"github.com/redis/go-redis/v9"
)

type RateLimiterConfig struct {
	RedisClient *redis.Client
	Strategy    RateLimiterStrategy
}

type RateLimiterStrategy struct {
	Name            string
	Threshold       int
	WindowInSeconds int
	BucketSize      int
}

var cacheInstanceLock = &sync.Mutex{}

var cacheInstance *cache.Cache

func getCache(rdb *redis.Client) *cache.Cache {
	var logger = logging.GetNewLogger()
	if cacheInstance == nil {
		cacheInstanceLock.Lock()
		defer cacheInstanceLock.Unlock()

		if cacheInstance == nil {
			logger.Info("Creating cache instance")
			cacheInstance = cache.New(rdb)
		} else {
			logger.Info("Cache instance has already been created")
		}
	} else {
		logger.Info("Cache instance has already been created")

	}
	return cacheInstance
}

// setup a logger for the rate limiter so it can log with request ID
// accept redis client instead of cache and initialize the cache with the redis client in the RateLimiter
// function signature - redisClient, algo - token_bucket, sliding_window_counter, sliding_window_log
func RateLimiter(cfg RateLimiterConfig) func(http.Handler) http.Handler {
	// get cache initialization from memory, if no cache, initialize cache
	var cache = getCache(cfg.RedisClient)
	var logger = logging.GetNewLogger()
	return func(next http.Handler) http.Handler {
		switch cfg.Strategy.Name {
		case "sliding_window":
			return handlers.SlidingWindowCounter(cache, logger, next)
		case "token_bucket":
			return handlers.SlidingWindowCounter(cache, logger, next)
		default:
			return handlers.SlidingWindowCounter(cache, logger, next)
		}
	}
}
