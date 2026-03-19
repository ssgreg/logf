package logf

import (
	"encoding/base64"
	"math"
	"sync"
	"time"
	"unsafe"
)

// TextEncoderConfig allows to configure the text Encoder.
type TextEncoderConfig struct {
	NoColor            bool
	DisableFieldTime   bool
	DisableFieldLevel  bool
	DisableFieldName   bool
	DisableFieldMsg    bool
	DisableFieldCaller bool

	EncodeTime     TimeEncoder
	EncodeDuration DurationEncoder
	EncodeError    ErrorEncoder
	EncodeLevel    LevelEncoder
	EncodeCaller   CallerEncoder
}

// WithDefaults returns the new config in which all uninitialized fields
// are filled with their default values.
func (c TextEncoderConfig) WithDefaults() TextEncoderConfig {
	if c.EncodeDuration == nil {
		c.EncodeDuration = StringDurationEncoder
	}
	if c.EncodeTime == nil {
		c.EncodeTime = LayoutTimeEncoder(time.StampMilli)
	}
	if c.EncodeError == nil {
		c.EncodeError = DefaultErrorEncoder
	}
	if c.EncodeLevel == nil {
		c.EncodeLevel = ShortTextLevelEncoder
	}
	if c.EncodeCaller == nil {
		c.EncodeCaller = ShortCallerEncoder
	}
	return c
}

// NewTextEncoder creates a new text Encoder with the given config.
func NewTextEncoder(cfg TextEncoderConfig) Encoder {
	return buildTextEncoder(cfg)
}

// Text returns a new TextEncoderBuilder with default settings.
// Colors are enabled by default. Use NoColor() to disable, or
// check the NO_COLOR environment variable (https://no-color.org):
//
//	enc := logf.Text().Build()
//	enc := logf.Text().NoColor().Build()
//
// Respect NO_COLOR convention:
//	b := logf.Text()
//	if _, ok := os.LookupEnv("NO_COLOR"); ok {
//	    b = b.NoColor()
//	}
func Text() *TextEncoderBuilder {
	return &TextEncoderBuilder{}
}

// TextEncoderBuilder configures and builds a text Encoder.
type TextEncoderBuilder struct {
	cfg TextEncoderConfig
}

func (b *TextEncoderBuilder) NoColor() *TextEncoderBuilder {
	b.cfg.NoColor = true
	return b
}

func (b *TextEncoderBuilder) DisableTime() *TextEncoderBuilder {
	b.cfg.DisableFieldTime = true
	return b
}

func (b *TextEncoderBuilder) DisableLevel() *TextEncoderBuilder {
	b.cfg.DisableFieldLevel = true
	return b
}

func (b *TextEncoderBuilder) DisableMsg() *TextEncoderBuilder {
	b.cfg.DisableFieldMsg = true
	return b
}

func (b *TextEncoderBuilder) DisableName() *TextEncoderBuilder {
	b.cfg.DisableFieldName = true
	return b
}

func (b *TextEncoderBuilder) DisableCaller() *TextEncoderBuilder {
	b.cfg.DisableFieldCaller = true
	return b
}

func (b *TextEncoderBuilder) EncodeTime(e TimeEncoder) *TextEncoderBuilder {
	b.cfg.EncodeTime = e
	return b
}

func (b *TextEncoderBuilder) EncodeDuration(e DurationEncoder) *TextEncoderBuilder {
	b.cfg.EncodeDuration = e
	return b
}

func (b *TextEncoderBuilder) EncodeLevel(e LevelEncoder) *TextEncoderBuilder {
	b.cfg.EncodeLevel = e
	return b
}

func (b *TextEncoderBuilder) EncodeCaller(e CallerEncoder) *TextEncoderBuilder {
	b.cfg.EncodeCaller = e
	return b
}

func (b *TextEncoderBuilder) EncodeError(e ErrorEncoder) *TextEncoderBuilder {
	b.cfg.EncodeError = e
	return b
}

// Build finalizes the configuration and returns a ready Encoder.
func (b *TextEncoderBuilder) Build() Encoder {
	return buildTextEncoder(b.cfg)
}

