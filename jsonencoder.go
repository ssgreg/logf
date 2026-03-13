package logf

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"sync"
	"time"
	"unicode/utf8"
	"unsafe"
)

// Default field keys.
const (
	DefaultFieldKeyLevel  = "level"
	DefaultFieldKeyMsg    = "msg"
	DefaultFieldKeyTime   = "ts"
	DefaultFieldKeyName   = "logger"
	DefaultFieldKeyCaller = "caller"
)

// JSONEncoderConfig allows to configure journal JSON Encoder.
type JSONEncoderConfig struct {
	FieldKeyMsg    string
	FieldKeyTime   string
	FieldKeyLevel  string
	FieldKeyName   string
	FieldKeyCaller string

	DisableFieldMsg    bool
	DisableFieldTime   bool
	DisableFieldLevel  bool
	DisableFieldName   bool
	DisableFieldCaller bool

	EncodeTime     TimeEncoder
	EncodeDuration DurationEncoder
	EncodeError    ErrorEncoder
	EncodeLevel    LevelEncoder
	EncodeCaller   CallerEncoder

	// Precomputed escaped key fragments: `"key":` — avoids EscapeString on every call.
	keyLevel  []byte
	keyTime   []byte
	keyName   []byte
	keyMsg    []byte
	keyCaller []byte
}

// WithDefaults returns the new config in which all uninitialized fields are
// filled with their default values.
func (c JSONEncoderConfig) WithDefaults() JSONEncoderConfig {
	// Handle default for predefined field names.
	if c.FieldKeyMsg == "" {
		c.FieldKeyMsg = DefaultFieldKeyMsg
	}
	if c.FieldKeyTime == "" {
		c.FieldKeyTime = DefaultFieldKeyTime
	}
	if c.FieldKeyLevel == "" {
		c.FieldKeyLevel = DefaultFieldKeyLevel
	}
	if c.FieldKeyName == "" {
		c.FieldKeyName = DefaultFieldKeyName
	}
	if c.FieldKeyCaller == "" {
		c.FieldKeyCaller = DefaultFieldKeyCaller
	}

	// Handle defaults for type encoder.
	if c.EncodeDuration == nil {
		c.EncodeDuration = StringDurationEncoder
	}
	if c.EncodeTime == nil {
		c.EncodeTime = RFC3339TimeEncoder
	}
	if c.EncodeError == nil {
		c.EncodeError = DefaultErrorEncoder
	}
	if c.EncodeLevel == nil {
		c.EncodeLevel = DefaultLevelEncoder
	}
	if c.EncodeCaller == nil {
		c.EncodeCaller = ShortCallerEncoder
	}

	c.keyLevel = precomputeKey(c.FieldKeyLevel)
	c.keyTime = precomputeKey(c.FieldKeyTime)
	c.keyName = precomputeKey(c.FieldKeyName)
	c.keyMsg = precomputeKey(c.FieldKeyMsg)
	c.keyCaller = precomputeKey(c.FieldKeyCaller)

	return c
}

// NewJSONEncoder creates a new JSON Encoder with the given config.
//
// Each encoder owns a dedicated sync.Pool whose New func pre-fills
// JSONEncoderConfig and slot. This avoids copying ~200 bytes of config
// on every Encode call (as a global pool would require) and eliminates
// pointer indirection that a *JSONEncoderConfig approach would add to
// every field access in the hot path.
func NewJSONEncoder(cfg JSONEncoderConfig) Encoder {
	enc := &jsonEncoder{JSONEncoderConfig: cfg.WithDefaults(), slot: AllocEncoderSlot()}
	enc.pool = &sync.Pool{New: func() any {
		return &jsonEncoder{
			JSONEncoderConfig: enc.JSONEncoderConfig,
			slot:              enc.slot,
		}
	}}
	return enc
}

type jsonEncoder struct {
	JSONEncoderConfig
	pool        *sync.Pool
	slot        int

	// Internal state.
	buf         *Buffer
	startBufLen int
}

func (f *jsonEncoder) TypeEncoder(buf *Buffer) TypeEncoder {
	f.buf = buf
	f.startBufLen = f.buf.Len()

	return f
}

