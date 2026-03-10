package logf

import (
	"context"
	"fmt"
	"log/slog"
)

// SlogHandlerOptions configures the slog→logf bridge handler.
type SlogHandlerOptions struct {
	// Level is the minimum enabled severity.
	// If nil, defaults to slog.LevelInfo.
	Level slog.Leveler

	// NestedGroups controls how slog.WithGroup maps to logf fields.
	//
	// When false (default), groups produce dot-prefixed flat keys
	// matching slog.TextHandler behavior:
	//
	//   WithGroup("http") + "method" → "http.method"
	//
	// When true, groups produce nested JSON objects
	// via logf.Object / ObjectEncoder:
	//
	//   WithGroup("http") + "method" → {"http":{"method":"GET"}}
	NestedGroups bool
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

	// bag holds pre-resolved logf fields from WithAttrs calls.
	bag *Bag

	// prefix accumulates dot-separated group names: "http.request."
	// Used only when NestedGroups is false.
	prefix string

	// groups accumulates group names: ["http", "request"]
	// Used only when NestedGroups is true.
	groups []string
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

	converted := h.convertAttrs(attrs)
	var bag *Bag
	if h.bag != nil {
		bag = h.bag.With(converted...)
	} else {
		bag = NewBag(converted...)
	}

	return &slogHandler{
		w:      h.w,
		opts:   h.opts,
		bag:    bag,
		prefix: h.prefix,
		groups: h.groups,
	}
}

// WithGroup returns a new handler that nests subsequent fields
// under the given group name.
func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	nh := &slogHandler{
		w:    h.w,
		opts: h.opts,
		bag:  h.bag,
	}

	if h.opts.NestedGroups {
		nh.groups = appendString(h.groups, name)
	} else {
		nh.prefix = h.prefix + name + "."
	}

	return nh
}

// buildEntry assembles a logf.Entry from a slog.Record.
func (h *slogHandler) buildEntry(r slog.Record) Entry {
	e := h.buildFields(r)
	e.Level = slogLevelToLogf(r.Level)
	e.Text = r.Message
	e.Time = r.Time
	e.CallerPC = r.PC

	return e
}

// buildFields populates LoggerBag and Fields of the entry.
//
// Dot-prefix mode: bag attrs go to LoggerBag, record attrs go to Fields.
// Nested mode:     all fields are merged, wrapped in Objects, put into Fields.
func (h *slogHandler) buildFields(r slog.Record) Entry {
	recordFields := h.collectRecordFields(r)

	if h.opts.NestedGroups && len(h.groups) > 0 {
		all := cloneAppend(h.bag.Fields(), recordFields)

		return Entry{Fields: wrapGroups(h.groups, all)}
	}

	return Entry{
		LoggerBag: h.bag,
		Fields:    recordFields,
	}
}

// collectRecordFields extracts fields from a slog.Record.
func (h *slogHandler) collectRecordFields(r slog.Record) []Field {
	if r.NumAttrs() == 0 {
		return nil
	}

	fields := make([]Field, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		if f := attrToField(h.prefix, a); f.Key != "" {
			fields = append(fields, f)
		}

		return true
	})

	return fields
}

// convertAttrs converts slog attributes to logf fields,
// applying the current prefix in dot-prefix mode.
func (h *slogHandler) convertAttrs(attrs []slog.Attr) []Field {
	prefix := h.prefix
	if h.opts.NestedGroups {
		prefix = ""
	}

	fields := make([]Field, 0, len(attrs))
	for _, a := range attrs {
		if f := attrToField(prefix, a); f.Key != "" {
			fields = append(fields, f)
		}
	}

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
//
// Conversion rules:
//   - KindBool/Int64/Uint64/Float64/String/Time/Duration → typed Field
//   - KindGroup with children → logf.Object (fieldsObject)
//   - KindGroup empty         → skipped (zero Field)
//   - error value             → logf.NamedError
//   - fmt.Stringer value      → logf.String (eagerly resolved)
//   - anything else           → logf.Any
//
// LogValuer values are resolved before conversion via slog.Value.Resolve.
func attrToField(prefix string, a slog.Attr) Field {
	a.Value = a.Value.Resolve()

	if a.Equal(slog.Attr{}) {
		return Field{}
	}

	key := prefix + a.Key

	switch a.Value.Kind() {
	case slog.KindBool:
		return Bool(key, a.Value.Bool())
	case slog.KindInt64:
		return Int64(key, a.Value.Int64())
	case slog.KindUint64:
		return Uint64(key, a.Value.Uint64())
	case slog.KindFloat64:
		return Float64(key, a.Value.Float64())
	case slog.KindString:
		return String(key, a.Value.String())
	case slog.KindTime:
		return Time(key, a.Value.Time())
	case slog.KindDuration:
		return Duration(key, a.Value.Duration())
	case slog.KindGroup:
		return groupAttrToField(key, a.Value.Group())
	default:
		return anyToField(key, a.Value.Any())
	}
}

// groupAttrToField converts a slog group attribute to a logf.Object field.
// Returns a zero Field if the group is empty.
func groupAttrToField(key string, attrs []slog.Attr) Field {
	if len(attrs) == 0 {
		return Field{}
	}

	inner := make(fieldsObject, 0, len(attrs))
	for _, ga := range attrs {
		if f := attrToField("", ga); f.Key != "" {
			inner = append(inner, f)
		}
	}

	return Object(key, inner)
}

// anyToField handles slog.KindAny with special cases for error
// and fmt.Stringer interfaces.
func anyToField(key string, v any) Field {
	if err, ok := v.(error); ok {
		return NamedError(key, err)
	}
	if s, ok := v.(fmt.Stringer); ok {
		return String(key, s.String())
	}

	return Any(key, v)
}

// wrapGroups folds fields into nested logf.Object fields, right-to-left:
//
//	["http", "request"] + [method=GET]
//	→ [Object("http", Object("request", method=GET))]
func wrapGroups(groups []string, fields []Field) []Field {
	for i := len(groups) - 1; i >= 0; i-- {
		fields = []Field{Object(groups[i], fieldsObject(fields))}
	}

	return fields
}

// fieldsObject adapts a []Field slice to the ObjectEncoder interface,
// allowing it to be used as the value of a logf.Object field.
type fieldsObject []Field

func (fo fieldsObject) EncodeLogfObject(enc FieldEncoder) error {
	for _, f := range fo {
		f.Accept(enc)
	}

	return nil
}

// cloneAppend creates a new slice containing all elements of base
// followed by all elements of extra. The base slice is never mutated.
func cloneAppend(base, extra []Field) []Field {
	result := make([]Field, len(base), len(base)+len(extra))
	copy(result, base)

	return append(result, extra...)
}

// appendString returns a new slice with s appended.
// The original slice is never mutated.
func appendString(ss []string, s string) []string {
	result := make([]string, len(ss)+1)
	copy(result, ss)
	result[len(ss)] = s

	return result
}
