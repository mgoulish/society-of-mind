package scenarios

import (
	"context"
	"testing"
	"time"

	"github.com/mgoulish/vision-society/internal/abstraction"
	"github.com/mgoulish/vision-society/internal/blackboard"
)

func TestEmitter_BasicScenario(t *testing.T) {
	// Simple test scenario that emits exactly 3 RedBlobs
	testScenario := func(ctx context.Context) <-chan abstraction.Abstraction {
		ch := make(chan abstraction.Abstraction, 3)
		for i := 1; i <= 3; i++ {
			ch <- abstraction.Abstraction{
				Type:    "RedBlob",
				FrameID: uint64(i),
				Payload: map[string]any{
					"size": float64(i * 10),
				},
			}
		}
		close(ch)
		return ch
	}

	bb := blackboard.New()
	emitter := NewEmitter("TestEmitter", testScenario)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		emitter.Run(ctx, bb)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("emitter did not finish in time")
	}

	// Verify that 3 RedBlobs were posted
	count := 0
	for frame := uint64(1); frame <= 3; frame++ {
		results := bb.QueryByType("RedBlob", frame)
		count += len(results)
	}

	if count != 3 {
		t.Errorf("expected 3 RedBlobs posted, got %d", count)
	}
}
