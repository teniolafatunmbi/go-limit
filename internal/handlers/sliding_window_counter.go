package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"teniolafatunmbi/go-limit/internal/cache"
	"teniolafatunmbi/go-limit/internal/logging"
	"teniolafatunmbi/go-limit/internal/utils"
	"time"

	"github.com/go-chi/chi/v5/middleware"
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

// move this into a config file
const WINDOW_IN_SECONDS = 60 // in seconds
const THRESHOLD = 40

func calculateNumberOfRequestsInCurrentWindow(
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

	numberOfRequestsInCurrentWindow = currentWindowInRedisJson.Count + (previousWindowInRedisJson.Count * int(math.Round(overlapWeight)))

	return &numberOfRequestsInCurrentWindow, nil
}

func advanceWindow(ctx context.Context, cache *cache.Cache, currentWindow string) error {
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

func initializeCurrentAndPreviousWindows(
	ctx context.Context,
	cache *cache.Cache,
	currentWindowKey string,
) error {
	currentWindowValue, _ := json.Marshal(WindowValue{Key: currentWindowKey, Count: 0})
	defaultWindowValue, _ := json.Marshal(WindowValue{})
	logger := logging.FromContext(ctx)

	err := cache.SetCurrentWindow(ctx, currentWindowValue)
	if err != nil {
		logger.ErrorContext(ctx, "Error occurred while setting current window value")
		return err
	}

	err = cache.SetPreviousWindow(ctx, defaultWindowValue)

	if err != nil {
		logger.ErrorContext(ctx, "Error occurred while setting previous window value")
		return err
	}

	return nil
}

func SlidingWindowCounter(cache *cache.Cache, base *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := middleware.GetReqID(r.Context())
		logger := base.With(slog.String("request_id", requestId), slog.String("algorithm", "sliding_window_counter"))
		ctx := logging.WithContext(r.Context(), logger)

		ipAddress, err := utils.GetIpFromRemoteAddr(r.RemoteAddr)
		if err != nil {
			logger.ErrorContext(ctx, "Malformed request: Can't get IP address")
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		now := time.Now()
		currentWindow := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())

		// if the map is empty, add the current window and previous window values
		currWindowExists, _ := cache.DoesCurrentWindowExist(ctx)
		prevWindowExists, _ := cache.DoesPreviousWindowExist(ctx)

		if *currWindowExists == 0 && *prevWindowExists == 0 {
			logger.InfoContext(ctx, "no curr and prev windows. Initializing new windows...")
			err = initializeCurrentAndPreviousWindows(ctx, cache, currentWindow)

			if err != nil {
				logger.ErrorContext(ctx, "An error occurred: "+err.Error())
				http.Error(w, fmt.Sprintf("Internal server error: %s", err.Error()), http.StatusInternalServerError)
				return
			}
		}

		currentWindowInRedis, currWindowErr := cache.GetCurrentWindow(ctx)

		if currWindowErr != nil {
			logger.ErrorContext(ctx, "An error occurred: "+err.Error())
			http.Error(w, fmt.Sprintf("Internal server error: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		var currentWindowJsonInRedis WindowValue
		err = json.Unmarshal([]byte(currentWindowInRedis), &currentWindowJsonInRedis)

		if err != nil {
			logger.Error("An error occurred: " + err.Error())
			http.Error(w, fmt.Sprintf("Internal server error: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		logger.Info("Current window value in cache ", slog.Any("current_window_value_in_cache", currentWindowJsonInRedis))

		// if currentWindow is not the windowMap.current.key,
		// we're in a new window, so update the current and previous windows
		if currentWindow != currentWindowJsonInRedis.Key {
			logger.InfoContext(ctx, "current window doesn't match the current window key-value in store. Advancing window...")

			err = advanceWindow(ctx, cache, currentWindow)

			if err != nil {
				logger.ErrorContext(ctx, fmt.Sprintf("Internal server error: %s", err.Error()))
				http.Error(w, fmt.Sprintf("Internal server error: %s", err.Error()), http.StatusInternalServerError)
				return
			}
		}

		numberOfRequestsInCurrentWindow, err := calculateNumberOfRequestsInCurrentWindow(ctx, cache, now)
		logger.Info("noOfRequestsInCurrentWindow.Sliding", slog.Any("no_of_requests_in_current_window", *numberOfRequestsInCurrentWindow))

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
			logger.ErrorContext(ctx, "Internal server error", err)
			return
		}

		logger.Info("Current window", slog.Any("current_window", currWindowInRedisJson))

		w.Header().Set("x-ip-address", *ipAddress)
		w.Header().Set("x-rate-limit-remaining", fmt.Sprintf("%d", THRESHOLD-(*numberOfRequestsInCurrentWindow)))

		next.ServeHTTP(w, r)
	})
}
