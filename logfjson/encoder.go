package logfjson

import (
	"encoding/base64"
	"encoding/json"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/ssgreg/logf"
)

func NewEncoder(c EncoderConfig) logf.Encoder {
	return &Encoder{c.WithDefaults(), logf.NewCache(100), nil, 0}
}

func NewTypeEncoderFactory(c EncoderConfig) logf.TypeEncoderFactory {
	return &Encoder{c.WithDefaults(), nil, nil, 0}
}

type Encoder struct {
	EncoderConfig
	cache *logf.Cache

	buf         *logf.Buffer
	startBufLen int
}

func (f *Encoder) TypeEncoder(buf *logf.Buffer) logf.TypeEncoder {
	f.buf = buf
	f.startBufLen = f.buf.Len()

	return f
}

func (f *Encoder) Encode(buf *logf.Buffer, e logf.Entry) error {
	// TODO: move to clone
	f.buf = buf
	f.startBufLen = f.buf.Len()

	buf.AppendByte('{')

	if f.FieldKeyLevel != "" {
		f.addKey(f.FieldKeyLevel)
		f.EncodeLevel(e.Level, f)
	}
	if f.FieldKeyTime != "" {
		f.EncodeFieldTime(f.FieldKeyTime, e.Time)
	}
	if f.FieldKeyName != "" && e.LoggerName != "" {
		f.EncodeFieldString(f.FieldKeyName, e.LoggerName)
	}
	if f.FieldKeyMsg != "" {
		f.EncodeFieldString(f.FieldKeyMsg, e.Text)
	}
	if f.FieldKeyCaller != "" && e.Caller.Specified {
		f.addKey(f.FieldKeyCaller)
		f.EncodeCaller(e.Caller, f)
	}

	for _, field := range e.Fields {
		field.Accept(f)
	}

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

	buf.AppendByte('}')
	buf.AppendByte('\n')

	return nil
}

func (f *Encoder) EncodeFieldAny(k string, v interface{}) {
	f.addKey(k)
	f.EncodeTypeAny(v)
}

func (f *Encoder) EncodeFieldBool(k string, v bool) {
	f.addKey(k)
	f.EncodeTypeBool(v)
}

func (f *Encoder) EncodeFieldInt64(k string, v int64) {
	f.addKey(k)
	f.EncodeTypeInt64(v)
}

func (f *Encoder) EncodeFieldInt32(k string, v int32) {
	f.addKey(k)
	f.EncodeTypeInt32(v)
}

func (f *Encoder) EncodeFieldInt16(k string, v int16) {
	f.addKey(k)
	f.EncodeTypeInt16(v)
}

func (f *Encoder) EncodeFieldInt8(k string, v int8) {
	f.addKey(k)
	f.EncodeTypeInt8(v)
}

func (f *Encoder) EncodeFieldUint64(k string, v uint64) {
	f.addKey(k)
	f.EncodeTypeUint64(v)
}

func (f *Encoder) EncodeFieldUint32(k string, v uint32) {
	f.addKey(k)
	f.EncodeTypeUint32(v)
}

func (f *Encoder) EncodeFieldUint16(k string, v uint16) {
	f.addKey(k)
	f.EncodeTypeUint16(v)
}

func (f *Encoder) EncodeFieldUint8(k string, v uint8) {
	f.addKey(k)
	f.EncodeTypeUint8(v)
}

func (f *Encoder) EncodeFieldFloat64(k string, v float64) {
	f.addKey(k)
	f.EncodeTypeFloat64(v)
}

func (f *Encoder) EncodeFieldFloat32(k string, v float32) {
	f.addKey(k)
	f.EncodeTypeFloat32(v)
}

func (f *Encoder) EncodeFieldString(k string, v string) {
	f.addKey(k)
	f.EncodeTypeString(v)
}

func (f *Encoder) EncodeFieldDuration(k string, v time.Duration) {
	f.addKey(k)
	f.EncodeTypeDuration(v)
}

