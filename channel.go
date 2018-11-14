package logf

import (
	"os"
	"runtime"
	"sync"
	"time"
)

// ChannelWriterConfig allows to configure ChannelWriter.
type ChannelWriterConfig struct {
	// Capacity specifies the underlying channel capacity.
	Capacity int

	// Appender specified the basic Appender for all Entries.
	//
	// Default Appender is WriterAppender with JSON Encoder.
	Appender Appender

	// ErrorAppender specifies the Appender for errors returning by basic
	// Appender.
	//
	// Default ErrorAppender does nothing.
	ErrorAppender Appender

	// EnableSyncOnError specifies whether Appender.Sync should be called
	// for messages with LevelError or not.
	//
	// Default value is false.
	EnableSyncOnError bool
}

// WithDefaults returns the new config in which all uninitialized fields are
// filled with their default values.
func (c ChannelWriterConfig) WithDefaults() ChannelWriterConfig {
	// Chan efficiency depends on the number of CPU installed in the system.
	// Tests shows that min chan capacity should be twice as big as CPU count.
	minCap := runtime.NumCPU() * 2
	if c.Capacity < minCap {
		c.Capacity = minCap
	}
	// No ErrorAppender by default.
	if c.ErrorAppender == nil {
		c.ErrorAppender = NewDiscardAppender()
	}
	// Default appender writes JSON-formatter messages to stdout.
	if c.Appender == nil {
		c.Appender = NewWriteAppender(os.Stdout, NewJSONEncoder.Default())
	}

	return c
}

// NewChannelWriter returns a new ChannelWriter with the given config.
var NewChannelWriter = channelWriterGetter(
	func(cfg ChannelWriterConfig) (EntryWriter, ChannelWriterCloseFunc) {
		l := &channelWriter{}
		l.init(cfg.WithDefaults())

		return l, ChannelWriterCloseFunc(
			func() {
				l.Close()
			})
	},
)

type ChannelWriterCloseFunc func()

type channelWriterGetter func(cfg ChannelWriterConfig) (EntryWriter, ChannelWriterCloseFunc)

func (c channelWriterGetter) Default() (EntryWriter, ChannelWriterCloseFunc) {
	return c(ChannelWriterConfig{})
}

type channelWriter struct {
	ChannelWriterConfig

	ch chan Entry
	wg sync.WaitGroup
}

func (l *channelWriter) Close() error {
	close(l.ch)
	l.wg.Wait()

	return nil
}

func (l *channelWriter) WriteEntry(e Entry) {
	l.ch <- e
}

func (l *channelWriter) Len() int {
	return len(l.ch)
}

func (l *channelWriter) Cap() int {
	return cap(l.ch)
}

func (l *channelWriter) init(cfg ChannelWriterConfig) {
	l.ChannelWriterConfig = cfg
	l.ch = make(chan Entry, l.Capacity)

	l.wg.Add(1)
	go l.worker()
}

func (l *channelWriter) worker() {
	defer l.wg.Done()

	var e Entry
	var ok bool
	for {
		select {
		case e, ok = <-l.ch:
		default:
			// Channel is empty. Force appender to Flush.
			l.flush()
			e, ok = <-l.ch
		}
		if !ok {
			break
		}

		l.append(e)
	}

	// Force appender to sync at exit.
	l.sync()
}

func (l *channelWriter) flush() {
	err := l.Appender.Flush()
	if err != nil {
		l.reportError("logf: failed to flush appender", err)
	}
}

func (l *channelWriter) sync() {
	err := l.Appender.Sync()
	if err != nil {
		l.reportError("logf: failed to sync appender", err)
	}
}

func (l *channelWriter) append(e Entry) {
	err := l.Appender.Append(e)
	if err != nil {
		l.reportError("logf: failed to append entry", err)
	}

	// Force appender to Sync if entry contains an error message.
	// This allows to commit buffered messages in case of futher unexpected
	// panic or crash.
	if e.Level <= LevelError {
		if l.EnableSyncOnError {
			l.sync()
		} else {
			l.flush()
		}
	}
}

func (l *channelWriter) reportError(text string, err error) {
	skipError(l.ErrorAppender.Append(newErrorEntry(text, Error(err))))
	skipError(l.ErrorAppender.Sync())
}

func skipError(_ error) {
}

func newErrorEntry(text string, fs ...Field) Entry {
	return Entry{
		LoggerID: -1,
		Level:    LevelError,
		Time:     time.Now(),
		Text:     text,
		Fields:   fs,
	}
}