func buildTextEncoder(cfg TextEncoderConfig) Encoder {
	cfg = cfg.WithDefaults()
	// Shared JSON encoder for nested object/array/any rendering.
	jsonTEF := buildJSONEncoder(JSONEncoderConfig{
		EncodeTime:     cfg.EncodeTime,
		EncodeDuration: cfg.EncodeDuration,
		EncodeError:    cfg.EncodeError,
	}).(TypeEncoderFactory)
	enc := &textEncoder{
		TextEncoderConfig: cfg,
		slot:              AllocEncoderSlot(),
		eseq:              escSeq{noColor: cfg.NoColor},
		jsonTEF:           jsonTEF,
	}
	enc.pool = &sync.Pool{New: func() any {
		return &textEncoder{
			TextEncoderConfig: enc.TextEncoderConfig,
			slot:              enc.slot,
			eseq:              enc.eseq,
			jsonTEF:           enc.jsonTEF,
		}
	}}
	return enc
}

type textEncoder struct {
	TextEncoderConfig
	pool *sync.Pool
	slot int

	buf          *Buffer
	startBufLen  int
	eseq         escSeq
	jsonTEF      TypeEncoderFactory
	fieldSepDone bool
	groupDepth   int
	groupPrefix  string
}

func (f *textEncoder) Clone() Encoder {
	return &textEncoder{
		TextEncoderConfig: f.TextEncoderConfig,
		slot:              f.slot,
		pool:              f.pool,
		eseq:              f.eseq,
	}
}

func (f *textEncoder) Encode(e Entry) (*Buffer, error) {
	clone := f.pool.Get().(*textEncoder)

	buf := GetBuffer()
	err := clone.encode(buf, e)

	clone.buf = nil
	clone.groupPrefix = ""
	clone.groupDepth = 0
	clone.fieldSepDone = false
	f.pool.Put(clone)

	if err != nil {
		buf.Free()
		return nil, err
	}
	return buf, nil
}

func (f *textEncoder) encode(buf *Buffer, e Entry) error {
	f.buf = buf
	f.startBufLen = buf.Len()

	// Time.
	if !f.DisableFieldTime && !e.Time.IsZero() {
		f.eseq.dim(f.buf, func() {
			f.appendTime(e.Time)
		})
	}

	// Level.
	if !f.DisableFieldLevel {
		f.appendSeparator()
		f.appendLevel(e.Level)
	}

	// Logger name.
	if !f.DisableFieldName && e.LoggerName != "" {
		f.appendSeparator()
		f.eseq.dimItalic(f.buf, func() {
			f.buf.AppendString(e.LoggerName)
			f.buf.AppendByte(':')
		})
	}

	// Message — bold + level color for WRN/ERR, bold only for others.
	if !f.DisableFieldMsg && e.Text != "" {
		f.appendSeparator()
		mc := msgColor(e.Level)
		if mc == escDefault {
			f.eseq.at(f.buf, escBold, func() {
				f.buf.AppendString(e.Text)
			})
		} else {
			f.eseq.at2(f.buf, escBold, mc, func() {
				f.buf.AppendString(e.Text)
			})
		}
	}

	// › separator will be emitted lazily on first addKey call.
	f.fieldSepDone = false

	// Skip trailing groups that would produce empty output.
	loggerBag := e.LoggerBag
	ctxBag := e.Bag
	if len(e.Fields) == 0 {
		loggerBag = skipTrailingGroups(loggerBag)
		if !bagHasFields(loggerBag) {
			ctxBag = skipTrailingGroups(ctxBag)
		}
	}

	// Context fields.
	f.encodeBag(ctxBag)

	// Logger's fields.
	f.encodeBag(loggerBag)

	// Entry's fields.
	for i := range e.Fields {
		e.Fields[i].Accept(f)
	}

	// Caller — at the very end, after all fields.
	if !f.DisableFieldCaller && e.CallerPC != 0 {
		f.appendSeparator()
		f.eseq.dimItalic(f.buf, func() {
			f.buf.AppendString("→ ")
			f.EncodeCaller(e.CallerPC, f.TypeEncoder(f.buf))
		})
	}

	f.buf.AppendByte('\n')
	return nil
}

