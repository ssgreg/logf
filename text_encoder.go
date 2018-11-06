package logf

import (
	"time"
)

// NewTextEncoder creates the new instance of the text Encoder with the
// given TextEncoderConfig. It delegates field encoding to the internal
// json Encoder.
var NewTextEncoder = textEncoderGetter(
	func(cfg TextEncoderConfig) Encoder {
		return &textEncoder{
			cfg.WithDefaults(),
			NewJSONTypeEncoderFactory(JSONEncoderConfig{
				EncodeTime:     cfg.EncodeTime,
				EncodeDuration: cfg.EncodeDuration,
				EncodeError:    cfg.EncodeError,
			}),
			nil,
			NewCache(100),
			0,
			false,
		}
	},
)

type textEncoderGetter func(cfg TextEncoderConfig) Encoder

func (c textEncoderGetter) Default() Encoder {
	return c(TextEncoderConfig{})
}

type textEncoder struct {
	TextEncoderConfig
	mf TypeEncoderFactory

	buf         *Buffer
	cache       *Cache
	startBufLen int
	isField     bool
}

func (f *textEncoder) Encode(buf *Buffer, e Entry) error {
	// TODO: move to clone
	f.buf = buf
	f.startBufLen = f.buf.Len()

	// Time.
	AtEscapeSequence(f.buf, EscBrightBlack, func() {
		appendTime(e.Time, f.buf, f.EncodeTime, f.mf.TypeEncoder(buf))
	})

	// Level.
	f.appendSeparator()
	appendLevel(buf, e.Level)

	// Logger name.
	if !f.DisableFieldName && e.LoggerName != "" {
		f.appendSeparator()
		AtEscapeSequence(f.buf, EscBrightBlack, func() {
			f.buf.AppendString(e.LoggerName)
			f.buf.AppendByte(':')
		})
	}

	// Message.
	f.appendSeparator()
	AtEscapeSequence(f.buf, EscBrightWhite, func() {
		f.buf.AppendString(e.Text)
	})

	// Logger's fields.
	if bytes, ok := f.cache.Get(e.LoggerID); ok {
		buf.AppendBytes(bytes)
	} else {
		le := buf.Len()
		for _, field := range e.DerivedFields {
			field.Accept(f)
		}

		bf := make([]byte, buf.Len()-le)
		copy(bf, buf.Data[le:])
		f.cache.Set(e.LoggerID, bf)
	}

	// Entry's fields.
	for _, field := range e.Fields {
		field.Accept(f)
	}

	// Caller.
	if !f.DisableFieldCaller && e.Caller.Specified {
		AtEscapeSequence(f.buf, EscBrightBlack, func() {
			f.appendSeparator()
			f.buf.AppendByte('@')
			f.EncodeCaller(e.Caller, f.mf.TypeEncoder(f.buf))
		})
	}

	buf.AppendByte('\n')

	return nil
}

func (f *textEncoder) EncodeFieldAny(k string, v interface{}) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeAny(v)
}

func (f *textEncoder) EncodeFieldBool(k string, v bool) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeBool(v)
}

func (f *textEncoder) EncodeFieldInt64(k string, v int64) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeInt64(v)
}

func (f *textEncoder) EncodeFieldInt32(k string, v int32) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeInt32(v)
}

func (f *textEncoder) EncodeFieldInt16(k string, v int16) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeInt16(v)
}

func (f *textEncoder) EncodeFieldInt8(k string, v int8) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeInt8(v)
}

func (f *textEncoder) EncodeFieldUint64(k string, v uint64) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeUint64(v)
}

func (f *textEncoder) EncodeFieldUint32(k string, v uint32) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeUint32(v)
}

func (f *textEncoder) EncodeFieldUint16(k string, v uint16) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeUint16(v)
}

func (f *textEncoder) EncodeFieldUint8(k string, v uint8) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeUint8(v)
}

func (f *textEncoder) EncodeFieldFloat64(k string, v float64) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeFloat64(v)
}

func (f *textEncoder) EncodeFieldFloat32(k string, v float32) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeFloat32(v)
}

