// WriterSlot example: lazy destination initialization.
//
// Start logging immediately, connect the real destination later
// (e.g., after config parsing or service discovery).
package main

import (
	"context"
	"os"

	"github.com/ssgreg/logf/v2"
)

func main() {
	// Create a slot with pre-Set buffering (4 KB).
	// Writes before Set are buffered, not dropped.
	slot := logf.NewWriterSlot(logf.WithSlotBuffer(4096))

	// Logger works immediately — writes go to buffer.
	logger := logf.NewLogger().Output(slot).Build()
	ctx := context.Background()

	logger.Info(ctx, "application starting")
	logger.Info(ctx, "loading config")

	// ... later, after config is parsed and destination is known:
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Connect the real destination.
	// Buffered data is flushed on the first Write after Set.
	slot.Set(file)

	// From now on, writes go directly to file.
	logger.Info(ctx, "config loaded, logging to file")
	logger.Info(ctx, "ready to serve")
}
