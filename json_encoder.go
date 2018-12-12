package logf

import (
	"encoding/base64"
	"encoding/json"
	"time"
	"unicode/utf8"
	"unsafe"
)

// NewJSONEncoder creates the new instance of the JSON Encoder with the
// given JSONEncoderConfig.
var NewJSONEncoder = jsonEncoderGetter(
	func(cfg JSONEncoderConfig) Encoder {
		return &jsonEncoder{cfg.WithDefaults(), NewCache(100), nil, 0}
	},
)

// NewJSONTypeEncoderFactory creates the new instance of the JSON
// TypeEncoderFactory with the given JSONEncoderConfig.
var NewJSONTypeEncoderFactory = jsonTypeEncoderFactoryGetter(
	func(c JSONEncoderConfig) TypeEncoderFactory {
		return &jsonEncoder{c.WithDefaults(), nil, nil, 0}
	},
)

type jsonEncoderGetter func(cfg JSONEncoderConfig) Encoder

func (c jsonEncoderGetter) Default() Encoder {
	return c(JSONEncoderConfig{})
}

type jsonTypeEncoderFactoryGetter func(cfg JSONEncoderConfig) TypeEncoderFactory

func (c jsonTypeEncoderFactoryGetter) Default() TypeEncoderFactory {
	return c(JSONEncoderConfig{})
}

type jsonEncoder struct {
	JSONEncoderConfig
	cache *Cache

	buf         *Buffer
	startBufLen int
}

func (f *jsonEncoder) TypeEncoder(buf *Buffer) TypeEncoder {
	f.buf = buf
	f.startBufLen = f.buf.Len()

	return f
}

func (f *jsonEncoder) Encode(buf *Buffer, e Entry) error {
	// TODO: move to clone
	f.buf = buf
	f.startBufLen = f.buf.Len()

	buf.AppendByte('{')

	// Level.
	if !f.DisableFieldLevel {
		f.addKey(f.FieldKeyLevel)
		f.EncodeLevel(e.Level, f)
	}

	// Time.
	if !f.DisableFieldTime {
		f.EncodeFieldTime(f.FieldKeyTime, e.Time)
	}

	// Logger name.
	if !f.DisableFieldName && e.LoggerName != "" {
		f.EncodeFieldString(f.FieldKeyName, e.LoggerName)
	}

	// Message.
	if !f.DisableFieldMsg {
		f.EncodeFieldString(f.FieldKeyMsg, e.Text)
	}

	// Caller.
	if !f.DisableFieldCaller && e.Caller.Specified {
		f.addKey(f.FieldKeyCaller)
		f.EncodeCaller(e.Caller, f)
	}

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

	buf.AppendByte('}')
	buf.AppendByte('\n')

	return nil
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

func (f *jsonEncoder) EncodeFieldInt32(k string, v int32) {
	f.addKey(k)
	f.EncodeTypeInt32(v)
}

func (f *jsonEncoder) EncodeFieldInt16(k string, v int16) {
	f.addKey(k)
	f.EncodeTypeInt16(v)
}

func (f *jsonEncoder) EncodeFieldInt8(k string, v int8) {
	f.addKey(k)
	f.EncodeTypeInt8(v)
}

func (f *jsonEncoder) EncodeFieldUint64(k string, v uint64) {
	f.addKey(k)
	f.EncodeTypeUint64(v)
}

func (f *jsonEncoder) EncodeFieldUint32(k string, v uint32) {
	f.addKey(k)
	f.EncodeTypeUint32(v)
}

func (f *jsonEncoder) EncodeFieldUint16(k string, v uint16) {
	f.addKey(k)
	f.EncodeTypeUint16(v)
}

func (f *jsonEncoder) EncodeFieldUint8(k string, v uint8) {
	f.addKey(k)
	f.EncodeTypeUint8(v)
}

func (f *jsonEncoder) EncodeFieldFloat64(k string, v float64) {
	f.addKey(k)
	f.EncodeTypeFloat64(v)
}

func (f *jsonEncoder) EncodeFieldFloat32(k string, v float32) {
	f.addKey(k)
	f.EncodeTypeFloat32(v)
}

func (f *jsonEncoder) EncodeFieldString(k string, v string) {
	f.addKey(k)
	f.EncodeTypeString(v)
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

func (f *jsonEncoder) EncodeFieldBytes(k string, v []byte) {
	f.addKey(k)
	f.EncodeTypeBytes(v)
}

func (f *jsonEncoder) EncodeFieldBools(k string, v []bool) {
	f.addKey(k)
	f.EncodeTypeBools(v)
}

func (f *jsonEncoder) EncodeFieldInts64(k string, v []int64) {
	f.addKey(k)
	f.EncodeTypeInts64(v)
}

func (f *jsonEncoder) EncodeFieldInts32(k string, v []int32) {
	f.addKey(k)
	f.EncodeTypeInts32(v)
}

func (f *jsonEncoder) EncodeFieldInts16(k string, v []int16) {
	f.addKey(k)
	f.EncodeTypeInts16(v)
}

func (f *jsonEncoder) EncodeFieldInts8(k string, v []int8) {
	f.addKey(k)
	f.EncodeTypeInts8(v)
}

func (f *jsonEncoder) EncodeFieldUints64(k string, v []uint64) {
	f.addKey(k)
	f.EncodeTypeUints64(v)
}

func (f *jsonEncoder) EncodeFieldUints32(k string, v []uint32) {
	f.addKey(k)
	f.EncodeTypeUints32(v)
}

func (f *jsonEncoder) EncodeFieldUints16(k string, v []uint16) {
	f.addKey(k)
	f.EncodeTypeUints16(v)
}

func (f *jsonEncoder) EncodeFieldUints8(k string, v []uint8) {
	f.addKey(k)
	f.EncodeTypeUints8(v)
}

func (f *jsonEncoder) EncodeFieldFloats64(k string, v []float64) {
	f.addKey(k)
	f.EncodeTypeFloats64(v)
}

func (f *jsonEncoder) EncodeFieldFloats32(k string, v []float32) {
	f.addKey(k)
	f.EncodeTypeFloats32(v)
}

func (f *jsonEncoder) EncodeFieldDurations(k string, v []time.Duration) {
	f.addKey(k)
	f.EncodeTypeDurations(v)
}

func (f *jsonEncoder) EncodeTypeAny(v interface{}) {
	e := json.NewEncoder(f.buf)
	e.Encode(v)

	if !f.empty() && f.buf.Back() == '\n' {
		f.buf.Data = f.buf.Data[0 : f.buf.Len()-1]
	}
}

func (f *jsonEncoder) EncodeTypeUnsafeBytes(v unsafe.Pointer) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeByteString(f.buf, *(*[]byte)(v))
	f.buf.AppendByte('"')
}

func (f *jsonEncoder) EncodeTypeBool(v bool) {
	f.appendSeparator()
	AppendBool(f.buf, v)
}

func (f *jsonEncoder) EncodeTypeString(v string) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeString(f.buf, v)
	f.buf.AppendByte('"')
}

