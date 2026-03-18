package logf

import "context"

// With returns a new context with the given fields added to the context's Bag.
// If the context already has a Bag, the fields are appended to it.
func With(ctx context.Context, fs ...Field) context.Context {
	bag := BagFromContext(ctx).With(fs...)

	return ContextWithBag(ctx, bag)
}

// HasField reports whether the context's Bag contains a field with the given key.
func HasField(ctx context.Context, key string) bool {
	return BagFromContext(ctx).HasField(key)
}

// Fields returns the fields from the context's Bag, or nil.
func Fields(ctx context.Context) []Field {
	return BagFromContext(ctx).Fields()
}

// FieldSource extracts fields from a context. It is used by ContextHandler
// to support external field sources (e.g. tracing, request ID middleware).
type FieldSource func(ctx context.Context) []Field

// ContextHandler is an Handler middleware that extracts the Bag and
// external fields from the context and attaches them to the Entry before
// passing it downstream.
type ContextHandler struct {
	next    Handler
	sources []FieldSource
}

// NewContextHandler returns a new ContextHandler wrapping the given Handler.
// Optional FieldSource functions are called on each Handle to collect
// additional fields from the context. These fields are prepended to Entry.Fields.
func NewContextHandler(next Handler, sources ...FieldSource) *ContextHandler {
	return &ContextHandler{next: next, sources: sources}
}

// Handle extracts the Bag from ctx, collects fields from external sources,
// and delegates to the next Handler.
func (w *ContextHandler) Enabled(ctx context.Context, lvl Level) bool {
	return w.next.Enabled(ctx, lvl)
}

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
