package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Event is the minimal structured log entry for v0.1.
type Event struct {
	Time        time.Time      `json:"time"`
	Abstractor  string         `json:"abstractor"`
	Event       string         `json:"event"` // "received", "fired", "posted", etc.
	FrameID     uint64         `json:"frame_id,omitempty"`
	Details     map[string]any `json:"details,omitempty"`
}

// Log writes a single JSON line to stdout.
// In later versions we can make this write to per-Abstractor files.
func Log(ev Event) {
	ev.Time = time.Now().UTC()
	b, err := json.Marshal(ev)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger error: %v\n", err)
		return
	}
	fmt.Println(string(b))
}
