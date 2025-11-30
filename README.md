## go-limit

## Prerequisites
- Go (v1.23.1+)

## Implementation

- This implementation employs an in-memory bucket to store the timestamps of all incoming requests:
```sh
    [timestamp1, timestamp2, ..., timestamp<n>]
```
- Each request's timestamp is checked against a 60-second sliding window to count how many requests have occurred within that window.
- If the number of requests in this window reaches a threshold of 60, the API responds with "Too many requests" error, preventing any further processing of the request.
- The rate limiting logic is encapsulated within the `RateLimiter` middleware which appends each incoming request timestamp to the bucket and filters out timestamps outside the current window.
