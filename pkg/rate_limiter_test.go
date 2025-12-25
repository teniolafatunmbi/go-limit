package pkg

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIpFromRemoteAddr(t *testing.T) {
	testCases := []struct {
		name       string
		remoteAddr string
		expected   string
	}{
		{
			name:       "valid remote address",
			remoteAddr: "127.0.0.1:50448",
			expected:   "127.0.0.1",
		},
		{
			name:       "remote address with double colon",
			remoteAddr: "127.0.0.1::50448",
			expected:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ipAddress, err := getIpFromRemoteAddr(tc.remoteAddr)

			t.Log(ipAddress, err)
			if err != nil {
				assert.Equal(t, tc.name, "remote address with double colon")
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, *ipAddress)
		})
	}
}

func TestInitializeNewWindows(t *testing.T) {
	windowMap := make(WindowMap)

	windowMap.InitializeNewWindows()

	assert.Contains(t, windowMap, CurrentWindowKey)
	assert.Contains(t, windowMap, PreviousWindowKey)
	assert.IsType(t, &WindowValue{}, windowMap[CurrentWindowKey])
	assert.IsType(t, &WindowValue{}, windowMap[PreviousWindowKey])
}

func TestAdvanceWindow(t *testing.T) {
	// initialize window
	windowMap := make(WindowMap)

	windowMap.InitializeNewWindows()

	now := time.Now()

	lastMinuteWindow := fmt.Sprintf("%2d.%2d", now.Hour(), now.Minute()-1)
	currentWindow := fmt.Sprintf("%2d.%2d", now.Hour(), now.Minute())

	windowMap[CurrentWindowKey].key = lastMinuteWindow

	windowMap.AdvanceWindow(currentWindow)

	assert.Equal(t, currentWindow, windowMap[CurrentWindowKey].key)
	assert.Equal(t, lastMinuteWindow, windowMap[PreviousWindowKey].key)
}

func TestCalculateNumberOfRequestsInCurrentWindow_WindowRequestIsAlwaysLessThanThreshold(t *testing.T) {
	windowMap := make(WindowMap)

	windowMap.InitializeNewWindows()

	now := time.Now()
	lastMinuteWindow := fmt.Sprintf("%2d.%2d", now.Hour(), now.Minute()-1)
	currentWindow := fmt.Sprintf("%2d.%2d", now.Hour(), now.Minute())

	//  populate previous window with the max number of requests
	windowMap[PreviousWindowKey].key = lastMinuteWindow
	windowMap[PreviousWindowKey].count = THRESHOLD

	//  add 10 requests to the current window
	windowMap[CurrentWindowKey].key = currentWindow
	windowMap[CurrentWindowKey].count = 10

	// add 30 seconds to the current timestamp to simulate passing time in the current window
	now = now.Add(30 * time.Second)

	noOfRequestsInCurrentWindow := windowMap.CalculateNumberOfRequestsInCurrentWindow(now)

	t.Logf("noOfRequestsInCurrentWindow: %d\n", noOfRequestsInCurrentWindow)

	// assert that the number of requests in current window is less than the threshold
	assert.Less(t, noOfRequestsInCurrentWindow, THRESHOLD)
}
