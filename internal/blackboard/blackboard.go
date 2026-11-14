package blackboard

import (
	"sync"

	"github.com/mgoulish/vision-society/internal/abstraction"
)

// Blackboard is the shared workspace where Abstractors post and discover Abstractions.
type Blackboard struct {
	mu          sync.RWMutex
	store       map[string][]abstraction.Abstraction // type -> recent Abstractions
	subscribers map[string][]chan abstraction.Abstraction
}

// New creates an empty Blackboard.
func New() *Blackboard {
	return &Blackboard{
		store:       make(map[string][]abstraction.Abstraction),
		subscribers: make(map[string][]chan abstraction.Abstraction),
	}
}

// Post adds an Abstraction to the blackboard and notifies all interested subscribers.
// Notifications are sent non-blocking so a slow subscriber cannot stall the system.
func (bb *Blackboard) Post(ab abstraction.Abstraction) {
	bb.mu.Lock()
	bb.store[ab.Type] = append(bb.store[ab.Type], ab)

	// collect channels while holding lock
	chs := make([]chan abstraction.Abstraction, len(bb.subscribers[ab.Type]))
	copy(chs, bb.subscribers[ab.Type])
	bb.mu.Unlock()

	// notify outside the lock
	for _, ch := range chs {
		select {
		case ch <- ab:
		default:
			// drop if subscriber is slow (v0.1 simple policy)
		}
	}
}

// Subscribe returns a channel that will receive new Abstractions whose Type
// is in the interests list. The caller is responsible for reading from the channel.
func (bb *Blackboard) Subscribe(interests []string) <-chan abstraction.Abstraction {
	ch := make(chan abstraction.Abstraction, 8) // small buffer

	bb.mu.Lock()
	for _, t := range interests {
		bb.subscribers[t] = append(bb.subscribers[t], ch)
	}
	bb.mu.Unlock()

	return ch
}

// QueryByType returns recent Abstractions of the given type for a specific frame.
// In v0.1 we keep this extremely simple (linear scan of the last few entries).
func (bb *Blackboard) QueryByType(typ string, frameID uint64) []abstraction.Abstraction {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	var result []abstraction.Abstraction
	for _, ab := range bb.store[typ] {
		if ab.FrameID == frameID {
			result = append(result, ab)
		}
	}
	return result
}
