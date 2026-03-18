package logf

import (
	"io"
	"os"
)

// NewLogger returns a LoggerBuilder for constructing a Logger with a
// single-destination sync pipeline. For async buffered or multi-destination
// setups use NewRouter + SlabWriter.
//
// Defaults: JSON encoder, LevelDebug, os.Stderr, caller enabled,
// no ContextHandler.
//
//	logger := logf.NewLogger().Build()
//
//	// Customized:
//	logger := logf.NewLogger().
//	    Level(logf.LevelInfo).
//	    EncoderFrom(logf.JSON().TimeKey("time")).
//	    Output(file).
//	    Context().
//	    Build()
func NewLogger() *LoggerBuilder {
	return &LoggerBuilder{}
}

// LoggerBuilder accumulates options and builds a Logger with a sync
// pipeline: Encoder → SyncHandler → [ContextHandler] → Logger.
type LoggerBuilder struct {
	level   Level
	enc     Encoder
	encB    EncoderBuilder
	w       io.Writer
	context bool
	sources []FieldSource
}

// Level sets the minimum logging level. Default is LevelDebug.
func (b *LoggerBuilder) Level(l Level) *LoggerBuilder {
	b.level = l
	return b
}

// Output sets the output writer. Default is os.Stderr.
func (b *LoggerBuilder) Output(w io.Writer) *LoggerBuilder {
	b.w = w
	return b
}

// Encoder sets a pre-built Encoder.
func (b *LoggerBuilder) Encoder(enc Encoder) *LoggerBuilder {
	b.enc = enc
	b.encB = nil
	return b
}

// EncoderFrom sets an EncoderBuilder whose BuildEncoder will be called
// during Build. This enables builder composition:
//
//	logf.NewLogger().EncoderFrom(logf.JSON().TimeKey("time")).Build()
func (b *LoggerBuilder) EncoderFrom(eb EncoderBuilder) *LoggerBuilder {
	b.encB = eb
	b.enc = nil
	return b
}

// Context enables ContextHandler which extracts Bag fields from context
// on each log entry. Optional FieldSource functions add external field
// extraction (e.g. trace IDs, request metadata).
func (b *LoggerBuilder) Context(sources ...FieldSource) *LoggerBuilder {
	b.context = true
	b.sources = append(b.sources, sources...)
	return b
}

// Build finalizes the configuration and returns a ready Logger.
//
//	logger := logf.NewLogger().Build()
func (b *LoggerBuilder) Build() *Logger {
	// Encoder.
	var enc Encoder
	switch {
	case b.encB != nil:
		enc = b.encB.Build()
	case b.enc != nil:
		enc = b.enc
	default:
		enc = NewJSONEncoder(JSONEncoderConfig{})
	}

	// Writer.
	w := b.w
	if w == nil {
		w = os.Stderr
	}

	// Handler pipeline.
	var h Handler = NewSyncHandler(b.level, w, enc)
	if b.context {
		h = NewContextHandler(h, b.sources...)
	}

	return New(h)
}
