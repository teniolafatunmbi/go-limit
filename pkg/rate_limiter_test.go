package pkg

import (
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

func TestGetTotalRequestsInCurrentWindow(t *testing.T) {
	var timestamps []time.Time
	currentTime := time.Now()
	const REQUESTS_IN_LAST_WINDOW_COUNT = 15

	// 10 timestamps for two minutes in the last 3 minutes.
	for i := 10; i < 20; i++ {
		timestamps = append(timestamps, currentTime.Add(time.Duration(-i*12)*time.Second))
	}

	// 15 timestamps for the last one minute
	for i := range REQUESTS_IN_LAST_WINDOW_COUNT {
		timestamps = append(timestamps, currentTime.Add(time.Duration(-i*4)*time.Second))
	}

	totalRequestsInCurrentWindow := getTotalRequestsInCurrentWindow(timestamps, WINDOW)

	assert.Equal(t, REQUESTS_IN_LAST_WINDOW_COUNT, totalRequestsInCurrentWindow)

}
