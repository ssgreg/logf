package logf

import (
	"os"
	"runtime"
	"sync"
)

type ChannelWriterConfig struct {
	Capacity      int
	Appender      Appender
	ErrorAppender Appender
}

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
		l.sync()
	}
}

func (l *channelWriter) reportError(text string, err error) {
	skipError(l.ErrorAppender.Append(newErrorEntry(text, Error(err))))
	skipError(l.ErrorAppender.Sync())
}

func skipError(_ error) {
}