func (f *jsonEncoder) Clone() Encoder {
	return &jsonEncoder{
		JSONEncoderConfig: f.JSONEncoderConfig,
		slot:              f.slot,
		pool:              f.pool,
	}
}


func (f *jsonEncoder) Encode(e Entry) (*Buffer, error) {
	clone := f.pool.Get().(*jsonEncoder)

	buf := GetBuffer()
	err := clone.encode(buf, e)

	clone.buf = nil
	f.pool.Put(clone)

	if err != nil {
		buf.Free()
		return nil, err
	}
	return buf, nil
}

func (f *jsonEncoder) encode(buf *Buffer, e Entry) error {
	f.buf = buf
	f.startBufLen = buf.Len()

	f.buf.AppendByte('{')

	// Level.
	if !f.DisableFieldLevel {
		f.addPrecomputedKey(f.keyLevel)
		cur := f.buf.Len()
		f.EncodeLevel(e.Level, f)
		if f.buf.Len() == cur {
			f.EncodeTypeString(e.Level.String())
		}
	}

	// Time.
	if !f.DisableFieldTime {
		f.addPrecomputedKey(f.keyTime)
		cur := f.buf.Len()
		f.EncodeTime(e.Time, f)
		if f.buf.Len() == cur {
			f.EncodeTypeInt64(e.Time.UnixNano())
		}
	}

	// Logger name.
	if !f.DisableFieldName && e.LoggerName != "" {
		f.addPrecomputedKey(f.keyName)
		f.EncodeTypeString(e.LoggerName)
	}

	// Message.
	if !f.DisableFieldMsg {
		f.addPrecomputedKey(f.keyMsg)
		f.EncodeTypeString(e.Text)
	}

	// Caller.
	if !f.DisableFieldCaller && e.CallerPC != 0 {
		f.addPrecomputedKey(f.keyCaller)
		cur := f.buf.Len()
		f.EncodeCaller(e.CallerPC, f)
		if f.buf.Len() == cur {
			f.EncodeTypeString("unknown")
		}
	}

	// Context fields (request-scoped).
	f.encodeBag(e.Bag)

	// Logger's fields (service-scoped).
	f.encodeBag(e.LoggerBag)

	// Entry's fields.
	for i := range e.Fields {
		e.Fields[i].Accept(f)
	}

	// Close open groups (from WithGroup in Bag chains).
	for n := countGroups(e.Bag) + countGroups(e.LoggerBag); n > 0; n-- {
		f.buf.AppendByte('}')
	}

	f.buf.AppendByte('}')
	f.buf.AppendByte('\n')

	return nil
}

// countGroups counts group nodes in a Bag chain.
func countGroups(bag *Bag) int {
	n := 0
	for b := bag; b != nil; b = b.parent {
		if b.group != "" {
			n++
		}
	}
	return n
}

func (f *jsonEncoder) encodeBag(bag *Bag) {
	if bag == nil {
		return
	}

	// Group node: just open a nested JSON object, no caching needed.
	if bag.group != "" {
		f.encodeBag(bag.parent)
		f.addKey(bag.group)
		f.buf.AppendByte('{')
		return
	}

	// Field node: use cache.
	if data := bag.LoadCache(f.slot); data != nil {
		f.buf.AppendBytes(data)
		return
	}

	start := f.buf.Len()

	// Walk parent first to preserve field order (parent before child).
	f.encodeBag(bag.parent)

	for _, field := range bag.fields {
		field.Accept(f)
	}

	// Cache the encoded bytes (includes parent content).
	if f.slot != 0 {
		encoded := make([]byte, f.buf.Len()-start)
		copy(encoded, f.buf.Data[start:])
		bag.StoreCache(f.slot, encoded)
	}
}

func (f *jsonEncoder) EncodeFieldAny(k string, v interface{}) {
	f.addKey(k)
	f.EncodeTypeAny(v)
}

func (f *jsonEncoder) EncodeFieldBool(k string, v bool) {
	f.addKey(k)
	f.EncodeTypeBool(v)
}

