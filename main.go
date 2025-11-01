package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type TokenBucket struct {
	mu       sync.Mutex    
	capacity int64         
	tokens   int64         
	rate     int64         
	interval time.Duration 
	ticker   *time.Ticker
	stop     chan bool
}

func NewTokenBucket(rate int64, capacity int64, interval time.Duration) *TokenBucket {
	tb := &TokenBucket{
		capacity: capacity,
		tokens:   capacity,
		rate:     rate,
		interval: interval,
		ticker:   time.NewTicker(interval),
		stop:     make(chan bool),
	}

	go tb.refill()

	return tb
}

func (tb *TokenBucket) refill() {
	for {
		select {
		case <-tb.ticker.C:
			tb.mu.Lock()
			tb.tokens += tb.rate
			if tb.tokens > tb.capacity {
				tb.tokens = tb.capacity
			}
			log.Printf("Refilled tokens. Current count: %d\n", tb.tokens)
			tb.mu.Unlock()

		case <-tb.stop:
			tb.ticker.Stop()
			return
		}
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func (tb *TokenBucket) Stop() {
	tb.stop <- true
}

func main() {
	capacity := int64(10)
	rate := int64(1)
	interval := 2 * time.Second

	limiter := NewTokenBucket(rate, capacity, interval)
	defer limiter.Stop()
	http.HandleFunc("/limited", func(w http.ResponseWriter, r *http.Request) {
		if limiter.Allow() {
			log.Println("Request ALLOWED for /limited")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Request was processed.")
		} else {
			log.Println("Request DENIED for /limited")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintln(w, "Too Many Requests.")
		}
	})

	http.HandleFunc("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request ALLOWED for /unlimited")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Unlimited request was processed.")
	})

	log.Println("Starting rate limiter service on :8080...")
	log.Println("Test with: http://localhost:8080/limited")
	log.Println("Test with: http://localhost:8080/unlimited")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
