package abstractors

import (
	"context"
	"time"

	"github.com/mgoulish/vision-society/internal/abstraction"
	"github.com/mgoulish/vision-society/internal/blackboard"
	"github.com/mgoulish/vision-society/internal/logger"
)

// SimpleBlobTracker subscribes to RedBlob and produces TrackedBlob.
type SimpleBlobTracker struct {
	name           string
	trackID        uint64
	lastSize       float64
	lastX          float64
	lastPostedSize float64
	lastPostedX    float64
	input          <-chan abstraction.Abstraction // if set, use this instead of subscribing
}

func NewSimpleBlobTracker() *SimpleBlobTracker {
	return &SimpleBlobTracker{
		name:    "SimpleBlobTracker",
		trackID: 42,
	}
}

// NewSimpleBlobTrackerFromChannel creates a Tracker that reads from the given channel
// instead of subscribing to a blackboard. Useful for replay and testing.
func NewSimpleBlobTrackerFromChannel(input <-chan abstraction.Abstraction) *SimpleBlobTracker {
	return &SimpleBlobTracker{
		name:    "SimpleBlobTracker",
		trackID: 42,
		input:   input,
	}
}

func (t *SimpleBlobTracker) Name() string     { return t.name }
func (t *SimpleBlobTracker) Interests() []string { return []string{"RedBlob"} }

func (t *SimpleBlobTracker) Run(ctx context.Context, bb *blackboard.Blackboard) {
	var ch <-chan abstraction.Abstraction
	if t.input != nil {
		ch = t.input
	} else {
		ch = bb.Subscribe(t.Interests())
	}

	for {
		select {
		case <-ctx.Done():
			return
		case ab := <-ch:
			// update simple state
			if payload, ok := ab.Payload.(map[string]any); ok {
				if size, ok := payload["size"].(float64); ok {
					t.lastSize = size
				}
				if x, ok := payload["x"].(float64); ok {
					t.lastX = x
				}
			}

			// Build richer details for logging (helps replay + debugging)
			receivedDetails := map[string]any{"type": ab.Type}
			if payload, ok := ab.Payload.(map[string]any); ok {
				if size, ok := payload["size"].(float64); ok {
					receivedDetails["size"] = size
				}
				if x, ok := payload["x"].(float64); ok {
					receivedDetails["x"] = x
				}
			}

			logger.Log(logger.Event{
				Abstractor: t.name,
				Event:      "received",
				FrameID:    ab.FrameID,
				Details:    receivedDetails,
			})

			// Only fire if there is a meaningful change (basic temporal consistency)
			sizeChanged := abs(t.lastSize-t.lastPostedSize) > (t.lastPostedSize * 0.05) // >5% change
			posChanged := abs(t.lastX-t.lastPostedX) > 1.0

			if sizeChanged || posChanged || t.lastPostedSize == 0 {
				// Post TrackedBlob
				tracked := abstraction.Abstraction{
					Type:      "TrackedBlob",
					FrameID:   ab.FrameID,
					Timestamp: time.Now(),
					Payload: map[string]any{
						"track_id": t.trackID,
						"size":     t.lastSize,
					},
				}
				bb.Post(tracked)

				logger.Log(logger.Event{
					Abstractor: t.name,
					Event:      "fired",
					FrameID:    ab.FrameID,
					Details: map[string]any{
						"posted_type": "TrackedBlob",
						"track_id":    t.trackID,
						"size":        t.lastSize,
					},
				})

				// Post TrackedRegion with spatial info for suppression
				region := abstraction.Abstraction{
					Type:      "TrackedRegion",
					FrameID:   ab.FrameID,
					Timestamp: time.Now(),
					Payload: map[string]any{
						"track_id": t.trackID,
						"x":        t.lastX,
						"size":     t.lastSize,
					},
				}
				bb.Post(region)

				// Remember what we last posted
				t.lastPostedSize = t.lastSize
				t.lastPostedX = t.lastX
			}
		}
	}
}
