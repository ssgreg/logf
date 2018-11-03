# logf

[![GoDoc](https://godoc.org/github.com/ssgreg/logf?status.svg)](https://godoc.org/github.com/ssgreg/logf)
[![Build Status](https://travis-ci.org/ssgreg/logf.svg?branch=master)](https://travis-ci.org/ssgreg/logf)
[![Go Report Status](https://goreportcard.com/badge/github.com/ssgreg/logf)](https://goreportcard.com/report/github.com/ssgreg/logf)
[![Coverage Status](https://coveralls.io/repos/github/ssgreg/logf/badge.svg?branch=master)](https://coveralls.io/github/ssgreg/logf?branch=master)

Faster-than-light, asynchronous, structured logger in Go with zero allocation count.

## Example

The following example creates the new `logf` logger and logs a message.

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