func (f *jsonEncoder) EncodeFieldInt64(k string, v int64) {
	f.addKey(k)
	f.EncodeTypeInt64(v)
}

func (f *jsonEncoder) EncodeFieldUint64(k string, v uint64) {
	f.addKey(k)
	f.EncodeTypeUint64(v)
}

func (f *jsonEncoder) EncodeFieldFloat64(k string, v float64) {
	f.addKey(k)
	f.EncodeTypeFloat64(v)
}

func (f *jsonEncoder) EncodeFieldString(k string, v string) {
	f.addKey(k)
	f.EncodeTypeString(v)
}

func (f *jsonEncoder) EncodeFieldStrings(k string, v []string) {
	f.addKey(k)
	f.EncodeTypeStrings(v)
}

func (f *jsonEncoder) EncodeFieldDuration(k string, v time.Duration) {
	f.addKey(k)
	f.EncodeTypeDuration(v)
}

func (f *jsonEncoder) EncodeFieldError(k string, v error) {
	// The only exception that has no EncodeX function. EncodeError can add
	// new fields by itself.
	f.EncodeError(k, v, f)
}

func (f *jsonEncoder) EncodeFieldTime(k string, v time.Time) {
	f.addKey(k)
	f.EncodeTypeTime(v)
}

func (f *jsonEncoder) EncodeFieldArray(k string, v ArrayEncoder) {
	f.addKey(k)
	f.EncodeTypeArray(v)
}

func (f *jsonEncoder) EncodeFieldObject(k string, v ObjectEncoder) {
	f.addKey(k)
	f.EncodeTypeObject(v)
}

func (f *jsonEncoder) EncodeFieldGroup(k string, fs []Field) {
	f.addKey(k)
	f.appendSeparator()
	f.buf.AppendByte('{')
	for _, field := range fs {
		field.Accept(f)
	}
	f.buf.AppendByte('}')
}

func (f *jsonEncoder) EncodeFieldBytes(k string, v []byte) {
	f.addKey(k)
	f.EncodeTypeBytes(v)
}

func (f *jsonEncoder) EncodeFieldInts64(k string, v []int64) {
	f.addKey(k)
	f.EncodeTypeInts64(v)
}

func (f *jsonEncoder) EncodeFieldFloats64(k string, v []float64) {
	f.addKey(k)
	f.EncodeTypeFloats64(v)
}

func (f *jsonEncoder) EncodeFieldDurations(k string, v []time.Duration) {
	f.addKey(k)
	f.EncodeTypeDurations(v)
}

func (f *jsonEncoder) EncodeTypeAny(v interface{}) {
	e := json.NewEncoder(f.buf)
	_ = e.Encode(v)

	if !f.empty() && f.buf.Back() == '\n' {
		f.buf.Data = f.buf.Data[0 : f.buf.Len()-1]
	}
}

func (f *jsonEncoder) EncodeTypeUnsafeBytes(v unsafe.Pointer) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	_ = EscapeString(f.buf, *(*[]byte)(v))
	f.buf.AppendByte('"')
}

func (f *jsonEncoder) EncodeTypeBool(v bool) {
	f.appendSeparator()
	f.buf.AppendBool(v)
}

func (f *jsonEncoder) EncodeTypeString(v string) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	_ = EscapeString(f.buf, v)
	f.buf.AppendByte('"')
}

func (f *jsonEncoder) EncodeTypeInt64(v int64) {
	f.appendSeparator()
	f.buf.AppendInt(v)
}

func (f *jsonEncoder) EncodeTypeUint64(v uint64) {
	f.appendSeparator()
	f.buf.AppendUint(uint64(v))
}

func (f *jsonEncoder) EncodeTypeFloat64(v float64) {
	f.appendSeparator()
	switch {
	case math.IsNaN(v):
		f.buf.AppendString(`"NaN"`)
	case math.IsInf(v, 1):
		f.buf.AppendString(`"+Inf"`)
	case math.IsInf(v, -1):
		f.buf.AppendString(`"-Inf"`)
	default:
		f.buf.AppendFloat64(v)
	}
}