func (f *textEncoder) encodeBag(bag *Bag) {
	if bag == nil {
		return
	}
	if bag.group != "" {
		f.encodeBag(bag.parent)
		// Push group prefix for nested fields.
		f.groupPrefix += bag.group + "."
		f.groupDepth++
		return
	}

	// Field node: use cache.
	if data := bag.LoadCache(f.slot); data != nil {
		f.buf.AppendBytes(data)
		return
	}

	start := f.buf.Len()
	f.encodeBag(bag.parent)
	for _, field := range bag.fields {
		field.Accept(f)
	}

	if f.slot != 0 {
		encoded := make([]byte, f.buf.Len()-start)
		copy(encoded, f.buf.Data[start:])
		bag.StoreCache(f.slot, encoded)
	}
}

// --- TypeEncoder ---

func (f *textEncoder) TypeEncoder(buf *Buffer) TypeEncoder {
	f.buf = buf
	f.startBufLen = f.buf.Len()
	return f
}

func (f *textEncoder) EncodeTypeAny(v interface{}) {
	f.jsonTypeEncoder().EncodeTypeAny(v)
}

func (f *textEncoder) EncodeTypeBool(v bool) {
	f.eseq.at(f.buf, escGreen, func() {
		f.buf.AppendBool(v)
	})
}

func (f *textEncoder) EncodeTypeInt64(v int64) {
	f.eseq.at(f.buf, escGreen, func() {
		f.buf.AppendInt(v)
	})
}

func (f *textEncoder) EncodeTypeUint64(v uint64) {
	f.eseq.at(f.buf, escGreen, func() {
		f.buf.AppendUint(v)
	})
}

func (f *textEncoder) EncodeTypeFloat64(v float64) {
	f.eseq.at(f.buf, escGreen, func() {
		switch {
		case math.IsNaN(v):
			f.buf.AppendString("NaN")
		case math.IsInf(v, 1):
			f.buf.AppendString("+Inf")
		case math.IsInf(v, -1):
			f.buf.AppendString("-Inf")
		default:
			f.buf.AppendFloat64(v)
		}
	})
}

func (f *textEncoder) EncodeTypeDuration(v time.Duration) {
	f.EncodeDuration(v, f)
}

func (f *textEncoder) EncodeTypeTime(v time.Time) {
	f.EncodeTime(v, f)
}

func (f *textEncoder) EncodeTypeString(v string) {
	// Quote if contains spaces or special chars.
	needsQuote := false
	for i := 0; i < len(v); i++ {
		if v[i] <= ' ' || v[i] == '"' || v[i] == '\\' {
			needsQuote = true
			break
		}
	}
	if needsQuote {
		f.buf.AppendByte('"')
		_ = EscapeString(f.buf, v)
		f.buf.AppendByte('"')
	} else {
		f.buf.AppendString(v)
	}
}

func (f *textEncoder) EncodeTypeStrings(v []string) {
	f.buf.AppendByte('[')
	for i, s := range v {
		if i > 0 {
			f.buf.AppendByte(',')
		}
		f.EncodeTypeString(s)
	}
	f.buf.AppendByte(']')
}

func (f *textEncoder) EncodeTypeBytes(v []byte) {
	f.buf.AppendByte('"')
	base64.StdEncoding.Encode(f.buf.ExtendBytes(base64.StdEncoding.EncodedLen(len(v))), v)
	f.buf.AppendByte('"')
}

func (f *textEncoder) EncodeTypeInts64(v []int64) {
	f.buf.AppendByte('[')
	for i, n := range v {
		if i > 0 {
			f.buf.AppendByte(',')
		}
		f.buf.AppendInt(n)
	}
	f.buf.AppendByte(']')
}

func (f *textEncoder) EncodeTypeFloats64(v []float64) {
	f.buf.AppendByte('[')
	for i, n := range v {
		if i > 0 {
			f.buf.AppendByte(',')
		}
		f.EncodeTypeFloat64(n)
	}
	f.buf.AppendByte(']')
}

