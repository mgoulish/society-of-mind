package main

import (
    "fmt"
    "image"
    _ "image/png" // Register PNG decoder
    "os"
    "path/filepath"
    "sync"
    "time"
)

// Agent represents a runtime entity that can produce data and allow others to subscribe to its outputs.
type Agent struct {
    Name        string
    run         func(*Agent) // Custom run behavior (different per instance)

    mu          sync.Mutex
    subscribers []chan interface{} // List of channels to broadcast outputs to
}

// Start launches the agent's run function in a goroutine.
func (a *Agent) Start() {
    go a.run(a)
}

// Subscribe allows another agent (or entity) to register a channel for receiving outputs.
func (a *Agent) Subscribe(ch chan interface{}) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.subscribers = append(a.subscribers, ch)
}

// Publish sends data to all subscribed channels (non-blocking; assumes channels are buffered if needed).
func (a *Agent) Publish(data interface{}) {
    a.mu.Lock()
    defer a.mu.Unlock()
    for _, ch := range a.subscribers {
        ch <- data // Send to each subscriber
    }
}

// Registry for discovering agents at runtime (thread-safe).
var registry sync.Map // map[string]*Agent

// RegisterAgent adds an agent to the registry by name.
func RegisterAgent(a *Agent) {
    registry.Store(a.Name, a)
    fmt.Print ( "RegisterAgent registered ", a.Name, "\n" )
}

// FindAgent looks up an agent by name.
func FindAgent(name string) *Agent {
    val, ok := registry.Load(name)
    if !ok {
        return nil
    }
    return val.(*Agent)
}

// Example: Create a simple image-reading agent.
// This agent's Run reads images from a directory (e.g., ./images/0001.png, 0002.png, etc.),
// decodes them, and publishes the image.Image to subscribers.
// It runs indefinitely, checking for new files every second if it reaches the end.
func NewImageReaderAgent(dir string) *Agent {
    return &Agent{
        Name: "ImageReader",
        run: func(self *Agent) {
            i := 1
            for {
                fileName := fmt.Sprintf("%04d.png", i)
                path := filepath.Join(dir, fileName)
                file, err := os.Open(path)
                if err != nil {
                    // If file doesn't exist, wait and retry (simulates waiting for new images)
                    time.Sleep(time.Second)
                    continue
                }
                img, _, err := image.Decode(file)
                file.Close()
                if err != nil {
                    fmt.Printf("Error decoding %s: %v\n", path, err)
                    i++
                    continue
                }
                self.Publish(img) // Send the image to subscribers
		fmt.Print ( "Published ", fileName, "\n" )
                i++
                //time.Sleep(50 * time.Millisecond) // Simulate ~20/sec; adjust as needed
                time.Sleep(time.Second) 
            }
        },
    }
}

func main() {
    // Assume images are in ./images directory
    reader := NewImageReaderAgent("./images")
    RegisterAgent(reader)
    reader.Start()

    // For now, no subscribers, so it just runs and publishes to nowhere.
    // In a real system, other agents would find it via FindAgent("ImageReader") and call Subscribe(theirChan).

    // Keep the program running indefinitely
    select {}
}
