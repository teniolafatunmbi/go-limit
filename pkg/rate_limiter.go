package pkg

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"time"
)

type WindowKey string
type WindowValue struct {
	key   string
	count int
}
type WindowMap map[WindowKey]*WindowValue

const (
	CurrentWindowKey  WindowKey = "current"
	PreviousWindowKey WindowKey = "previous"
)

var windowMap = make(WindowMap)

// move this into a config file
const WINDOW_IN_SECONDS = 60 // in seconds
const THRESHOLD = 40

func getIpFromRemoteAddr(remoteAddr string) (*string, error) {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &ip, nil
}

func calculateNumberOfRequestsInCurrentWindow(now time.Time, windowMap map[WindowKey]*WindowValue) int {

	startOfMinute := time.Date(
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), 0, 0, now.Location(),
	)

	// subtract the time between now and the start of the current minute
	secondsIntoTheCurrentWindow := int(now.Sub(startOfMinute).Seconds())

	overlapWeight := float64(WINDOW_IN_SECONDS-secondsIntoTheCurrentWindow) / float64(WINDOW_IN_SECONDS)

	// calculate the number of requests in this window
	numberOfRequestsInCurrentWindow := windowMap[CurrentWindowKey].count + (windowMap[PreviousWindowKey].count * int(math.Round(overlapWeight)))

	return numberOfRequestsInCurrentWindow
}

func advanceWindow(currentWindow string, windowMap *WindowMap) {
	(*windowMap)[PreviousWindowKey] = (*windowMap)[CurrentWindowKey]
	(*windowMap)[CurrentWindowKey] = &WindowValue{key: currentWindow, count: 0}
}

func initializeNewWindows(windowMap *WindowMap) {
	(*windowMap)[CurrentWindowKey] = &WindowValue{}
	(*windowMap)[PreviousWindowKey] = &WindowValue{}
}

func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RemoteAddr)
		ipAddress, err := getIpFromRemoteAddr(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		now := time.Now()
		currentWindow := fmt.Sprintf("%2d.%2d", now.Hour(), now.Minute())

		// if the map is empty, add the current window and previous window values
		_, currentWindowKeyExistsInMap := windowMap[CurrentWindowKey]
		_, previousWindowKeyExistsInMap := windowMap[PreviousWindowKey]

		if (currentWindowKeyExistsInMap && previousWindowKeyExistsInMap) == false {
			initializeNewWindows(&windowMap)
		}

		// if currentWindow is not the windowMap.current.key,
		// we're in a new window, so update the current and previous windows
		if currentWindow != windowMap[CurrentWindowKey].key {
			advanceWindow(currentWindow, &windowMap)
		}

		numberOfRequestsInCurrentWindow := calculateNumberOfRequestsInCurrentWindow(now, windowMap)

		fmt.Println("noOfRequestsInCurrentWindow.Sliding", numberOfRequestsInCurrentWindow)

		// if noOfRequestsInCurrentWindow == Zero, discard request, else handle
		if numberOfRequestsInCurrentWindow == THRESHOLD {
			http.Error(w, "Too many request", http.StatusTooManyRequests)
			return
		}

		windowMap[CurrentWindowKey].count++

		fmt.Print("current.window.inferred ", currentWindow, "\n")
		fmt.Print("current.window.map ", *windowMap[CurrentWindowKey], "\n")

		fmt.Print(*windowMap[CurrentWindowKey], *windowMap[PreviousWindowKey], "\n")

		w.Header().Set("x-ip-address", *ipAddress)
		w.Header().Set("x-rate-limit-remaining", fmt.Sprintf("%d", THRESHOLD-numberOfRequestsInCurrentWindow))

		next.ServeHTTP(w, r)
	})
}
