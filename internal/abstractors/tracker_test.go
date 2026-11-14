package abstractors

import (
	"context"
	"testing"
	"time"

	"github.com/mgoulish/vision-society/internal/abstraction"
	"github.com/mgoulish/vision-society/internal/blackboard"
)

func TestSimpleBlobTracker_ReplayMode(t *testing.T) {
	// Create a channel and feed it a few synthetic RedBlob Abstractions
	input := make(chan abstraction.Abstraction, 10)

	// Send 3 RedBlobs with increasing frame IDs and sizes
	for i := 1; i <= 3; i++ {
		input <- abstraction.Abstraction{
			Type:    "RedBlob",
			FrameID: uint64(i),
			Payload: map[string]any{
				"size": 10.0 + float64(i),
				"x":    100.0 + float64(i),
			},
		}
	}
	close(input)

	bb := blackboard.New()
	tracker := NewSimpleBlobTrackerFromChannel(input)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Run the tracker in a goroutine
	done := make(chan struct{})
	go func() {
		tracker.Run(ctx, bb)
		close(done)
	}()

	// Wait for the tracker to finish
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("tracker did not finish in time")
	}

	// Check that the tracker posted TrackedBlob Abstractions
	count := 0
	for frame := uint64(1); frame <= 3; frame++ {
		results := bb.QueryByType("TrackedBlob", frame)
		count += len(results)
	}

	if count != 3 {
		t.Errorf("expected 3 TrackedBlob outputs, got %d", count)
	}
}