func (f *textEncoder) EncodeFieldString(k string, v string) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeString(v)
}

func (f *textEncoder) EncodeFieldDuration(k string, v time.Duration) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeDuration(v)
}

func (f *textEncoder) EncodeFieldError(k string, v error) {
	f.EncodeError(k, v, f)
}

func (f *textEncoder) EncodeFieldTime(k string, v time.Time) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeTime(v)
}

func (f *textEncoder) EncodeFieldArray(k string, v ArrayEncoder) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeArray(v)
}

func (f *textEncoder) EncodeFieldObject(k string, v ObjectEncoder) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeObject(v)
}

func (f *textEncoder) EncodeFieldBytes(k string, v []byte) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeBytes(v)
}

func (f *textEncoder) EncodeFieldBools(k string, v []bool) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeBools(v)
}

func (f *textEncoder) EncodeFieldInts64(k string, v []int64) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeInts64(v)
}

func (f *textEncoder) EncodeFieldInts32(k string, v []int32) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeInts32(v)
}

func (f *textEncoder) EncodeFieldInts16(k string, v []int16) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeInts16(v)
}

func (f *textEncoder) EncodeFieldInts8(k string, v []int8) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeInts8(v)
}

func (f *textEncoder) EncodeFieldUints64(k string, v []uint64) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeUints64(v)
}

func (f *textEncoder) EncodeFieldUints32(k string, v []uint32) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeUints32(v)
}

func (f *textEncoder) EncodeFieldUints16(k string, v []uint16) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeUints16(v)
}

func (f *textEncoder) EncodeFieldUints8(k string, v []uint8) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeUints8(v)
}

func (f *textEncoder) EncodeFieldFloats64(k string, v []float64) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeFloats64(v)
}

func (f *textEncoder) EncodeFieldFloats32(k string, v []float32) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeFloats32(v)
}

func (f *textEncoder) EncodeFieldDurations(k string, v []time.Duration) {
	f.addKey(k)
	f.mf.TypeEncoder(f.buf).EncodeTypeDurations(v)
}

func (f *textEncoder) appendSeparator() {
	if f.empty() {
		return
	}

	switch f.buf.Back() {
	case '=':
		return
	}
	f.buf.AppendByte(' ')
}

func (f *textEncoder) empty() bool {
	return f.buf.Len() == f.startBufLen
}

func (f *textEncoder) addKey(k string) {
	f.appendSeparator()
	AtEscapeSequence(f.buf, EscGreen, func() {
		f.buf.AppendString(k)
	})

	AtEscapeSequence(f.buf, EscBrightBlack, func() {
		f.buf.AppendByte('=')
	})
}

func (f *textEncoder) setIsField(isField bool) {
	f.isField = isField
}

func appendLevel(buf *Buffer, lvl Level) {
	buf.AppendByte('|')

	switch lvl {
	case LevelDebug:
		AtEscapeSequence(buf, EscMagenta, func() {
			buf.AppendString("DEBU")
		})
	case LevelInfo:
		AtEscapeSequence(buf, EscCyan, func() {
			buf.AppendString("INFO")
		})
	case LevelWarn:
		AtEscapeSequence2(buf, EscBrightYellow, EscReverse, func() {
			buf.AppendString("WARN")
		})
	case LevelError:
		AtEscapeSequence2(buf, EscBrightRed, EscReverse, func() {
			buf.AppendString("ERRO")
		})
	default:
		AtEscapeSequence(buf, EscBrightRed, func() {
			buf.AppendString("UNKN")
		})
	}

	buf.AppendByte('|')
}

func appendTime(t time.Time, buf *Buffer, enc TimeEncoder, encType TypeEncoder) {
	start := buf.Len()
	enc(t, encType)
	end := buf.Len()

	// Get rid of possible quotes.
	if end != start {
		if buf.Data[start] == '"' && buf.Back() == '"' {
			copy(buf.Data[start:end], buf.Data[start+1:end-2])
			buf.Data = buf.Data[0 : end-2]
		}
	}
}
