package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mgoulish/vision-society/internal/abstractor"
	"github.com/mgoulish/vision-society/internal/abstractors"
	"github.com/mgoulish/vision-society/internal/blackboard"
)

func main() {
	fmt.Println("=== Vision Society v0.1 starting ===")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// graceful shutdown on Ctrl-C or SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutdown signal received...")
		cancel()
	}()

	bb := blackboard.New()

	emitter := abstractors.NewSyntheticBlobEmitter()
	tracker := abstractors.NewSimpleBlobTracker()

	abstractor.RunAll(ctx, bb, emitter, tracker)

	// Run for a fixed number of frames (emitter produces ~10 per second)
	time.Sleep(3 * time.Second)
	cancel()

	// give goroutines a moment to exit
	time.Sleep(200 * time.Millisecond)

	fmt.Println("=== Vision Society v0.1 finished ===")
}
