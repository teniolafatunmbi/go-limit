package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"teniolafatunmbi/go-limit/pkg"
)

func main() {
	r := chi.NewRouter()
	PORT := 8080

	r.Use(middleware.Logger)
	r.Use(pkg.RateLimiter)

	r.Get("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Unlimited! Let's Go!"))
	})

	r.Get("/limited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Limited, don't over use me!"))
	})

	fmt.Printf("Server started at port: %d", PORT)
	http.ListenAndServe(fmt.Sprintf("localhost:%d", PORT), r)
}
