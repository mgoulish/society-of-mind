package scenarios

import (
	"context"

	"github.com/mgoulish/vision-society/internal/blackboard"
	"github.com/mgoulish/vision-society/internal/logger"
)

// Emitter runs a Scenario and posts its Abstractions to the blackboard.
type Emitter struct {
	name     string
	scenario Scenario
}

func NewEmitter(name string, scenario Scenario) *Emitter {
	return &Emitter{
		name:     name,
		scenario: scenario,
	}
}

func (e *Emitter) Name() string { return e.name }

func (e *Emitter) Run(ctx context.Context, bb *blackboard.Blackboard) {
	ch := e.scenario(ctx)

	for ab := range ch {
		bb.Post(ab)

		logger.Log(logger.Event{
			Abstractor: e.name,
			Event:      "posted",
			FrameID:    ab.FrameID,
			Details: map[string]any{
				"type": ab.Type,
			},
		})
	}
}
