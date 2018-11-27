package logf

import (
	"testing"

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
	}()
	defer close()

	w.WriteEntry(Entry{LoggerID: 42})
}
