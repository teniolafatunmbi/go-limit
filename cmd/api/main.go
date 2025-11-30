package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"teniolafatunmbi/go-limit/pkg"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(pkg.RateLimiter)

	r.Get("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Unlimited! Let's Go!"))
	})

	r.Get("/limited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Limited, don't over use me!"))
	})

	http.ListenAndServe("localhost:8080", r)
}
