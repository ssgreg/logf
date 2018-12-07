package logf

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChannelWriterConfigDefaults(t *testing.T) {
	cfg := ChannelWriterConfig{}
	cfgWithDefaults := cfg.WithDefaults()

	// Capacity, ErrorAppender, Appender must be configured by WithDefaults.
	assert.True(t, cfgWithDefaults.Capacity > 0)
	assert.NotNil(t, cfgWithDefaults.ErrorAppender)
	assert.NotNil(t, cfgWithDefaults.Appender)

	// EnableSyncOnError must not be changed.
	assert.Equal(t, cfg.EnableSyncOnError, cfgWithDefaults.EnableSyncOnError)
}

func TestChannelWriterNewDefault(t *testing.T) {
	w, close := NewChannelWriter.Default()
	defer close()

	cw := w.(*channelWriter)
	// Capacity, ErrorAppender, Appender must be configured by WithDefaults.
	assert.True(t, cw.Capacity > 0)
	assert.NotNil(t, cw.ErrorAppender)
	assert.NotNil(t, cw.Appender)
	// EnableSyncOnError is false by default.
	assert.False(t, cw.EnableSyncOnError)
}

func TestChannelWriterDoubleClose(t *testing.T) {
	_, close := NewChannelWriter.Default()
	defer close()
	close()
}

func TestChannelWriterPanicOnWriteWhenClosed(t *testing.T) {
	w, close := NewChannelWriter.Default()
	close()

	// Panic is expected calling WriteEnter after close.
	assert.Panics(t, func() {
		w.WriteEntry(Entry{})
	})
}

func TestChannelWriterWrite(t *testing.T) {
	appender := testAppender{}
	errorAppender := testAppender{}

	w, close := NewChannelWriter(ChannelWriterConfig{Appender: &appender, ErrorAppender: &errorAppender})
	defer func() {
		assert.NotEmpty(t, appender.Entries)
		assert.EqualValues(t, 42, appender.Entries[0].LoggerID)
		assert.True(t, appender.FlushCallCounter > 0 && appender.FlushCallCounter < 3, "expected one flush at exit (and sometimes additional flush on empty channel)")
		assert.Equal(t, 1, appender.SyncCallCounter, "expected one sync at exit")
	}()
	defer close()

	w.WriteEntry(Entry{LoggerID: 42, Level: LevelInfo})
}

func TestChannelWriterFlushOnEmptyChannel(t *testing.T) {
	appender := testAppender{}
	errorAppender := testAppender{}

	w, close := NewChannelWriter(ChannelWriterConfig{Appender: &appender, ErrorAppender: &errorAppender})
	defer func() {
		assert.Len(t, appender.Entries, 1)
		assert.EqualValues(t, 42, appender.Entries[0].LoggerID)
		assert.Equal(t, 2, appender.FlushCallCounter, "expected one flush at exit and one flush on empty channel")
		assert.Equal(t, 1, appender.SyncCallCounter, "expected one sync at exit")
	}()
	defer close()

	w.WriteEntry(Entry{LoggerID: 42, Level: LevelInfo})
	runtime.Gosched()
	time.Sleep(time.Second)
}

func TestChannelWriterTestAppendError(t *testing.T) {
	appendErr := errors.New("append error")

	appender := testAppender{AppendError: appendErr}
	errorAppender := testAppender{}

	w, close := NewChannelWriter(ChannelWriterConfig{Appender: &appender, ErrorAppender: &errorAppender})
	defer func() {
		assert.Empty(t, appender.Entries, "no entries expected")
		assert.Len(t, errorAppender.Entries, 1, "expected a error message in error appender")
		assert.Equal(t, LevelError, errorAppender.Entries[0].Level)
		assert.Equal(t, 1, errorAppender.FlushCallCounter, "expected one flush on message add")
		assert.Equal(t, 1, errorAppender.SyncCallCounter, "expected one sync on message add")
	}()
	defer close()

	w.WriteEntry(Entry{LoggerID: 42, Level: LevelInfo})
}

func TestChannelWriterTestFlushError(t *testing.T) {
	flushErr := errors.New("flush error")

	appender := testAppender{FlushError: flushErr}
	errorAppender := testAppender{}

	w, close := NewChannelWriter(ChannelWriterConfig{Appender: &appender, ErrorAppender: &errorAppender})
	defer func() {
		assert.Len(t, appender.Entries, 1)
		assert.EqualValues(t, 42, appender.Entries[0].LoggerID)
		assert.Equal(t, 0, appender.FlushCallCounter, "expected no flushes because of error")
		assert.Equal(t, 1, appender.SyncCallCounter, "expected one sync at exit")

		assert.Len(t, errorAppender.Entries, 1, "expected a error message in error appender")
		assert.Equal(t, LevelError, errorAppender.Entries[0].Level)
		assert.Equal(t, 1, errorAppender.FlushCallCounter, "expected one flush on message add")
		assert.Equal(t, 1, errorAppender.SyncCallCounter, "expected one sync on message add")
	}()
	defer close()

	w.WriteEntry(Entry{LoggerID: 42, Level: LevelInfo})
}

func TestChannelWriterTestSyncError(t *testing.T) {
	syncErr := errors.New("sync error")

	appender := testAppender{SyncError: syncErr}
	errorAppender := testAppender{}

	w, close := NewChannelWriter(ChannelWriterConfig{Appender: &appender, ErrorAppender: &errorAppender})
	defer func() {
		assert.Len(t, appender.Entries, 1)
		assert.EqualValues(t, 42, appender.Entries[0].LoggerID)
		assert.Equal(t, 1, appender.FlushCallCounter, "expected one flush at exit")
		assert.Equal(t, 0, appender.SyncCallCounter, "expected no syncs because of error")

		assert.Len(t, errorAppender.Entries, 1, "expected a error message in error appender")
		assert.Equal(t, LevelError, errorAppender.Entries[0].Level)
		assert.Equal(t, 1, errorAppender.FlushCallCounter, "expected one flush on message add")
		assert.Equal(t, 1, errorAppender.SyncCallCounter, "expected one sync on message add")
	}()
	defer close()

	w.WriteEntry(Entry{LoggerID: 42, Level: LevelInfo})
}

func TestChannelWriterTestAppendErrorAndErrorAppenderError(t *testing.T) {
	appendErr := errors.New("append error")

	appender := testAppender{AppendError: appendErr}
	errorAppender := testAppender{AppendError: appendErr}

	w, close := NewChannelWriter(ChannelWriterConfig{Appender: &appender, ErrorAppender: &errorAppender})
	defer func() {
		assert.Empty(t, appender.Entries, "no entries expected")
		assert.Empty(t, errorAppender.Entries, "no entries expected")
	}()
	defer close()

	w.WriteEntry(Entry{LoggerID: 42, Level: LevelInfo})
}

func TestChannelWriterSyncOnErrorWhenEnabled(t *testing.T) {
	appender := testAppender{}
	errorAppender := testAppender{}

	w, close := NewChannelWriter(ChannelWriterConfig{Appender: &appender, ErrorAppender: &errorAppender, EnableSyncOnError: true})
	defer func() {
		assert.NotEmpty(t, appender.Entries)
		assert.EqualValues(t, 42, appender.Entries[0].LoggerID)
		assert.True(t, appender.FlushCallCounter > 1 && appender.FlushCallCounter < 4, "expected one flush at exit, one on message add (and sometimes additional flush on empty channel)")
		assert.Equal(t, 2, appender.SyncCallCounter, "expected one sync at exit and one sync on message add")
	}()
	defer close()

	w.WriteEntry(Entry{LoggerID: 42, Level: LevelError})
}
