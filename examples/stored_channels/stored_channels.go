package main

import (
    "fmt"
    "sync"
    "time"
)

// StoredChannel is a generic type that manages a bounded buffer of items from an input channel.
// It handles requests for count or specific items via a request channel.
type StoredChannel[T any] struct {
    input     <-chan T          // Input channel for data to store
    requests  chan Request[T]   // Channel for incoming requests
    storage   []T               // Slice to store items, newest at index 0
    limit     int               // Maximum number of items to store
    mu        sync.Mutex        // Mutex for thread-safe access to storage
    closeChan chan struct{}     // Signal channel for closing
    wg        sync.WaitGroup    // WaitGroup to wait for the goroutine to finish
}

// Request represents a request to the StoredChannel.
// Use Kind to specify if it's a CountRequest or ItemRequest.
// Include a reply channel for the response.
type Request[T any] struct {
    Kind  RequestKind
    Index int             // Only used for ItemRequest
    Reply chan Response[T]
}

// RequestKind differentiates between request types.
type RequestKind int

const (
    CountRequest RequestKind = iota
    ItemRequest
)

// Response is sent back on the reply channel.
// For CountRequest: Count is set, Item and Err are nil/empty.
// For ItemRequest: Item is set if successful, Err if not.
type Response[T any] struct {
    Count int
    Item  T
    Err   error
}

// NewStoredChannel creates and starts a new StoredChannel.
// It returns the StoredChannel instance, which includes the Requests channel for sending requests.
// Call Close() when done to shut it down gracefully.
func NewStoredChannel[T any](input <-chan T, limit int) *StoredChannel[T] {
    if limit <= 0 {
        panic("limit must be positive")
    }

    sc := &StoredChannel[T]{
        input:     input,
        requests:  make(chan Request[T]),
        storage:   make([]T, 0, limit),
        limit:     limit,
        closeChan: make(chan struct{}),
    }

    sc.wg.Add(1)
    go sc.run()

    return sc
}

// run is the main loop that handles input and requests.
func (sc *StoredChannel[T]) run() {
    defer sc.wg.Done()

    for {
        select {
        case item, ok := <-sc.input:
            if !ok {
                // Input channel closed, but continue handling requests
                sc.input = nil
                continue
            }
            sc.mu.Lock()
            // Insert at front (index 0)
            sc.storage = append([]T{item}, sc.storage...)
            // Trim if over limit
            if len(sc.storage) > sc.limit {
                sc.storage = sc.storage[:sc.limit]
            }
            sc.mu.Unlock()

        case req := <-sc.requests:
            sc.mu.Lock()
            resp := Response[T]{}
            switch req.Kind {
            case CountRequest:
                resp.Count = len(sc.storage)
            case ItemRequest:
                if req.Index < 0 || req.Index >= len(sc.storage) {
                    resp.Err = fmt.Errorf("index %d out of bounds (0 to %d)", req.Index, len(sc.storage)-1)
                } else {
                    resp.Item = sc.storage[req.Index]
                }
            }
            sc.mu.Unlock()
            req.Reply <- resp // Send response

        case <-sc.closeChan:
            return // Shutdown
        }
    }
}

// Close shuts down the StoredChannel gracefully.
func (sc *StoredChannel[T]) Close() {
    close(sc.closeChan)
    sc.wg.Wait()
    close(sc.requests)
}

// Example usage
func main() {
    // Create an input channel
    input := make(chan int)
    sc := NewStoredChannel(input, 5) // Limit of 5 items

    // Simulate sending data
    go func() {
        for i := 1; i <= 10; i++ {
            input <- i
            time.Sleep(100 * time.Millisecond)
        }
        close(input)
    }()

    // Simulate requests from multiple goroutines
    var wg sync.WaitGroup
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 5; j++ {
                // Request count
                reply := make(chan Response[int])
                sc.requests <- Request[int]{Kind: CountRequest, Reply: reply}
                resp := <-reply
                fmt.Printf("Goroutine %d: Count = %d\n", id, resp.Count)

                // Request item at index 0 (newest)
                reply = make(chan Response[int])
                sc.requests <- Request[int]{Kind: ItemRequest, Index: 0, Reply: reply}
                resp = <-reply
                if resp.Err != nil {
                    fmt.Printf("Goroutine %d: Error = %v\n", id, resp.Err)
                } else {
                    fmt.Printf("Goroutine %d: Newest item = %d\n", id, resp.Item)
                }

                time.Sleep(200 * time.Millisecond)
            }
        }(i)
    }

    wg.Wait()
    sc.Close()
}
