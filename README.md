# Cache Proxy
A simple cache proxy.
- HTTP GET and HEAD methods are supported.
- Caches HTTP response body and headers.

### Prerequisites
Install redis-server on your local machine.

### Build
```bash
go build .
```

### Run

Start the redis server on your local machine.
```bash
redis-server
```

Run the cache-proxy binary.
```bash
./cache-proxy --origin REMOTE_HOST --port LOCAL_PORT
```

### Example
Start the cache proxy.
```bash
./cache-proxy --origin https://jsonplaceholder.typicode.com --port 3000
```

Make a GET request.
```bash
curl -v https://localhost:3000/posts/1
```

Notice the X-CACHE header for the first request.
```
...
X-Cache: MISS
...
```

Make the same request again.
```bash
curl -v https://localhost:3000/posts/1
```

Notice the X-CACHE header.
```
...
X-Cache: HIT
...
```

Check the logs.
```
2025/04/28 23:37:36 GET /posts/1 cache
```

