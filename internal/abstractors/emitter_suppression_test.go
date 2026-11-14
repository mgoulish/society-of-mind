package abstractors

import (
	"context"
	"testing"
	"time"

	"github.com/mgoulish/vision-society/internal/blackboard"
)

func TestSyntheticBlobEmitter_Suppression(t *testing.T) {
	bb := blackboard.New()

	// Start a tracker so it can post TrackedRegion
	tracker := NewSimpleBlobTracker()
	go tracker.Run(context.Background(), bb)

	// Give the tracker a moment to be ready
	time.Sleep(10 * time.Millisecond)

	emitter := NewSyntheticBlobEmitter()
	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		emitter.Run(ctx, bb)
		close(done)
	}()

	<-done

	// Count how many RedBlob were actually posted
	redBlobCount := 0
	for frame := uint64(1); frame <= 8; frame++ {
		results := bb.QueryByType("RedBlob", frame)
		redBlobCount += len(results)
	}

	// Because suppression should kick in after the first track is established,
	// we expect significantly fewer RedBlobs than frames
	if redBlobCount >= 7 {
		t.Errorf("expected suppression to reduce RedBlob count, got %d RedBlobs", redBlobCount)
	}
}
