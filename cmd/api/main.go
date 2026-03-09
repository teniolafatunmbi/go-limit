package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"teniolafatunmbi/go-limit/internal/redis"
	"teniolafatunmbi/go-limit/pkg"
	"teniolafatunmbi/go-limit/pkg/cache"
)

func main() {
	r := chi.NewRouter()
	PORT := ":8080"

	// redis configuration
	rdb := redis.NewRedisClient()
	cache := cache.New(rdb)
	rateLimiter := pkg.RateLimiter(cache)

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Unlimited! Let's Go!"))
	})

	r.Route("/limited", func(limited chi.Router) {
		limited.Use(rateLimiter)
		limited.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Limited, don't over use me!"))
		})
	})

	fmt.Printf("Server starting on port %s\n", PORT)
	server := &http.Server{
		Addr:         PORT,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}

}
