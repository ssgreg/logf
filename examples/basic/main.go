package main

import (
	"runtime"

	"github.com/ssgreg/logf"
)

func main() {
	// The default channel writer writes to stdout using json encoder.
	writer, writerClose := logf.NewChannelWriter.Default()
	defer writerClose()

	logger := logf.NewLogger(logf.LevelInfo, writer)

	logger.Info("got cpu info", logf.Int("count", runtime.NumCPU()))
}
