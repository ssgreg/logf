package logf

import "context"

// With returns a new context carrying the given fields. If the context
// already has a Bag, the fields are appended to it. This is the primary
// way to attach request-scoped data (trace IDs, user info, etc.) that
// will automatically appear in every log entry — no need to pass fields
// around manually.
func With(ctx context.Context, fs ...Field) context.Context {
	bag := BagFromContext(ctx).With(fs...)

	return ContextWithBag(ctx, bag)
}

// HasField reports whether the context's Bag contains a field with the
// given key. Useful for conditional field injection — for example,
// adding a trace ID only if one is not already present.
func HasField(ctx context.Context, key string) bool {
	return BagFromContext(ctx).HasField(key)
}

// Fields returns all fields from the context's Bag, or nil if the context
// has no Bag.
func Fields(ctx context.Context) []Field {
	return BagFromContext(ctx).Fields()
}

// FieldSource is a function that extracts fields from a context. Pass one
// to NewContextHandler or LoggerBuilder.Context to automatically inject
// fields from external sources — tracing libraries, request ID middleware,
// authentication context, you name it.
type FieldSource func(ctx context.Context) []Field

// ContextHandler is the Handler middleware that makes context-based logging
// work. It extracts the Bag from the context (populated by logf.With) and
// any external fields from FieldSource functions, attaches them to the
// Entry, and passes it downstream. Without a ContextHandler in the
// pipeline, context fields are silently ignored.
type ContextHandler struct {
	next    Handler
	sources []FieldSource
}

// NewContextHandler returns a new ContextHandler wrapping the given Handler.
// Optional FieldSource functions are called on every Handle to pull in
// additional fields from the context (prepended to Entry.Fields so they
// appear before per-call fields).
func NewContextHandler(next Handler, sources ...FieldSource) *ContextHandler {
	return &ContextHandler{next: next, sources: sources}
}

// Enabled delegates to the downstream Handler to check whether the given
// level is active.
func (w *ContextHandler) Enabled(ctx context.Context, lvl Level) bool {
	return w.next.Enabled(ctx, lvl)
}

// Handle extracts the Bag from the context, collects fields from any
// registered FieldSource functions, attaches everything to the Entry,
// and hands it off to the downstream Handler.
func (w *ContextHandler) Handle(ctx context.Context, e Entry) error {
	if bag := BagFromContext(ctx); bag != nil {
		e.Bag = bag
	}

	if len(w.sources) > 0 {
		var extra []Field
		for _, src := range w.sources {
			extra = append(extra, src(ctx)...)
		}
		if len(extra) > 0 {
			e.Fields = append(extra, e.Fields...)
		}
	}

	return w.next.Handle(ctx, e)
}
