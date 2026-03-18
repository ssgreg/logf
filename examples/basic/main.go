package main

import (
	"context"
	"runtime"

	"github.com/ssgreg/logf/v2"
)

func main() {
	logger := logf.NewLogger().Build()

	logger.Info(context.Background(), "got cpu info", logf.Int("count", runtime.NumCPU()))
}