func (f *Encoder) EncodeFieldError(k string, v error) {
	// The only exception that has no EncodeX function. EncodeError can add
	// new fields by itself.
	f.EncodeError(k, v, f)
}

func (f *Encoder) EncodeFieldTime(k string, v time.Time) {
	f.addKey(k)
	f.EncodeTypeTime(v)
}

func (f *Encoder) EncodeFieldArray(k string, v logf.ArrayEncoder) {
	f.addKey(k)
	f.EncodeTypeArray(v)
}

func (f *Encoder) EncodeFieldObject(k string, v logf.ObjectEncoder) {
	f.addKey(k)
	f.EncodeTypeObject(v)
}

func (f *Encoder) EncodeFieldBytes(k string, v []byte) {
	f.addKey(k)
	f.EncodeTypeBytes(v)
}

func (f *Encoder) EncodeFieldBools(k string, v []bool) {
	f.addKey(k)
	f.EncodeTypeBools(v)
}

func (f *Encoder) EncodeFieldInts64(k string, v []int64) {
	f.addKey(k)
	f.EncodeTypeInts64(v)
}

func (f *Encoder) EncodeFieldInts32(k string, v []int32) {
	f.addKey(k)
	f.EncodeTypeInts32(v)
}

func (f *Encoder) EncodeFieldInts16(k string, v []int16) {
	f.addKey(k)
	f.EncodeTypeInts16(v)
}

func (f *Encoder) EncodeFieldInts8(k string, v []int8) {
	f.addKey(k)
	f.EncodeTypeInts8(v)
}

func (f *Encoder) EncodeFieldUints64(k string, v []uint64) {
	f.addKey(k)
	f.EncodeTypeUints64(v)
}

func (f *Encoder) EncodeFieldUints32(k string, v []uint32) {
	f.addKey(k)
	f.EncodeTypeUints32(v)
}

func (f *Encoder) EncodeFieldUints16(k string, v []uint16) {
	f.addKey(k)
	f.EncodeTypeUints16(v)
}

func (f *Encoder) EncodeFieldUints8(k string, v []uint8) {
	f.addKey(k)
	f.EncodeTypeUints8(v)
}

func (f *Encoder) EncodeFieldFloats64(k string, v []float64) {
	f.addKey(k)
	f.EncodeTypeFloats64(v)
}

func (f *Encoder) EncodeFieldFloats32(k string, v []float32) {
	f.addKey(k)
	f.EncodeTypeFloats32(v)
}

func (f *Encoder) EncodeFieldDurations(k string, v []time.Duration) {
	f.addKey(k)
	f.EncodeTypeDurations(v)
}

func (f *Encoder) EncodeTypeAny(v interface{}) {
	e := json.NewEncoder(f.buf)
	e.Encode(v)

	if !f.empty() && f.buf.Back() == '\n' {
		f.buf.Data = f.buf.Data[0 : f.buf.Len()-1]
	}
}

func (f *Encoder) EncodeTypeByte(v byte) {
	// TODO: fix as  default marhaller do
	f.appendSeparator()
	f.buf.AppendByte(v)
}

func (f *Encoder) EncodeTypeUnsafeBytes(v unsafe.Pointer) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeByteString(f.buf, *(*[]byte)(v))
	f.buf.AppendByte('"')
}

func (f *Encoder) EncodeTypeBool(v bool) {
	f.appendSeparator()
	logf.AppendBool(f.buf, v)
}

func (f *Encoder) EncodeTypeString(v string) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeString(f.buf, v)
	f.buf.AppendByte('"')
}

func (f *Encoder) EncodeTypeInt64(v int64) {
	f.appendSeparator()
	logf.AppendInt(f.buf, v)
}

func (f *Encoder) EncodeTypeInt32(v int32) {
	f.appendSeparator()
	logf.AppendInt(f.buf, int64(v))
}

