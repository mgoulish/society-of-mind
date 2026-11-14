package scenarios

import (
	"context"

	"github.com/mgoulish/vision-society/internal/abstraction"
)

// Scenario is a function that produces a stream of Abstractions.
// It is responsible for closing the channel when done or when context is cancelled.
type Scenario func(ctx context.Context) <-chan abstraction.Abstraction

// GrowingRedBlob returns a scenario that emits RedBlob Abstractions
// representing a blob that steadily grows and moves slightly as frames progress.
func GrowingRedBlob() Scenario {
	return func(ctx context.Context) <-chan abstraction.Abstraction {
		ch := make(chan abstraction.Abstraction, 8)

		go func() {
			defer close(ch)
			frame := uint64(0)

			for {
				select {
				case <-ctx.Done():
					return
				default:
					frame++
					size := 5.0 + float64(frame)*0.4
					x := 100.0 + float64(frame)*0.1

					ab := abstraction.Abstraction{
						Type:    "RedBlob",
						FrameID: frame,
						Payload: map[string]any{
							"size": size,
							"x":    x,
							"y":    120.0,
						},
					}
					ch <- ab
				}
			}
		}()

		return ch
	}
}