func (f *textEncoder) EncodeTypeDurations(v []time.Duration) {
	f.buf.AppendByte('[')
	for i, d := range v {
		if i > 0 {
			f.buf.AppendByte(',')
		}
		f.EncodeTypeDuration(d)
	}
	f.buf.AppendByte(']')
}

func (f *textEncoder) EncodeTypeArray(v ArrayEncoder) {
	f.jsonTypeEncoder().EncodeTypeArray(v)
}

func (f *textEncoder) EncodeTypeObject(v ObjectEncoder) {
	f.jsonTypeEncoder().EncodeTypeObject(v)
}

// jsonTypeEncoder returns a JSON TypeEncoder writing to f.buf.
// Used for nested objects/arrays where JSON is more readable than key=value.
func (f *textEncoder) jsonTypeEncoder() TypeEncoder {
	return f.jsonTEF.TypeEncoder(f.buf)
}

func (f *textEncoder) EncodeTypeUnsafeBytes(v unsafe.Pointer) {
	f.EncodeTypeString(*(*string)(v))
}

// --- FieldEncoder ---

func (f *textEncoder) addKey(k string) {
	if !f.fieldSepDone {
		// First field — emit › separator.
		f.fieldSepDone = true
		f.appendSeparator()
		f.eseq.dim(f.buf, func() {
			f.buf.AppendString("›")
		})
	}
	f.appendSeparator()
	f.eseq.at2(f.buf, escBrightBlue, escItalic, func() {
		if f.groupPrefix != "" {
			f.buf.AppendString(f.groupPrefix)
		}
		f.buf.AppendString(k)
	})
	f.eseq.dim(f.buf, func() {
		f.buf.AppendByte('=')
	})
}

func (f *textEncoder) EncodeFieldAny(k string, v interface{}) { f.addKey(k); f.EncodeTypeAny(v) }
func (f *textEncoder) EncodeFieldBool(k string, v bool)       { f.addKey(k); f.EncodeTypeBool(v) }
func (f *textEncoder) EncodeFieldInt64(k string, v int64)     { f.addKey(k); f.EncodeTypeInt64(v) }
func (f *textEncoder) EncodeFieldUint64(k string, v uint64)   { f.addKey(k); f.EncodeTypeUint64(v) }
func (f *textEncoder) EncodeFieldFloat64(k string, v float64) { f.addKey(k); f.EncodeTypeFloat64(v) }
func (f *textEncoder) EncodeFieldDuration(k string, v time.Duration) {
	f.addKey(k)
	f.EncodeTypeDuration(v)
}
func (f *textEncoder) EncodeFieldTime(k string, v time.Time)   { f.addKey(k); f.EncodeTypeTime(v) }
func (f *textEncoder) EncodeFieldString(k string, v string)    { f.addKey(k); f.EncodeTypeString(v) }
func (f *textEncoder) EncodeFieldStrings(k string, v []string) { f.addKey(k); f.EncodeTypeStrings(v) }
func (f *textEncoder) EncodeFieldBytes(k string, v []byte)     { f.addKey(k); f.EncodeTypeBytes(v) }
func (f *textEncoder) EncodeFieldInts64(k string, v []int64)   { f.addKey(k); f.EncodeTypeInts64(v) }
func (f *textEncoder) EncodeFieldFloats64(k string, v []float64) {
	f.addKey(k)
	f.EncodeTypeFloats64(v)
}
func (f *textEncoder) EncodeFieldDurations(k string, v []time.Duration) {
	f.addKey(k)
	f.EncodeTypeDurations(v)
}
func (f *textEncoder) EncodeFieldArray(k string, v ArrayEncoder) { f.addKey(k); f.EncodeTypeArray(v) }

func (f *textEncoder) EncodeFieldObject(k string, v ObjectEncoder) {
	if k == "" {
		_ = v.EncodeLogfObject(f)
		return
	}
	f.addKey(k)
	f.EncodeTypeObject(v)
}

