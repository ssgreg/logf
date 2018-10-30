package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/ssgreg/logf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogger(l logf.Level, w io.Writer) (*logf.Logger, logf.Channel) {
	encoder := logf.NewJSONEncoder(&logf.FormatterConfig{
		FormatTime: logf.RFC3339TimeFormatter,
	})

	channel := logf.NewBasicChannel(logf.ChannelConfig{
		Appender:      logf.NewWriteAppender(w, encoder),
		ErrorAppender: logf.NewWriteAppender(os.Stderr, encoder),
	})

	return logf.NewLogger(logf.NewMutableLevel(l).Checker(), channel), channel
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

	logger, channel := newLogger(logf.LevelDebug, os.Stdout) // bytes.NewBuffer(tmp))
	defer channel.Close()

	// logger := newZapLogger(zap.DebugLevel)
	// defer logger.Sync()

	wg := sync.WaitGroup{}
	wg.Add(1000)

	for g := 0; g < 1000; g++ {
		go func() {
			defer wg.Done()

			for i := 0; i < 1000; i++ {
				logger.Info(getMessage(i), logf.Int("test", 1), logf.Int("test", 1), logf.Int("test", 1))
				// time.Sleep(time.Millisecond)
			}
		}()
	}

	wg.Wait()
}
