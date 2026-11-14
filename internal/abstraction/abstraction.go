package abstraction

import "time"

// Abstraction is the basic unit that flows through the blackboard.
// All communication between Abstractors happens via Abstractions.
type Abstraction struct {
	Type      string    // e.g. "RedBlob", "TrackedBlob"
	FrameID   uint64    // monotonically increasing frame identifier
	Timestamp time.Time // when this Abstraction was created
	Payload   any       // numeric data, structs, etc. (keep simple in v0.1)
}

// String returns a short human-readable representation.
func (a Abstraction) String() string {
	return a.Type + " frame=" + string(rune(a.FrameID)) // simple for v0.1
}
