package abstractor

import (
	"context"

	"github.com/mgoulish/vision-society/internal/blackboard"
)

// Abstractor represents one autonomous agent in the society.
// Each Abstractor runs in its own goroutine.
type Abstractor interface {
	Name() string
	Interests() []string
	Run(ctx context.Context, bb *blackboard.Blackboard)
}

// RunAll starts all Abstractors and blocks until the context is cancelled.
// It is a convenience helper for v0.1.
func RunAll(ctx context.Context, bb *blackboard.Blackboard, abs ...Abstractor) {
	for _, a := range abs {
		go a.Run(ctx, bb)
	}
}
