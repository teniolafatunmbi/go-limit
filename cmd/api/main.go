package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var bucket = make(map[string][]string)

func getIp(str string) string {
	ip := str
	if strings.Contains(str, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}

func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RemoteAddr)
		ipAddress := getIp(r.RemoteAddr)

		w.Header().Set("x-ip-address", ipAddress)

		fmt.Println(bucket)

		// if the ipAddress doesn't key exists in the bucket, create it
		_, exists := bucket[ipAddress]

		if !exists {
			bucket[ipAddress] = []string{"x", "x", "x", "x", "x", "x", "x", "x", "x", "x"}

			ticker := time.NewTicker(time.Second)

			// goroutine to add to the ip address bucket and check if the bucket is full
			go func() {
				for {
					t := <-ticker.C
					fmt.Printf("Tick at: %s for %s \n", t, ipAddress)
					if len(bucket[ipAddress]) == 10 {
						fmt.Printf("%s Bucket is full: %s \n", ipAddress, bucket[ipAddress])
					} else {
						bucket[ipAddress] = append(bucket[ipAddress], "x")
						fmt.Print(bucket)
					}
				}
			}()
		}

		if len(bucket[ipAddress]) == 0 {
			http.Error(w, "Too many request", http.StatusTooManyRequests)
			return
		}

		// Remove one token from the IP's bucket and serve the request
		bucket[ipAddress] = bucket[ipAddress][:len(bucket[ipAddress])-1]

		fmt.Printf("Popped one. Remaining %s", bucket[ipAddress])
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