func (f *jsonEncoder) EncodeTypeInt64(v int64) {
	f.appendSeparator()
	AppendInt(f.buf, v)
}

func (f *jsonEncoder) EncodeTypeInt32(v int32) {
	f.appendSeparator()
	AppendInt(f.buf, int64(v))
}

func (f *jsonEncoder) EncodeTypeInt16(v int16) {
	f.appendSeparator()
	AppendInt(f.buf, int64(v))
}

func (f *jsonEncoder) EncodeTypeInt8(v int8) {
	f.appendSeparator()
	AppendInt(f.buf, int64(v))
}

func (f *jsonEncoder) EncodeTypeUint64(v uint64) {
	f.appendSeparator()
	AppendUint(f.buf, uint64(v))
}

func (f *jsonEncoder) EncodeTypeUint32(v uint32) {
	f.appendSeparator()
	AppendUint(f.buf, uint64(v))
}

func (f *jsonEncoder) EncodeTypeUint16(v uint16) {
	f.appendSeparator()
	AppendUint(f.buf, uint64(v))
}

func (f *jsonEncoder) EncodeTypeUint8(v uint8) {
	f.appendSeparator()
	AppendUint(f.buf, uint64(v))
}

func (f *jsonEncoder) EncodeTypeFloat64(v float64) {
	f.appendSeparator()
	AppendFloat64(f.buf, v)
}

func (f *jsonEncoder) EncodeTypeFloat32(v float32) {
	f.appendSeparator()
	AppendFloat32(f.buf, v)
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

func (f *jsonEncoder) EncodeTypeBools(v []bool) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeBool(v[i])
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

func (f *jsonEncoder) EncodeTypeInts32(v []int32) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeInt32(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeInts16(v []int16) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeInt16(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeInts8(v []int8) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeInt8(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeUints64(v []uint64) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeUint64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeUints32(v []uint32) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeUint32(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeUints16(v []uint16) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeUint16(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeUints8(v []uint8) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeUint8(v[i])
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

func (f *jsonEncoder) EncodeTypeFloats32(v []float32) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeFloat32(v[i])
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
	v.EncodeLogfArray(f)
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) EncodeTypeObject(v ObjectEncoder) {
	f.appendSeparator()
	f.buf.AppendByte('{')
	v.EncodeLogfObject(f)
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

func (f *jsonEncoder) addKey(k string) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeString(f.buf, k)
	f.buf.AppendByte('"')
	f.buf.AppendByte(':')
}

const hex = "0123456789abcdef"

// EscapeString processes a single escape sequence to the given Buffer.
func EscapeString(buf *Buffer, s string) error {
	p := 0
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c < utf8.RuneSelf && c >= 0x20 && c != '\\' && c != '"':
			i++

			continue

		case c < utf8.RuneSelf:
			buf.AppendString(s[p:i])
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
		v, wd := utf8.DecodeRuneInString(s[i:])
		if v == utf8.RuneError && wd == 1 {
			buf.AppendString(s[p:i])
			buf.AppendString(`\ufffd`)
			i++
			p = i

			continue
		} else {
			i += wd
		}
	}
	buf.AppendString(s[p:])

	return nil
}

// EscapeByteString processes a single escape sequence to the given Buffer.
// TODO: use EscapeString instead
func EscapeByteString(buf *Buffer, s []byte) error {
	p := 0
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c >= 0x20 && c != '\\' && c != '"':
			i++

			continue

		default:
			buf.AppendBytes(s[p:i])
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
	}
	buf.AppendBytes(s[p:])

	return nil
}
