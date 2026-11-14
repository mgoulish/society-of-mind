package abstractors

import (
	"context"
	"time"

	"github.com/mgoulish/vision-society/internal/abstraction"
	"github.com/mgoulish/vision-society/internal/blackboard"
	"github.com/mgoulish/vision-society/internal/logger"
)

// SyntheticBlobEmitter generates synthetic RedBlob Abstractions.
type SyntheticBlobEmitter struct {
	name string
}

func NewSyntheticBlobEmitter() *SyntheticBlobEmitter {
	return &SyntheticBlobEmitter{name: "SyntheticBlobEmitter"}
}

func (e *SyntheticBlobEmitter) Name() string     { return e.name }
func (e *SyntheticBlobEmitter) Interests() []string { return nil }

func (e *SyntheticBlobEmitter) Run(ctx context.Context, bb *blackboard.Blackboard) {
	frame := uint64(0)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			frame++
			size := 5.0 + float64(frame)*0.4
			x := 100.0 + float64(frame)*0.1

			// Check for existing TrackedRegion and suppress if too close
			regions := bb.QueryByType("TrackedRegion", frame)
			if len(regions) == 0 && frame > 0 {
				regions = bb.QueryByType("TrackedRegion", frame-1)
			}

			suppressed := false
			for _, r := range regions {
				if payload, ok := r.Payload.(map[string]any); ok {
					trackedX, _ := payload["x"].(float64)
					trackedSize, _ := payload["size"].(float64)

					distance := abs(x - trackedX)
					threshold := trackedSize * 0.7
					if distance < threshold && trackedSize > 0 {
						logger.Log(logger.Event{
							Abstractor: e.name,
							Event:      "suppressed",
							FrameID:    frame,
							Details: map[string]any{
								"reason":    "already_tracked",
								"track_x":   trackedX,
								"my_x":      x,
								"distance":  distance,
								"threshold": threshold,
							},
						})
						suppressed = true
						break
					}
				}
			}

			if suppressed {
				continue
			}

			ab := abstraction.Abstraction{
				Type:      "RedBlob",
				FrameID:   frame,
				Timestamp: time.Now(),
				Payload: map[string]any{
					"size": size,
					"x":    x,
					"y":    120.0,
				},
			}
			bb.Post(ab)

			logger.Log(logger.Event{
				Abstractor: e.name,
				Event:      "posted",
				FrameID:    frame,
				Details: map[string]any{
					"type": "RedBlob",
					"size": size,
					"x":    x,
				},
			})
		}
	}
}

// abs is a small helper for the suppression distance check
func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