func (f *jsonEncoder) EncodeTypeDuration(v time.Duration) {
	f.appendSeparator()
	f.EncodeDuration(v, f)
}

func (f *jsonEncoder) EncodeTypeTime(v time.Time) {
	f.appendSeparator()
	f.EncodeTime(v, f)
}

func (f *jsonEncoder) EncodeTypeBytes(v []byte) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	base64.StdEncoding.Encode(f.buf.ExtendBytes(base64.StdEncoding.EncodedLen(len(v))), v)
	f.buf.AppendByte('"')
}

func (f *jsonEncoder) EncodeTypeStrings(v []string) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeString(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeInts64(v []int64) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeInt64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeFloats64(v []float64) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeFloat64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeDurations(v []time.Duration) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeDuration(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeArray(v ArrayEncoder) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	_ = v.EncodeLogfArray(f)
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeObject(v ObjectEncoder) {
	f.appendSeparator()
	f.buf.AppendByte('{')
	_ = v.EncodeLogfObject(f)
	f.buf.AppendByte('}')
}

func (f *jsonEncoder) appendSeparator() {
	if f.empty() {
		return
	}

	switch f.buf.Back() {
	case '{', '[', ':', ',':
		return
	}
	f.buf.AppendByte(',')
}

func (f *jsonEncoder) empty() bool {
	return f.buf.Len() == f.startBufLen
}

func (f *jsonEncoder) addPrecomputedKey(key []byte) {
	f.appendSeparator()
	f.buf.AppendBytes(key)
}

func (f *jsonEncoder) addKey(k string) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	_ = EscapeString(f.buf, k)
	f.buf.AppendByte('"')
	f.buf.AppendByte(':')
}

const hex = "0123456789abcdef"

// EscapeString JSON-escapes s and appends the result to buf.
// s can be a string or []byte.
func EscapeString[S []byte | string](buf *Buffer, s S) error {
	p := 0
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c >= 0x20 && c < utf8.RuneSelf && c != '\\' && c != '"':
			i++

			continue

		case c < utf8.RuneSelf:
			appendBuf(buf, s[p:i])
			switch c {
			case '\t':
				buf.AppendString(`\t`)
			case '\r':
				buf.AppendString(`\r`)
			case '\n':
				buf.AppendString(`\n`)
			case '\\':
				buf.AppendString(`\\`)
			case '"':
				buf.AppendString(`\"`)
			default:
				buf.AppendString(`\u00`)
				buf.AppendByte(hex[c>>4])
				buf.AppendByte(hex[c&0xf])
			}
			i++
			p = i

			continue
		}

		// Multi-byte UTF-8.
		r, size := decodeRuneIn(s[i:])
		if r == utf8.RuneError && size == 1 {
			appendBuf(buf, s[p:i])
			buf.AppendString(`\ufffd`)
			i++
			p = i
		} else {
			i += size
		}
	}
	appendBuf(buf, s[p:])

	return nil
}

func appendBuf[S []byte | string](buf *Buffer, s S) {
	switch v := any(s).(type) {
	case string:
		buf.AppendString(v)
	case []byte:
		buf.AppendBytes(v)
	}
}

func decodeRuneIn[S []byte | string](s S) (rune, int) {
	switch v := any(s).(type) {
	case string:
		return utf8.DecodeRuneInString(v)
	case []byte:
		return utf8.DecodeRune(v)
	}
	return utf8.RuneError, 0
}

// precomputeKey builds the escaped JSON key fragment `"key":` once.
func precomputeKey(k string) []byte {
	out := make([]byte, 0, len(k)+3)
	out = append(out, '"')
	needsEscape := false
	for i := 0; i < len(k); i++ {
		if k[i] < 0x20 || k[i] == '\\' || k[i] == '"' || k[i] >= utf8.RuneSelf {
			needsEscape = true
			break
		}
	}
	if !needsEscape {
		out = append(out, k...)
	} else {
		b := GetBuffer()
		_ = EscapeString(b, k)
		out = append(out, b.Data[:b.Len()]...)
		b.Free()
	}
	out = append(out, '"', ':')
	return out
}
