package handlers

import (
	"log/slog"
	"net/http"
	"teniolafatunmbi/go-limit/internal/cache"
)

func TokenBucket(cache *cache.Cache, logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
