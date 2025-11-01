package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// ## 1. The TokenBucket Struct
// This holds the state of our rate limiter.
type TokenBucket struct {
	mu       sync.Mutex    // A mutex to make it concurrency-safe
	capacity int64         // Max tokens the bucket can hold
	tokens   int64         // Current number of tokens
	rate     int64         // Number of tokens to add per interval
	interval time.Duration // The duration between token refills
	ticker   *time.Ticker  // The ticker that drives the refills
	stop     chan bool     // Channel to stop the refill goroutine
}

// ## 2. The Constructor
// NewTokenBucket creates a new TokenBucket and starts its refill goroutine.
func NewTokenBucket(rate int64, capacity int64, interval time.Duration) *TokenBucket {
	// Initialize the bucket with full capacity
	tb := &TokenBucket{
		capacity: capacity,
		tokens:   capacity, // Start full
		rate:     rate,
		interval: interval,
		ticker:   time.NewTicker(interval),
		stop:     make(chan bool),
	}

	// Start the background goroutine to refill tokens
	go tb.refill()

	return tb
}

// ## 3. The Refill Goroutine
// refill runs in the background, adding tokens at each tick.
func (tb *TokenBucket) refill() {
	for {
		select {
		case <-tb.ticker.C:
			// A tick has occurred, time to add tokens
			tb.mu.Lock()
			// Add tokens, but don't exceed the capacity
			tb.tokens += tb.rate
			if tb.tokens > tb.capacity {
				tb.tokens = tb.capacity
			}
			log.Printf("Refilled tokens. Current count: %d\n", tb.tokens)
			tb.mu.Unlock()

		case <-tb.stop:
			// Stop signal received
			tb.ticker.Stop()
			return
		}
	}
}

// ## 4. The Core Logic: Allow()
// Allow checks if a request can be processed. It is concurrency-safe.
func (tb *TokenBucket) Allow() bool {
	// Lock the mutex to safely check and update the token count
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Check if there are tokens available
	if tb.tokens > 0 {
		// Yes, consume one token
		tb.tokens--
		return true
	}

	// No tokens, deny the request
	return false
}

// Stop gracefully shuts down the ticker and refill goroutine.
func (tb *TokenBucket) Stop() {
	tb.stop <- true
}

// ## 5. The Microservice Implementation
func main() {
	// --- Configuration ---
	// Our bucket will have a capacity of 10 tokens.
	// It will refill 1 token every 2 seconds.
	// This allows for a burst of 10 requests, then 1 request every 2 seconds.
	capacity := int64(10)
	rate := int64(1)
	interval := 2 * time.Second

	// Create our global rate limiter
	limiter := NewTokenBucket(rate, capacity, interval)
	defer limiter.Stop() // Ensure we stop the goroutine on exit

	// --- HTTP Handlers ---

	// A limited endpoint
	http.HandleFunc("/limited", func(w http.ResponseWriter, r *http.Request) {
		// Check if the limiter allows the request
		if limiter.Allow() {
			// Request is allowed
			log.Println("Request ALLOWED for /limited")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Request was processed.")
		} else {
			// Request is denied
			log.Println("Request DENIED for /limited")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintln(w, "Too Many Requests.")
		}
	})

	// An unlimited endpoint for comparison
	http.HandleFunc("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request ALLOWED for /unlimited")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Unlimited request was processed.")
	})

	// --- Start the Server ---
	log.Println("Starting rate limiter service on :8080...")
	log.Println("Test with: http://localhost:8080/limited")
	log.Println("Test with: http://localhost:8080/unlimited")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}