package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var windowToCounterMap = make(map[string]int)

const counterThreshold = 60

func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Println(windowToCounterMap)
		currentTime := time.Now()
		currentWindow := fmt.Sprintf("%02d.%02d", currentTime.Hour(), currentTime.Minute())

		_, exists := windowToCounterMap[currentWindow]

		if !exists {
			windowToCounterMap[currentWindow] = 0
		}

		// counter >= threshold, discard request
		if windowToCounterMap[currentWindow] >= counterThreshold {
			fmt.Printf("See window counter: %v\n", windowToCounterMap)
			http.Error(w, "Too many request", http.StatusTooManyRequests)
			return
		}

		// increment the window counter
		windowToCounterMap[currentWindow] += 1

		next.ServeHTTP(w, r)
	})
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(RateLimiter)

	r.Get("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Unlimited! Let's Go!"))
	})

	r.Get("/limited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Limited, don't over use me!"))
	})

	http.ListenAndServe("localhost:8080", r)
}
