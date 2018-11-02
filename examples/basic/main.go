package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/ssgreg/logf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogger(l logf.Level, w io.Writer) (*logf.Logger, logf.ChannelWriterCloseFunc) {
	ew, close := logf.NewChannelWriter.Default()

	return logf.NewLogger(l, ew), close
}

var messages = makePseudoMessages(1000)

func makePseudoMessages(n int) []string {
	messages := make([]string, n)
	for i := range messages {
		messages[i] = fmt.Sprintf("A text that pretend to be a real message in case of length %d", i)
	}
	return messages
}

func getMessage(n int) string {
	return messages[n%len(messages)]
}

func newZapLogger(lvl zapcore.Level) *zap.Logger {
	c := zap.NewProductionConfig()
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	c.DisableCaller = true
	c.OutputPaths = []string{"stdout"}
	c.Sampling = nil

	logger, _ := c.Build()
	return logger

	// ec := zap.NewProductionEncoderConfig()
	// // ec.EncodeDuration = zapcore.NanosDurationEncoder
	// // ec.EncodeTime = zapcore.EpochNanosTimeEncoder
	// ec.EncodeTime = zapcore.ISO8601TimeEncoder
	// enc := zapcore.NewJSONEncoder(ec)
	// return zap.New(zapcore.NewCore(
	// 	enc,
	// 	&Discarder{},
	// 	lvl,
	// ))
}

// A Syncer is a spy for the Sync portion of zapcore.WriteSyncer.
type Syncer struct {
	err    error
	called bool
}

// SetError sets the error that the Sync method will return.
func (s *Syncer) SetError(err error) {
	s.err = err
}

// Sync records that it was called, then returns the user-supplied error (if
// any).
func (s *Syncer) Sync() error {
	s.called = true
	return s.err
}

// Called reports whether the Sync method was called.
func (s *Syncer) Called() bool {
	return s.called
}

// A Discarder sends all writes to ioutil.Discard.
type Discarder struct{ Syncer }

// Write implements io.Writer.
func (d *Discarder) Write(b []byte) (int, error) {
	// return 0, errors.New("ficj")
	return ioutil.Discard.Write(b)
	// fmt.Println(string(b))
	// panic(33)
	// return 0, nil
}

func main() {
	// tmp := make([]byte, 0, 1024*1024*200)
	logger, close := newLogger(logf.LevelDebug, os.Stdout) // bytes.NewBuffer(tmp))
	defer close()

	// logger := newZapLogger(zap.DebugLevel)
	// defer logger.Sync()

	wg := sync.WaitGroup{}
	wg.Add(1)

	for g := 0; g < 1; g++ {
		go func() {
			defer wg.Done()

			for i := 0; i < 1; i++ {
				logger.Info(getMessage(i), logf.Bytes("test", []byte(`1), logf.Int("test", 1), logf.Int("test", 1)`)))
				time.Sleep(time.Second)
			}
		}()
	}

	wg.Wait()
}
