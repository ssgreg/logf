# logf

[![GoDoc](https://godoc.org/github.com/ssgreg/logf?status.svg)](https://godoc.org/github.com/ssgreg/logf)
[![Build Status](https://travis-ci.org/ssgreg/logf.svg?branch=master)](https://travis-ci.org/ssgreg/logf)
[![Go Report Status](https://goreportcard.com/badge/github.com/ssgreg/logf)](https://goreportcard.com/report/github.com/ssgreg/logf)
[![GoCover](https://gocover.io/_badge/github.com/ssgreg/logf)](https://gocover.io/github.com/ssgreg/logf)

Faster-than-light, asynchronous, structured logger in Go with zero allocation count.

## Example

The following example creates a new `logf` logger and logs a message.

```go
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
```

The output is the following:

```json
{"level":"info","ts":"2018-11-03T09:49:56+03:00","msg":"got cpu info","count":8}
```

## Benchmarks

TODO