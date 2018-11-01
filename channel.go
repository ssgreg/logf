package logf

import (
	"fmt"
	"runtime"
	"sync"
)

type ChannelWriter interface {
	Write(Entry)
}

type ChannelCloser interface {
	Close()
}

type Channel interface {
	ChannelWriter
	ChannelCloser
}

type ChannelConfig struct {
	Capacity      int
	Appender      Appender
	ErrorAppender Appender
}

func NewBasicChannel(cfg ChannelConfig) Channel {
	minCap := runtime.NumCPU() * 2
	if cfg.Capacity < minCap {
		cfg.Capacity = minCap
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	l := &basicChannel{
		channel:       make(chan Entry, cfg.Capacity),
		wg:            &wg,
		ChannelConfig: cfg,
	}
	go l.worker()

	return l
}

type basicChannel struct {
	ChannelConfig
	channel chan Entry
	wg      *sync.WaitGroup
}

func (l *basicChannel) Close() {
	close(l.channel)
	l.wg.Wait()
}

// Log TODO
func (l *basicChannel) Write(e Entry) {
	l.channel <- e
}

func (l *basicChannel) Len() int {
	return len(l.channel)
}

func (l *basicChannel) worker() {
	defer l.wg.Done()

	var e Entry
	var ok bool
	for {
		select {
		case e, ok = <-l.channel:
		default:
			// Channel is empty. Force appender to Flush.
			l.flush()
			e, ok = <-l.channel
		}
		if !ok {
			break
		}

		l.append(e)
	}

	// Force appender to sync at exit.
	l.sync()
	l.close()
}

func (l *basicChannel) flush() {
	err := l.Appender.Flush()
	if err != nil {
		l.reportError(fmt.Sprintf("logf: failed to flush appender: %+v", err))
	}
}

func (l *basicChannel) sync() {
	err := l.Appender.Sync()
	if err != nil {
		l.reportError(fmt.Sprintf("logf: failed to sync appender: %+v", err))
	}
}

func (l *basicChannel) append(e Entry) {
	err := l.Appender.Append(e)
	if err != nil {
		l.reportError(fmt.Sprintf("logf: failed to append entry to appender: %+v", err))
	}

	if e.Level <= LevelError {
		// The entry contains error message. Force appender to Sync.
		l.sync()
	}
}

func (l *basicChannel) close() {
	err := l.Appender.Close()
	if err != nil {
		l.reportError(fmt.Sprintf("logf: failed to close appender: %+v", err))
	}
	if l.ErrorAppender != nil {
		_ = l.ErrorAppender.Close()
	}
}

func (l *basicChannel) reportError(text string) {
	if l.ErrorAppender != nil {
		// TODO: pass error as field value
		_ = l.ErrorAppender.Append(NewErrorEntry(text))
		_ = l.ErrorAppender.Sync()
	}
}
