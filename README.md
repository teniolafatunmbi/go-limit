# go-limit (Token Bucket algorithm)

## TODO:
- Write tests

## Implementation

- This implementation uses an in-memory bucket per user (identified by IP address), structured as follows:
```sh
    {
        "ip_address": [token, token, ..., token]
    }
```

- Each IP address maps to a slice (list) representing the tokens currently available in its bucket.

- Each new IP address gets initialized with 10 tokens.

- Tokens are replenished for each IP address individually:
  - A goroutine and ticker are started per IP when it is first seen.
  - Every second, as long as the bucket has fewer than 10 tokens, one token is added back.
  - If the bucket already has 10 tokens, no more are added (the bucket is full).

- All data is stored in memory. If the application restarts, all buckets reset.