func (f *Encoder) EncodeTypeInt16(v int16) {
	f.appendSeparator()
	logf.AppendInt(f.buf, int64(v))
}

func (f *Encoder) EncodeTypeInt8(v int8) {
	f.appendSeparator()
	logf.AppendInt(f.buf, int64(v))
}

func (f *Encoder) EncodeTypeUint64(v uint64) {
	f.appendSeparator()
	logf.AppendUint(f.buf, uint64(v))
}

func (f *Encoder) EncodeTypeUint32(v uint32) {
	f.appendSeparator()
	logf.AppendUint(f.buf, uint64(v))
}

func (f *Encoder) EncodeTypeUint16(v uint16) {
	f.appendSeparator()
	logf.AppendUint(f.buf, uint64(v))
}

func (f *Encoder) EncodeTypeUint8(v uint8) {
	f.appendSeparator()
	logf.AppendUint(f.buf, uint64(v))
}

func (f *Encoder) EncodeTypeFloat64(v float64) {
	f.appendSeparator()
	logf.AppendFloat64(f.buf, v)
}

func (f *Encoder) EncodeTypeFloat32(v float32) {
	f.appendSeparator()
	logf.AppendFloat32(f.buf, v)
}

func (f *Encoder) EncodeTypeDuration(v time.Duration) {
	f.appendSeparator()
	f.EncodeDuration(v, f)
}

func (f *Encoder) EncodeTypeTime(v time.Time) {
	f.appendSeparator()
	f.EncodeTime(v, f)
}

func (f *Encoder) EncodeTypeBytes(v []byte) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	base64.StdEncoding.Encode(f.buf.ExtendBytes(base64.StdEncoding.EncodedLen(len(v))), v)
	f.buf.AppendByte('"')
}

func (f *Encoder) EncodeTypeBools(v []bool) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeBool(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeInts64(v []int64) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeInt64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeInts32(v []int32) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeInt32(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeInts16(v []int16) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeInt16(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeInts8(v []int8) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeInt8(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeUints64(v []uint64) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeUint64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeUints32(v []uint32) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeUint32(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeUints16(v []uint16) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeUint16(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeUints8(v []uint8) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeUint8(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeFloats64(v []float64) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeFloat64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeFloats32(v []float32) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeFloat32(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeDurations(v []time.Duration) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	for i := range v {
		f.EncodeTypeDuration(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeArray(v logf.ArrayEncoder) {
	f.appendSeparator()
	f.buf.AppendByte('[')
	v.EncodeLogfArray(f)
	f.buf.AppendByte(']')
}

func (f *Encoder) EncodeTypeObject(v logf.ObjectEncoder) {
	f.appendSeparator()
	f.buf.AppendByte('{')
	v.EncodeLogfObject(f)
	f.buf.AppendByte('}')
}

func (f *Encoder) appendSeparator() {
	if f.empty() {
		return
	}

	switch f.buf.Back() {
	case '{', '[', ':', ',':
		return
	}
	f.buf.AppendByte(',')
}

func (f *Encoder) empty() bool {
	return f.buf.Len() == f.startBufLen
}

func (f *Encoder) addKey(k string) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeString(f.buf, k)
	f.buf.AppendByte('"')
	f.buf.AppendByte(':')
}

const hex = "0123456789abcdef"

func EscapeString(buf *logf.Buffer, s string) error {
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
		switch v {
		case utf8.RuneError:
			if wd == 1 {
				buf.AppendString(s[p:i])
				buf.AppendString(`\ufffd`)
				p = i
			}
		case '\u2028', '\u2029':
			buf.AppendString(s[p:i])
			buf.AppendString(`\u202`)
			buf.AppendByte(hex[v&0xf])
			p = i
		}
		i += wd
	}
	buf.AppendString(s[p:])

	return nil
}

func EscapeByteString(buf *logf.Buffer, s []byte) error {
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
