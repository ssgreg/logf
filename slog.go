package logf

import (
	"context"
	"log/slog"
	"math"
	"unsafe"
)

// NewSlogHandler returns a [slog.Handler] that bridges the standard library's
// slog package to logf's pipeline. Use this when you want third-party code
// that speaks slog to flow through your logf Handler, Encoder, and Writer
// setup.
//
// Fields added with [slog.Logger.With] become [Entry.LoggerBag] (cached by
// the encoder). The handler propagates context to [Handler.Handle], so
// field bags attached via [With] are resolved by [NewContextHandler].
func NewSlogHandler(w Handler) slog.Handler {
	return &slogHandler{w: w, addCaller: true}
}

type slogHandler struct {
	w Handler

	bag       *Bag
	name      string
	addCaller bool
}

// Enabled reports whether the handler is enabled for the given level.
func (h *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.w.Enabled(ctx, slogLevelToLogf(level))
}

// Handle converts a slog.Record to a logf.Entry and writes it.
func (h *slogHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.w.Handle(ctx, h.buildEntry(r))
}

// WithAttrs returns a new handler whose pre-resolved fields include
// the given attributes. A new Bag (with a new version) is allocated.
func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	return &slogHandler{
		w:         h.w,
		bag:       h.bag.With(convertAttrs(attrs)...),
		name:      h.name,
		addCaller: h.addCaller,
	}
}

// WithGroup returns a new handler that nests subsequent fields
// under the given group name.
func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &slogHandler{
		w:         h.w,
		bag:       h.bag.WithGroup(name),
		name:      h.name,
		addCaller: h.addCaller,
	}
}

// buildEntry assembles a logf.Entry from a slog.Record.
func (h *slogHandler) buildEntry(r slog.Record) Entry {
	e := Entry{
		LoggerBag:  h.bag,
		Fields:     h.collectRecordFields(r),
		Level:      slogLevelToLogf(r.Level),
		LoggerName: h.name,
		Text:       r.Message,
		Time:       r.Time,
	}
	if h.addCaller {
		e.CallerPC = r.PC
	}

	return e
}

func (h *slogHandler) collectRecordFields(r slog.Record) []Field {
	if r.NumAttrs() == 0 {
		return nil
	}

	fields := make([]Field, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		if f := attrToField(a); f.Key != "" || f.Type == FieldTypeGroup {
			fields = append(fields, f)
		}

		return true
	})

	return fields
}

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

func attrToField(a slog.Attr) Field {
	a.Value = a.Value.Resolve()

	if a.Equal(slog.Attr{}) {
		return Field{}
	}

	key := a.Key
	v := a.Value
	switch v.Kind() {
	case slog.KindBool:
		var val int64
		if v.Bool() {
			val = 1
		}
		return Field{Key: key, Type: FieldTypeBool, Val: val}
	case slog.KindInt64:
		return Field{Key: key, Type: FieldTypeInt64, Val: v.Int64()}
	case slog.KindUint64:
		return Field{Key: key, Type: FieldTypeUint64, Val: int64(v.Uint64())}
	case slog.KindFloat64:
		return Field{Key: key, Type: FieldTypeFloat64, Val: int64(math.Float64bits(v.Float64()))}
	case slog.KindString:
		s := v.String()
		return Field{Key: key, Type: FieldTypeBytesToString, Ptr: unsafe.Pointer(unsafe.StringData(s)), Val: int64(len(s))}
	case slog.KindTime:
		t := v.Time()
		return Field{Key: key, Type: FieldTypeTime, Val: t.UnixNano(), Any: t.Location()}
	case slog.KindDuration:
		return Field{Key: key, Type: FieldTypeDuration, Val: int64(v.Duration())}
	case slog.KindGroup:
		return groupAttrToField(key, v.Group())
	default:
		return Any(key, v.Any())
	}
}

func groupAttrToField(key string, attrs []slog.Attr) Field {
	if len(attrs) == 0 {
		return Field{}
	}

	return Group(key, convertAttrs(attrs)...)
}

func convertAttrs(attrs []slog.Attr) []Field {
	fields := make([]Field, 0, len(attrs))
	for _, a := range attrs {
		if f := attrToField(a); f.Key != "" || f.Type == FieldTypeGroup {
			fields = append(fields, f)
		}
	}

	return fields
}
