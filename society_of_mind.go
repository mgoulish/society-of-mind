package main

import (
    "fmt"
    "image"
    _ "image/png" // Register PNG decoder
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "sync"
    "time"
)

// Agent represents a runtime entity that can produce data and allow others to subscribe to its outputs.
type Agent struct {
    Name string
    run  func(*Agent) // Custom run behavior (different per instance)

    mu          sync.Mutex
    subscribers []chan interface{} // List of channels to broadcast outputs to
}

// Log prints a timestamped message for the agent (now with microseconds).
func (a *Agent) Log(msg string) {
    fmt.Printf("[%s] %s: %s\n", time.Now().Format("2006-01-02T15:04:05.000000"), a.Name, msg)
}

// Start launches the agent's run function in a goroutine.
func (a *Agent) Start() {
    a.Log("Starting up")
    go a.run(a)
}

// Subscribe allows another agent (or entity) to register a channel for receiving outputs.
func (a *Agent) Subscribe(ch chan interface{}) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.subscribers = append(a.subscribers, ch)
    a.Log(fmt.Sprintf("New subscriber registered (total: %d)", len(a.subscribers)))
}

// Publish sends data to all subscribed channels (non-blocking; assumes channels are buffered if needed).
func (a *Agent) Publish(data interface{}) {
    a.mu.Lock()
    defer a.mu.Unlock()
    for _, ch := range a.subscribers {
        select {
        case ch <- data:
            // Sent successfully
        default:
            a.Log("Warning: Dropped message to subscriber (channel full)")
        }
    }
}

// Registry for discovering agents at runtime (thread-safe).
var registry sync.Map // map[string]*Agent

// RegisterAgent adds an agent to the registry by name.
func RegisterAgent(a *Agent) {
    registry.Store(a.Name, a)
    a.Log("Registered in central registry")
}

// FindAgent looks up an agent by name.
func FindAgent(name string) *Agent {
    val, ok := registry.Load(name)
    if !ok {
        return nil
    }
    return val.(*Agent)
}

// NewImageReaderAgent creates an agent that reads images from a directory in a loop.
func NewImageReaderAgent(dir string) *Agent {
    return &Agent{
        Name: "ImageReader",
        run: func(self *Agent) {
            // Scan directory to find the maximum image number
            maxI := 0
            files, err := filepath.Glob(filepath.Join(dir, "*.png"))
            if err != nil {
                self.Log(fmt.Sprintf("Error scanning directory %s: %v", dir, err))
                return
            }
            for _, f := range files {
                name := filepath.Base(f)
                name = strings.TrimSuffix(name, ".png")
                if num, err := strconv.Atoi(name); err == nil && num > maxI {
                    maxI = num
                }
            }
            if maxI == 0 {
                self.Log("No PNG images found in directory")
                return
            }
            self.Log(fmt.Sprintf("Found %d images; starting loop", maxI))

            i := 1
            for {
                fileName := fmt.Sprintf("%04d.png", i)
                path := filepath.Join(dir, fileName)
                file, err := os.Open(path)
                if err != nil {
                    self.Log(fmt.Sprintf("Error opening %s: %v (this shouldn't happen in loop)", path, err))
                    i++
                    if i > maxI {
                        i = 1
                    }
                    continue
                }
                img, _, err := image.Decode(file)
                file.Close()
                if err != nil {
                    self.Log(fmt.Sprintf("Error decoding %s: %v", path, err))
                    i++
                    if i > maxI {
                        i = 1
                    }
                    continue
                }
                self.Publish(img) // Send the image to subscribers
                self.Log(fmt.Sprintf("Publishing image %s", fileName))
                i++
                if i > maxI {
                    i = 1
                }
                //time.Sleep(50 * time.Millisecond) // ~20 images per second
                time.Sleep(time.Second) // ~20 images per second
            }
        },
    }
}

// NewRoadFollowerAgent creates an agent that subscribes to ImageReader and receives images.
func NewRoadFollowerAgent() *Agent {
    return &Agent{
        Name: "RoadFollower",
        run: func(self *Agent) {
            self.Log("Looking for ImageReader to subscribe")
            producer := FindAgent("ImageReader")
            if producer == nil {
                self.Log("Error: ImageReader not found in registry")
                return
            }

            ch := make(chan interface{}, 100) // Buffered to handle ~20/sec without blocking
            producer.Subscribe(ch)
            self.Log("Subscribed to ImageReader; waiting for images")

            for data := range ch { // Receives indefinitely (channel never closed)
                // Use 'data' to log image details (fixes unused variable)
                if img, ok := data.(image.Image); ok {
                    bounds := img.Bounds()
                    self.Log(fmt.Sprintf("Received image of size %dx%d", bounds.Dx(), bounds.Dy()))
                } else {
                    self.Log("Received non-image data")
                }
            }
        },
    }
}

func main() {
    // Assume images are in ./images directory (create it with 0001.png, etc., for testing)
    reader := NewImageReaderAgent("./images")
    RegisterAgent(reader)
    reader.Start()

    // Add the new RoadFollower agent
    follower := NewRoadFollowerAgent()
    RegisterAgent(follower)
    follower.Start()

    // Keep the program running indefinitely
    select {}
}
