package logf

import (
	"io"
	"os"
)

// NewLogger returns a LoggerBuilder — the easiest way to get a Logger up
// and running. It builds a single-destination sync pipeline, which is
// perfect for most applications. For async buffered or multi-destination
// setups, reach for NewRouter + SlabWriter instead.
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
// pipeline: Encoder -> SyncHandler -> [ContextHandler] -> Logger.
// Chain its methods and finish with Build. For advanced multi-destination
// pipelines, use NewRouter directly.
type LoggerBuilder struct {
	level   Level
	enc     Encoder
	encB    EncoderBuilder
	w       io.Writer
	context bool
	sources []FieldSource
}

// Level sets the minimum severity level. Messages below this level are
// discarded. Default is LevelDebug (everything gets through).
func (b *LoggerBuilder) Level(l Level) *LoggerBuilder {
	b.level = l
	return b
}

// Output sets where encoded log entries are written. Default is os.Stderr.
func (b *LoggerBuilder) Output(w io.Writer) *LoggerBuilder {
	b.w = w
	return b
}

// Encoder sets a pre-built Encoder directly. Use this when you already
// have an Encoder instance; otherwise prefer EncoderFrom for builder
// composition.
func (b *LoggerBuilder) Encoder(enc Encoder) *LoggerBuilder {
	b.enc = enc
	b.encB = nil
	return b
}

// EncoderFrom sets an EncoderBuilder whose Build method will be called
// when LoggerBuilder.Build is called. This enables clean builder
// composition — no need to call Build on the encoder separately:
//
//	logf.NewLogger().EncoderFrom(logf.JSON().TimeKey("time")).Build()
func (b *LoggerBuilder) EncoderFrom(eb EncoderBuilder) *LoggerBuilder {
	b.encB = eb
	b.enc = nil
	return b
}

// Context enables the ContextHandler middleware, which extracts Bag fields
// from context on every log call. This is what makes logf.With(ctx, ...)
// work — without it, context fields are silently ignored. Optional
// FieldSource functions let you pull in external fields too (trace IDs,
// request metadata, etc.).
func (b *LoggerBuilder) Context(sources ...FieldSource) *LoggerBuilder {
	b.context = true
	b.sources = append(b.sources, sources...)
	return b
}

// Build finalizes the configuration and returns a ready-to-use Logger.
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
