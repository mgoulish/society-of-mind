package replay

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/mgoulish/vision-society/internal/abstraction"
)

// Event is a minimal copy of the logger.Event for parsing logs.
// In a more advanced version we would share the type.
type Event struct {
	Abstractor string         `json:"abstractor"`
	Event      string         `json:"event"`
	FrameID    uint64         `json:"frame_id,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

// NewAbstractionChannelFromLog creates a channel that emits Abstractions
// reconstructed from a log file for a specific Abstractor.
// This is a basic first version focused on replaying "received" events.
func NewAbstractionChannelFromLog(filename string, abstractorName string) (<-chan abstraction.Abstraction, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	ch := make(chan abstraction.Abstraction, 16)

	go func() {
		defer file.Close()
		defer close(ch)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Bytes()

			var ev Event
			if err := json.Unmarshal(line, &ev); err != nil {
				continue // skip bad lines
			}

			if ev.Abstractor != abstractorName || ev.Event != "received" {
				continue
			}

			// Reconstruct a minimal Abstraction from the logged details.
			// This is limited in v1 of replay — we only pull out what was logged.
			ab := abstraction.Abstraction{
				Type:    getString(ev.Details, "type"),
				FrameID: ev.FrameID,
			}

			// Try to reconstruct payload if present in details (future enhancement)
			if payloadData, ok := ev.Details["payload"]; ok {
				if p, ok := payloadData.(map[string]any); ok {
					ab.Payload = p
				}
			}

			ch <- ab
		}
	}()

	return ch, nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
