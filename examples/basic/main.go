package main

import (
	"context"
	"runtime"

	"github.com/ssgreg/logf/v2"
)

func main() {
	// The default channel writer writes to stdout using json encoder.
	writer, writerClose := logf.NewChannelWriter.Default()
	defer writerClose()

	logger := logf.NewLogger(logf.LevelInfo, writer)

	logger.Info(context.Background(), "got cpu info", logf.Int("count", runtime.NumCPU()))
}
