package main

import (
    "fmt"
    "image"
    "image/draw"
    _ "image/png" // Register PNG decoder
    "os"
    "path/filepath"
    "reflect"
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
    inputs      []chan interface{} // Multiple input channels
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

// SubscribeTo allows this agent to subscribe to a producer, adding a new input channel.
func (a *Agent) SubscribeTo(producer *Agent) chan interface{} {
    ch := make(chan interface{}, 100) // Buffered
    producer.Subscribe(ch)
    a.mu.Lock()
    defer a.mu.Unlock()
    a.inputs = append(a.inputs, ch)
    a.Log(fmt.Sprintf("Subscribed to %s (total inputs: %d)", producer.Name, len(a.inputs)))
    return ch
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

// NewImageReaderAgent creates an agent that reads images from ./simulated_data/ImageReader/ in a loop.
func NewImageReaderAgent() *Agent {
    dir := "./simulated_data/ImageReader"
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



// copyData deep-copies data based on type for stability.
func copyData(item interface{}) interface{} {
    switch v := item.(type) {
    case image.Image:
        bounds := v.Bounds()
        copyImg := image.NewRGBA(bounds)
        draw.Draw(copyImg, bounds, v, bounds.Min, draw.Src)
        return copyImg
    default:
        return v // As-is for other types
    }
}



// NewRoadFollowerAgent (updated with copying)
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
            self.SubscribeTo(producer)
            self.Log("Subscribed to ImageReader; waiting for inputs")

            // Scan own simulated output directory
            outputDir := "./simulated_data/RoadFollower"
            maxI := 0
            files, err := filepath.Glob(filepath.Join(outputDir, "*.png"))
            if err != nil {
                self.Log(fmt.Sprintf("Error scanning output directory %s: %v", outputDir, err))
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
                self.Log("No PNG output images found in simulated directory")
                return
            }
            self.Log(fmt.Sprintf("Found %d simulated output images; ready to process", maxI))

            i := 1 // Start with first simulated output
            for {
                // Block until at least one input is ready (even for single input)
                self.mu.Lock()
                inputChannels := append([]chan interface{}(nil), self.inputs...) // Copy to avoid lock during reflect
                self.mu.Unlock()
                if len(inputChannels) == 0 {
                    self.Log("No inputs available; exiting run loop")
                    return
                }

                cases := make([]reflect.SelectCase, len(inputChannels))
                for idx, ch := range inputChannels {
                    cases[idx] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
                }
                chosen, value, ok := reflect.Select(cases)
                if !ok {
                    self.Log("Input channel closed; removing it")
                    self.mu.Lock()
                    self.inputs = append(self.inputs[:chosen], self.inputs[chosen+1:]...)
                    self.mu.Unlock()
                    continue
                }
                data := value.Interface()

                // Process the initial item
                batch := []interface{}{data}
                self.Log("Received initial input from channel")

                // Drain other channels non-blockingly (no-op for single)
                for idx, ch := range inputChannels {
                    if idx == chosen {
                        continue
                    }
                    for {
                        select {
                        case extra, ok := <-ch:
                            if !ok {
                                self.Log("Input channel closed during drain")
                                self.mu.Lock()
                                for k, inputCh := range self.inputs {
                                    if inputCh == ch {
                                        self.inputs = append(self.inputs[:k], self.inputs[k+1:]...)
                                        break
                                    }
                                }
                                self.mu.Unlock()
                                continue
                            }
                            batch = append(batch, extra)
                            self.Log("Drained additional input")
                        default:
                            // No more available
                        }
                    }
                }

                // Copy the batch for stable processing
                copiedBatch := make([]interface{}, len(batch))
                for j, item := range batch {
                    copiedBatch[j] = copyData(item)
                }
                self.Log(fmt.Sprintf("Copied %d input items for processing", len(copiedBatch)))

                // Process copied batch
                for _, item := range copiedBatch {
                    if img, ok := item.(image.Image); ok {
                        bounds := img.Bounds()
                        self.Log(fmt.Sprintf("Processing copied input image of size %dx%d", bounds.Dx(), bounds.Dy()))

                        // Load next simulated output image
                        outputFileName := fmt.Sprintf("%04d.png", i)
                        outputPath := filepath.Join(outputDir, outputFileName)
                        outputFile, err := os.Open(outputPath)
                        if err != nil {
                            self.Log(fmt.Sprintf("Error opening simulated output %s: %v", outputPath, err))
                            i = (i % maxI) + 1
                            continue
                        }
                        outputImg, _, err := image.Decode(outputFile)
                        outputFile.Close()
                        if err != nil {
                            self.Log(fmt.Sprintf("Error decoding simulated output %s: %v", outputPath, err))
                            i = (i % maxI) + 1
                            continue
                        }

                        // Publish the simulated output (no need to copy here unless subscribers mutate)
                        self.Publish(outputImg)
                        self.Log(fmt.Sprintf("Processed input to simulated output image %s", outputFileName))

                        i = (i % maxI) + 1
                    } else {
                        self.Log("Received non-image input")
                    }
                }
            }
        },
    }
}



func main() {
    // Agents use ./simulated_data/{AGENT_NAME}/ for input/output simulation
    reader := NewImageReaderAgent()
    RegisterAgent(reader)
    reader.Start()

    // Add the RoadFollower agent (now using multi-input mechanism for single input)
    follower := NewRoadFollowerAgent()
    RegisterAgent(follower)
    follower.Start()

    // Keep the program running indefinitely
    select {}
}

