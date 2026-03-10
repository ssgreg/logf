package logf

import (
	"context"
	"log/slog"
)

// SlogHandlerOptions configures the slog→logf bridge handler.
type SlogHandlerOptions struct {
	// Level is the minimum enabled severity.
	// If nil, defaults to slog.LevelInfo.
	Level slog.Leveler
}

// NewSlogHandler returns a [slog.Handler] that writes log records
// to the given [EntryWriter].
//
// Fields added with [slog.Logger.With] become [Entry.LoggerBag],
// which the JSON encoder caches per unique Bag version.
// Each call to WithAttrs allocates a new Bag automatically.
//
// The handler propagates context to [EntryWriter.WriteEntry],
// so field bags attached via [With] are resolved by [NewContextWriter].
func NewSlogHandler(w EntryWriter, opts *SlogHandlerOptions) slog.Handler {
	if opts == nil {
		opts = &SlogHandlerOptions{}
	}

	return &slogHandler{
		w:    w,
		opts: *opts,
	}
}

type slogHandler struct {
	w    EntryWriter
	opts SlogHandlerOptions

	// bag holds pre-resolved logf fields from WithAttrs calls
	// and group nodes from WithGroup calls (via Bag.WithGroup).
	bag *Bag
}

// Enabled reports whether the handler is enabled for the given level.
func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel()
}

// Handle converts a slog.Record to a logf.Entry and writes it.
func (h *slogHandler) Handle(ctx context.Context, r slog.Record) error {
	e := h.buildEntry(r)
	return h.w.WriteEntry(ctx, e)
}

// WithAttrs returns a new handler whose pre-resolved fields include
// the given attributes. A new Bag (with a new version) is allocated.
func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	return &slogHandler{
		w:    h.w,
		opts: h.opts,
		bag:  h.bag.With(convertAttrs(attrs)...),
	}
}

// WithGroup returns a new handler that nests subsequent fields
// under the given group name.
func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &slogHandler{
		w:    h.w,
		opts: h.opts,
		bag:  h.bag.WithGroup(name),
	}
}

// buildEntry assembles a logf.Entry from a slog.Record.
func (h *slogHandler) buildEntry(r slog.Record) Entry {
	return Entry{
		LoggerBag: h.bag,
		Fields:    h.collectRecordFields(r),
		Level:     slogLevelToLogf(r.Level),
		Text:      r.Message,
		Time:      r.Time,
		CallerPC:  r.PC,
	}
}

// collectRecordFields extracts fields from a slog.Record.
func (h *slogHandler) collectRecordFields(r slog.Record) []Field {
	if r.NumAttrs() == 0 {
		return nil
	}

	fields := make([]Field, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		if f := attrToField(a); f.Key != "" {
			fields = append(fields, f)
		}

		return true
	})

	return fields
}

// minLevel returns the configured minimum level, defaulting to LevelInfo.
func (h *slogHandler) minLevel() slog.Level {
	if h.opts.Level != nil {
		return h.opts.Level.Level()
	}

	return slog.LevelInfo
}

// slogLevelToLogf maps a slog severity to the nearest logf level.
// Intermediate slog levels (e.g. slog.LevelInfo+2) map down to the
// logf level at or below.
func slogLevelToLogf(l slog.Level) Level {
	switch {
	case l >= slog.LevelError:
		return LevelError
	case l >= slog.LevelWarn:
		return LevelWarn
	case l >= slog.LevelInfo:
		return LevelInfo
	default:
		return LevelDebug
	}
}

// attrToField converts a single slog.Attr to a logf.Field.
func attrToField(a slog.Attr) Field {
	a.Value = a.Value.Resolve()

	if a.Equal(slog.Attr{}) {
		return Field{}
	}

	switch a.Value.Kind() {
	case slog.KindBool:
		return Bool(a.Key, a.Value.Bool())
	case slog.KindInt64:
		return Int64(a.Key, a.Value.Int64())
	case slog.KindUint64:
		return Uint64(a.Key, a.Value.Uint64())
	case slog.KindFloat64:
		return Float64(a.Key, a.Value.Float64())
	case slog.KindString:
		return String(a.Key, a.Value.String())
	case slog.KindTime:
		return Time(a.Key, a.Value.Time())
	case slog.KindDuration:
		return Duration(a.Key, a.Value.Duration())
	case slog.KindGroup:
		return groupAttrToField(a.Key, a.Value.Group())
	default:
		return Any(a.Key, a.Value.Any())
	}
}

// groupAttrToField converts a slog group attribute to a logf.Group field.
// Returns a zero Field if the group is empty.
func groupAttrToField(key string, attrs []slog.Attr) Field {
	if len(attrs) == 0 {
		return Field{}
	}

	return Group(key, convertAttrs(attrs)...)
}

// convertAttrs converts slog attributes to logf fields.
func convertAttrs(attrs []slog.Attr) []Field {
	fields := make([]Field, 0, len(attrs))
	for _, a := range attrs {
		if f := attrToField(a); f.Key != "" {
			fields = append(fields, f)
		}
	}

	return fields
}
