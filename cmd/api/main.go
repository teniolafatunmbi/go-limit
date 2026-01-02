package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"teniolafatunmbi/go-limit/pkg"
)

func main() {
	r := chi.NewRouter()
	PORT := ":8080"

	r.Use(middleware.Logger)
	r.Use(pkg.RateLimiter)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Unlimited! Let's Go!"))
	})

	r.Get("/limited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Limited, don't over use me!"))
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
