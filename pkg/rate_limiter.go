package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"teniolafatunmbi/go-limit/pkg/cache"
	"time"
)

type WindowKey string
type WindowValue struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
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

func (windowMap *WindowMap) CalculateNumberOfRequestsInCurrentWindow(
	ctx context.Context,
	cache *cache.Cache,
	now time.Time) (*int, error) {
	// read the current and previous window key values from redis
	startOfMinute := time.Date(
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), 0, 0, now.Location(),
	)

	// subtract the time between now and the start of the current minute
	secondsIntoTheCurrentWindow := int(now.Sub(startOfMinute).Seconds())

	overlapWeight := float64(WINDOW_IN_SECONDS-secondsIntoTheCurrentWindow) / float64(WINDOW_IN_SECONDS)

	// calculate the number of requests in this window
	var numberOfRequestsInCurrentWindow int

	// get current window value
	currentWindowInRedis, err := cache.GetCurrentWindow(ctx)
	previousWindowInRedis, err := cache.GetPreviousWindow(ctx)

	var currentWindowInRedisJson WindowValue
	var previousWindowInRedisJson WindowValue

	json.Unmarshal([]byte(currentWindowInRedis), &currentWindowInRedisJson)
	json.Unmarshal([]byte(previousWindowInRedis), &previousWindowInRedisJson)

	if err != nil {
		return &numberOfRequestsInCurrentWindow, err
	}

	// currentWindowCountToInt, err := strconv.Atoi(previousWindowInRedis)

	// if err != nil {
	// 	return &numberOfRequestsInCurrentWindow, err
	// }

	// previousWindowCountToInt, err := strconv.Atoi(previousWindowCount)

	// if err != nil {
	// 	return &numberOfRequestsInCurrentWindow, err
	// }

	numberOfRequestsInCurrentWindow = currentWindowInRedisJson.Count + (previousWindowInRedisJson.Count * int(math.Round(overlapWeight)))

	return &numberOfRequestsInCurrentWindow, nil
}

func (windowMap *WindowMap) AdvanceWindow(ctx context.Context, cache *cache.Cache, currentWindow string) error {
	currentWindowInRedis, err := cache.GetCurrentWindow(ctx)

	if err != nil {
		return err
	}

	// set the previous window key in redis to the current window value
	err = cache.SetPreviousWindow(ctx, currentWindowInRedis)

	if err != nil {
		return err
	}

	newCurrentWindowValue, err := json.Marshal(WindowValue{Key: currentWindow, Count: 0})

	if err != nil {
		return err
	}

	err = cache.SetCurrentWindow(ctx, newCurrentWindowValue)

	if err != nil {
		return err
	}

	return nil
}

func (windowMap *WindowMap) InitializeCurrentAndPreviousWindows(
	ctx context.Context,
	cache *cache.Cache,
	currentWindowKey string,
) error {
	currentWindowValue, _ := json.Marshal(WindowValue{Key: currentWindowKey, Count: 0})
	defaultWindowValue, _ := json.Marshal(WindowValue{})

	err := cache.SetCurrentWindow(ctx, currentWindowValue)
	if err != nil {
		return err
	}

	err = cache.SetPreviousWindow(ctx, defaultWindowValue)

	if err != nil {
		return err
	}
	return nil
}

// setup a logger for the rate limiter so it can log with request ID
func RateLimiter(cache *cache.Cache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ipAddress, err := getIpFromRemoteAddr(r.RemoteAddr)
			if err != nil {
				http.Error(w, "Malformed request", http.StatusBadRequest)
				return
			}

			ctx := context.Background()

			now := time.Now()
			currentWindow := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
			fmt.Println(currentWindow, "got here")

			// if the map is empty, add the current window and previous window values
			currWindowExists, _ := cache.DoesCurrentWindowExist(ctx)
			prevWindowExists, _ := cache.DoesPreviousWindowExist(ctx)

			if *currWindowExists == 0 && *prevWindowExists == 0 {
				fmt.Println("no curr and prev windows. Initializing new windows...")
				// set current window to `currentWindow`
				err = windowMap.InitializeCurrentAndPreviousWindows(ctx, cache, currentWindow)

				if err != nil {
					http.Error(w, fmt.Sprintf("Internal server error: %s", err.Error()), http.StatusInternalServerError)
					return
				}
			}

			currentWindowInRedis, currWindowErr := cache.GetCurrentWindow(ctx)

			if currWindowErr != nil {
				http.Error(w, fmt.Sprintf("Internal server error: %s", err.Error()), http.StatusInternalServerError)
				return
			}

			var currentWindowJsonInRedis WindowValue
			err = json.Unmarshal([]byte(currentWindowInRedis), &currentWindowJsonInRedis)

			if err != nil {
				http.Error(w, fmt.Sprintf("Internal server error: %s", err.Error()), http.StatusInternalServerError)
				return
			}

			fmt.Println(currentWindowJsonInRedis, "curr window val in redis")

			// if currentWindow is not the windowMap.current.key,
			// we're in a new window, so update the current and previous windows
			if currentWindow != currentWindowJsonInRedis.Key {
				fmt.Println("current window doesn't match the current window key-value in store. Advancing window...")

				err = windowMap.AdvanceWindow(ctx, cache, currentWindow)

				if err != nil {
					http.Error(w, fmt.Sprintf("Internal server error: %s", err.Error()), http.StatusInternalServerError)
					return
				}
			}

			numberOfRequestsInCurrentWindow, err := windowMap.CalculateNumberOfRequestsInCurrentWindow(ctx, cache, now)

			fmt.Println("noOfRequestsInCurrentWindow.Sliding", *numberOfRequestsInCurrentWindow)

			// if noOfRequestsInCurrentWindow == threshold, discard request, else handle
			if (*numberOfRequestsInCurrentWindow) >= THRESHOLD {
				http.Error(w, "Too many request", http.StatusTooManyRequests)
				return
			}

			currWindowInRedis, err := cache.GetCurrentWindow(ctx)

			var currWindowInRedisJson WindowValue

			json.Unmarshal([]byte(currWindowInRedis), &currWindowInRedisJson)

			currWindowInRedisJson.Count++

			marshalledCurrWindow, err := json.Marshal(currWindowInRedisJson)

			err = cache.SetCurrentWindow(ctx, marshalledCurrWindow)

			if err != nil {
				log.Fatalf("Internal server error", err)
				return
			}

			fmt.Print("current_window", currWindowInRedisJson, "\n")

			w.Header().Set("x-ip-address", *ipAddress)
			w.Header().Set("x-rate-limit-remaining", fmt.Sprintf("%d", THRESHOLD-(*numberOfRequestsInCurrentWindow)))

			next.ServeHTTP(w, r)
		})

	}
}