func (f *textEncoder) EncodeFieldGroup(k string, fs []Field) {
	if k == "" {
		for _, field := range fs {
			field.Accept(f)
		}
		return
	}
	// Push group prefix, encode fields, pop.
	saved := f.groupPrefix
	f.groupPrefix += k + "."
	for _, field := range fs {
		field.Accept(f)
	}
	f.groupPrefix = saved
}

func (f *textEncoder) EncodeFieldError(k string, v error) {
	f.EncodeError(k, v, f)
}

// --- helpers ---

func (f *textEncoder) appendSeparator() {
	if f.buf.Len() == f.startBufLen {
		return
	}
	if f.buf.Back() == '=' {
		return
	}
	f.buf.AppendByte(' ')
}

func (f *textEncoder) appendTime(t time.Time) {
	start := f.buf.Len()
	f.EncodeTime(t, f)
	end := f.buf.Len()
	// Strip quotes if TimeEncoder added them.
	if end > start && f.buf.Data[start] == '"' && f.buf.Back() == '"' {
		copy(f.buf.Data[start:], f.buf.Data[start+1:end-1])
		f.buf.Data = f.buf.Data[:end-2]
	}
}

func (f *textEncoder) appendLevel(lvl Level) {
	f.eseq.dim(f.buf, func() {
		f.buf.AppendByte('[')
	})
	f.eseq.at2(f.buf, escBold, levelColor(lvl), func() {
		f.EncodeLevel(lvl, f)
	})
	f.eseq.dim(f.buf, func() {
		f.buf.AppendByte(']')
	})
}

func levelColor(lvl Level) escCode {
	switch lvl {
	case LevelDebug:
		return escMagenta
	case LevelInfo:
		return escCyan
	case LevelWarn:
		return escBrightYellow
	case LevelError:
		return escBrightRed
	default:
		return escBrightRed
	}
}

func msgColor(lvl Level) escCode {
	switch lvl {
	case LevelWarn:
		return escBrightYellow
	case LevelError:
		return escBrightRed
	default:
		return escDefault // bold only, terminal default color
	}
}

const escDefault escCode = 0 // no color, used with bold for default text

// --- ANSI escape sequences ---

type escCode int8

const (
	escBold         escCode = 1
	escItalic       escCode = 3
	escReverse      escCode = 7
	escGreen        escCode = 32
	escBlue         escCode = 34
	escMagenta      escCode = 35
	escCyan         escCode = 36
	escWhite        escCode = 37
	escBrightBlack  escCode = 90
	escBrightRed    escCode = 91
	escBrightBlue   escCode = 94
	escBrightCyan   escCode = 96
	escBrightYellow escCode = 93
	escBrightWhite  escCode = 97
)

type escSeq struct{ noColor bool }

// dim emits the muted auxiliary style (time, brackets, separators).
func (es escSeq) dim(buf *Buffer, fn func()) {
	if es.noColor {
		fn()
		return
	}
	buf.AppendString("\x1b[0;2m")
	fn()
	buf.AppendString("\x1b[0m")
}

// dimItalic emits the muted auxiliary style with italic (logger name, caller).
func (es escSeq) dimItalic(buf *Buffer, fn func()) {
	if es.noColor {
		fn()
		return
	}
	buf.AppendString("\x1b[0;2;3m")
	fn()
	buf.AppendString("\x1b[0m")
}

func (es escSeq) at(buf *Buffer, clr escCode, fn func()) {
	if es.noColor {
		fn()
		return
	}
	buf.AppendString("\x1b[")
	buf.AppendInt(int64(clr))
	buf.AppendByte('m')
	fn()
	buf.AppendString("\x1b[0m")
}


func (es escSeq) at2(buf *Buffer, clr1, clr2 escCode, fn func()) {
	if es.noColor {
		fn()
		return
	}
	buf.AppendString("\x1b[")
	buf.AppendInt(int64(clr1))
	buf.AppendByte(';')
	buf.AppendInt(int64(clr2))
	buf.AppendByte('m')
	fn()
	buf.AppendString("\x1b[0m")
}
