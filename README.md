# Go Token Bucket Rate Limiter

This project is a simple microservice that demonstrates a **Token Bucket rate limiter** built from scratch in Go. It includes a concurrency-safe `TokenBucket` library and a simple HTTP server to show it in action.

The Token Bucket algorithm is a common and effective way to enforce rate limits. It allows for **bursts** of traffic up to the bucket's capacity, then throttles requests to a fixed **refill rate**.

## ✨ Features

  * **Token Bucket Algorithm:** Implements the token bucket algorithm from scratch.
  * **Concurrency-Safe:** Uses a `sync.Mutex` to ensure that the token count is handled safely across many simultaneous requests (goroutines).
  * **Graceful Shutdown:** Uses a `stop` channel to gracefully shut down the background refill goroutine.
  * **HTTP Microservice:** Wraps the limiter in a simple HTTP server with `/limited` and `/unlimited` endpoints to demonstrate its use.

-----

## 🚀 How to Run

### Prerequisites

  * Go (Version 1.18 or newer)

### 1\. Initialize Your Go Module

In your project's directory, run this command to create a `go.mod` file:

```bash
go mod init rate-limiter
```

### 2\. Run the Server

Execute the `main.go` file. The server will start and log its status.

```bash
go run .
```

You will see the following output, and your service will be running:

```bash
Starting rate limiter service on :8080...
Test with: http://localhost:8080/limited
Test with: http://localhost:8080/unlimited
```

-----

## 🧪 How to Test

You can test the rate limiter in two ways.

### 1\. Manual Test (Browser)

1.  Open your browser and go to `http://localhost:8080/limited`.
2.  The page will load, and you'll see "Request was processed."
3.  Refresh the page as fast as you can. You will be able to load it 10 times (the bucket's `capacity`).
4.  After the 10th request, your browser will show "Too Many Requests."
5.  Wait 2 seconds (the `interval`), and you will be able to make one more successful request.

### 2\. Automatic Test (Terminal)

This is the best way to see the limiter working. Open a **new terminal** (leave the server running in the first one) and run this `curl` loop. It will try to make a request every 0.2 seconds.

```bash
while true; do curl -w "\n" http://localhost:8080/limited; sleep 0.2; done
```

You will clearly see the first 10 "burst" requests get processed, followed by a series of "Too Many Requests," with one successful request allowed every 2 seconds when the bucket refills.

#### Example Output:

```
Request was processed.
Request was processed.
... (8 more times) ...
Request was processed.
Too Many Requests.
Too Many Requests.
Too Many Requests.
Too Many Requests.
Too Many Requests.
Request was processed.  <-- (Refilled after 2 seconds)
Too Many Requests.
Too Many Requests.
...
```

-----

## 🧠 Key Concepts

### 1\. The Token Bucket Algorithm

This implementation works like a real-world bucket.

  * **`capacity` (10):** The bucket can hold a maximum of 10 tokens.
  * **`rate` (1) & `interval` (2s):** A background process adds 1 token to the bucket every 2 seconds, as long as it's not full.
  * **`Allow()` function:** When a request comes in, the `Allow()` function checks if the bucket has at least 1 token.
      * If **yes**, it "takes" 1 token from the bucket and returns `true` (allowing the request).
      * If **no**, it returns `false` (denying the request).

This allows for a "burst" of 10 requests at once (emptying the bucket), after which the system is limited to the refill rate of 1 request every 2 seconds.

### 2\. Concurrency-Safe (`sync.Mutex`)

The `TokenBucket` struct has a `mu sync.Mutex`. Both the `Allow()` function (removing tokens) and the `refill()` goroutine (adding tokens) **must lock this mutex** before they can read or write the `tokens` variable.

This `Mutex` acts like a "talking stick," ensuring that only one function can modify the token count at a time, preventing race conditions.

### 3\. Background Goroutine (`go tb.refill()`)

When `NewTokenBucket` is called, it immediately starts a new **goroutine** that runs the `refill()` function in the background. This goroutine runs for the entire lifetime of the application.

### 4\. The Ticker (`time.Ticker`)

The `refill()` goroutine uses a `time.Ticker`. This is a Go construct that sends a "tick" (a message) down a channel at a fixed interval (in our case, every 2 seconds). The goroutine blocks and waits for the next tick, at which point it wakes up, adds a token, and goes back to waiting.
