package pkg

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

// maybe use a go func that runs in 1 minute intervals to clear logs that fall out of the
// set sliding `window` interval
var bucket []time.Time

const WINDOW = 60 * time.Second
const THRESHOLD = 60

func getIpFromRemoteAddr(remoteAddr string) (*string, error) {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &ip, nil
}

func getTotalRequestsInCurrentWindow(bucket []time.Time, window time.Duration) int {
	noOfRequestsInCurrentWindow := 0

	// start iterating from the end because new request timestamps are appended to the array
	for i := len(bucket) - 1; i >= 0; i-- {
		ts := bucket[i]
		if time.Since(ts) < window {
			noOfRequestsInCurrentWindow += 1
		}
	}

	return noOfRequestsInCurrentWindow
}

func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RemoteAddr)
		ipAddress, err := getIpFromRemoteAddr(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		w.Header().Set("x-ip-address", *ipAddress)

		requestTimestamp := time.Now()

		// calculate the number of requests in the current <window>
		noOfRequestsInCurrentWindow := getTotalRequestsInCurrentWindow(bucket, WINDOW)

		fmt.Println("noOfRequestsInCurrentWindow", noOfRequestsInCurrentWindow)

		// if noOfRequestsInCurrentWindow = threshold, discard request, else handle
		if noOfRequestsInCurrentWindow == THRESHOLD {
			http.Error(w, "Too many request", http.StatusTooManyRequests)
			return
		}

		bucket = append(bucket, requestTimestamp)

		fmt.Println(bucket)
		next.ServeHTTP(w, r)
	})
}
