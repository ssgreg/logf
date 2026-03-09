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

// FieldSource extracts fields from a context. It is used by ContextWriter
// to support external field sources (e.g. tracing, request ID middleware).
type FieldSource func(ctx context.Context) []Field

// ContextWriter is an EntryWriter middleware that extracts the Bag and
// external fields from the context and attaches them to the Entry before
// passing it downstream.
type ContextWriter struct {
	next    EntryWriter
	sources []FieldSource
}

// NewContextWriter returns a new ContextWriter wrapping the given EntryWriter.
// Optional FieldSource functions are called on each WriteEntry to collect
// additional fields from the context. These fields are prepended to Entry.Fields.
func NewContextWriter(next EntryWriter, sources ...FieldSource) *ContextWriter {
	return &ContextWriter{next: next, sources: sources}
}

// WriteEntry extracts the Bag from ctx, collects fields from external sources,
// and delegates to the next EntryWriter.
func (w *ContextWriter) WriteEntry(ctx context.Context, e Entry) {
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

	w.next.WriteEntry(ctx, e)
}
