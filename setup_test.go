package logf

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoggerDefaults(t *testing.T) {
	// Build with no options — should use JSON encoder, LevelDebug, os.Stderr.
	// We redirect output to verify it produces something.
	var buf bytes.Buffer
	logger := NewLogger().Output(&buf).Build()
	require.NotNil(t, logger)

	logger.Info(context.Background(), "hello")
	assert.Contains(t, buf.String(), `"msg":"hello"`)
}

func TestNewLoggerLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger().Level(LevelError).Output(&buf).Build()

	logger.Debug(context.Background(), "debug-msg")
	logger.Info(context.Background(), "info-msg")
	logger.Warn(context.Background(), "warn-msg")
	assert.Empty(t, buf.String(), "debug/info/warn should be filtered at LevelError")

	logger.Error(context.Background(), "error-msg")
	assert.Contains(t, buf.String(), "error-msg")
}

func TestNewLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger().Output(&buf).Build()

	logger.Error(context.Background(), "to-buffer")
	assert.Contains(t, buf.String(), "to-buffer")
}

func TestNewLoggerEncoder(t *testing.T) {
	var buf bytes.Buffer
	enc := NewTextEncoder(TextEncoderConfig{NoColor: true, DisableFieldTime: true})
	logger := NewLogger().Encoder(enc).Output(&buf).Build()

	logger.Error(context.Background(), "text-output")
	out := buf.String()
	// Text encoder output contains level in brackets, not JSON.
	assert.Contains(t, out, "[ERR]")
	assert.Contains(t, out, "text-output")
}

func TestNewLoggerEncoderFrom(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger().
		EncoderFrom(JSON().TimeKey("time").DisableLevel()).
		Output(&buf).
		Build()

	logger.Error(context.Background(), "enc-from")
	out := buf.String()
	assert.Contains(t, out, `"time":`)
	assert.NotContains(t, out, `"level":`)
	assert.Contains(t, out, `"msg":"enc-from"`)
}

func TestNewLoggerContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger().Output(&buf).Context().Build()

	ctx := With(context.Background(), String("req_id", "abc123"))
	logger.Info(ctx, "with-context")
	assert.Contains(t, buf.String(), "abc123")
}

func TestNewLoggerContextWithFieldSource(t *testing.T) {
	var buf bytes.Buffer
	src := func(ctx context.Context) []Field {
		return []Field{String("injected", "yes")}
	}
	logger := NewLogger().Output(&buf).Context(src).Build()

	logger.Info(context.Background(), "with-source")
	assert.Contains(t, buf.String(), "injected")
	assert.Contains(t, buf.String(), "yes")
}

func TestNewLoggerCombined(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger().
		Level(LevelInfo).
		EncoderFrom(JSON().MsgKey("message")).
		Output(&buf).
		Context().
		Build()

	logger.Debug(context.Background(), "should-not-appear")
	assert.Empty(t, buf.String())

	logger.Info(context.Background(), "should-appear")
	assert.Contains(t, buf.String(), `"message":"should-appear"`)
}

func TestNewLoggerEncoderClearsEncoderFrom(t *testing.T) {
	// Setting Encoder after EncoderFrom should use the direct encoder.
	var buf bytes.Buffer
	enc := NewTextEncoder(TextEncoderConfig{NoColor: true, DisableFieldTime: true})
	logger := NewLogger().
		EncoderFrom(JSON()). // set builder first
		Encoder(enc).        // then override with direct encoder
		Output(&buf).
		Build()

	logger.Error(context.Background(), "direct-enc")
	assert.Contains(t, buf.String(), "[ERR]")
}

func TestNewLoggerEncoderFromClearsEncoder(t *testing.T) {
	// Setting EncoderFrom after Encoder should use the builder.
	var buf bytes.Buffer
	enc := NewTextEncoder(TextEncoderConfig{NoColor: true, DisableFieldTime: true})
	logger := NewLogger().
		Encoder(enc).          // set direct encoder first
		EncoderFrom(JSON()).   // then override with builder
		Output(&buf).
		Build()

	logger.Error(context.Background(), "builder-enc")
	// Should be JSON, not text.
	assert.Contains(t, buf.String(), `"msg":"builder-enc"`)
}
