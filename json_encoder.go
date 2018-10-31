package logf

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
	"unicode/utf8"
	"unsafe"
)

func NewJSONEncoder(c *FormatterConfig) Encoder {
	return &jsonEncoder{c, NewCache(100), nil}
}

func NewJSONTypeMarshallerFactory(c *FormatterConfig) TypeMarshallerFactory {
	return &jsonEncoder{c, nil, nil}
}

type jsonEncoder struct {
	*FormatterConfig
	cache *Cache

	buf *Buffer
}

func (f *jsonEncoder) TypeMarshaller(buf *Buffer) TypeMarshaller {
	f.buf = buf
	return f
}

func (f *jsonEncoder) MarshalFieldAny(k string, v interface{}) {
	f.addKey(k)
	f.MarshalAny(v)
}

func (f *jsonEncoder) MarshalFieldBool(k string, v bool) {
	f.addKey(k)
	f.MarshalBool(v)
}

func (f *jsonEncoder) MarshalFieldInt64(k string, v int64) {
	f.addKey(k)
	f.MarshalInt64(v)
}

func (f *jsonEncoder) MarshalFieldInt32(k string, v int32) {
	f.addKey(k)
	f.MarshalInt32(v)
}

func (f *jsonEncoder) MarshalFieldInt16(k string, v int16) {
	f.addKey(k)
	f.MarshalInt16(v)
}

func (f *jsonEncoder) MarshalFieldInt8(k string, v int8) {
	f.addKey(k)
	f.MarshalInt8(v)
}

func (f *jsonEncoder) MarshalFieldUint64(k string, v uint64) {
	f.addKey(k)
	f.MarshalUint64(v)
}

func (f *jsonEncoder) MarshalFieldUint32(k string, v uint32) {
	f.addKey(k)
	f.MarshalUint32(v)
}

func (f *jsonEncoder) MarshalFieldUint16(k string, v uint16) {
	f.addKey(k)
	f.MarshalUint16(v)
}

func (f *jsonEncoder) MarshalFieldUint8(k string, v uint8) {
	f.addKey(k)
	f.MarshalUint8(v)
}

func (f *jsonEncoder) MarshalFieldFloat64(k string, v float64) {
	f.addKey(k)
	f.MarshalFloat64(v)
}

func (f *jsonEncoder) MarshalFieldFloat32(k string, v float32) {
	f.addKey(k)
	f.MarshalFloat32(v)
}

func (f *jsonEncoder) MarshalFieldString(k string, v string) {
	f.addKey(k)
	f.MarshalString(v)
}

func (f *jsonEncoder) MarshalFieldDuration(k string, v time.Duration) {
	f.addKey(k)
	f.MarshalDuration(v)
}

func (f *jsonEncoder) MarshalFieldError(k string, v error) {
	// The only exception that has no MarshalX function. FormatError can add
	// new fields by itself.
	f.FormatError(k, v, f)
}

func (f *jsonEncoder) MarshalFieldTime(k string, v time.Time) {
	f.addKey(k)
	f.MarshalTime(v)
}

func (f *jsonEncoder) MarshalFieldArray(k string, v ArrayMarshaller) {
	f.addKey(k)
	f.MarshalArray(v)
}

func (f *jsonEncoder) MarshalFieldObject(k string, v ObjectMarshaller) {
	f.addKey(k)
	f.MarshalObject(v)
}

func (f *jsonEncoder) MarshalFieldBytes(k string, v []byte) {
	f.addKey(k)
	f.MarshalBytes(v)
}

func (f *jsonEncoder) MarshalFieldBools(k string, v []bool) {
	f.addKey(k)
	f.MarshalBools(v)
}

func (f *jsonEncoder) MarshalFieldInts64(k string, v []int64) {
	f.addKey(k)
	f.MarshalInts64(v)
}

func (f *jsonEncoder) MarshalFieldInts32(k string, v []int32) {
	f.addKey(k)
	f.MarshalInts32(v)
}

func (f *jsonEncoder) MarshalFieldInts16(k string, v []int16) {
	f.addKey(k)
	f.MarshalInts16(v)
}

func (f *jsonEncoder) MarshalFieldInts8(k string, v []int8) {
	f.addKey(k)
	f.MarshalInts8(v)
}

func (f *jsonEncoder) MarshalFieldUints64(k string, v []uint64) {
	f.addKey(k)
	f.MarshalUints64(v)
}

func (f *jsonEncoder) MarshalFieldUints32(k string, v []uint32) {
	f.addKey(k)
	f.MarshalUints32(v)
}

func (f *jsonEncoder) MarshalFieldUints16(k string, v []uint16) {
	f.addKey(k)
	f.MarshalUints16(v)
}

func (f *jsonEncoder) MarshalFieldUints8(k string, v []uint8) {
	f.addKey(k)
	f.MarshalUints8(v)
}

func (f *jsonEncoder) MarshalFieldFloats64(k string, v []float64) {
	f.addKey(k)
	f.MarshalFloats64(v)
}

func (f *jsonEncoder) MarshalFieldFloats32(k string, v []float32) {
	f.addKey(k)
	f.MarshalFloats32(v)
}

func (f *jsonEncoder) MarshalFieldDurations(k string, v []time.Duration) {
	f.addKey(k)
	f.MarshalDurations(v)
}

func (f *jsonEncoder) MarshalAny(v interface{}) {
	if !KnownTypeToBuf(f.buf, v) {
		e := json.NewEncoder(f.buf)
		e.Encode(v)
	}
}

func (f *jsonEncoder) MarshalByte(v byte) {
	// TODO: fix as json default marhaller do
	f.appendSeparator()
	f.buf.AppendByte(v)
}

func (f *jsonEncoder) MarshalUnsafeBytes(v unsafe.Pointer) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeByteString(f.buf, *(*[]byte)(v))
	f.buf.AppendByte('"')
}

func (f *jsonEncoder) MarshalBool(v bool) {
	f.appendSeparator()
	AppendBool(f.buf, v)
}

func (f *jsonEncoder) MarshalString(v string) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeString(f.buf, v)
	f.buf.AppendByte('"')
}

func (f *jsonEncoder) MarshalInt64(v int64) {
	f.appendSeparator()
	AppendInt(f.buf, v)
}

func (f *jsonEncoder) MarshalInt32(v int32) {
	f.appendSeparator()
	AppendInt(f.buf, int64(v))
}

func (f *jsonEncoder) MarshalInt16(v int16) {
	f.appendSeparator()
	AppendInt(f.buf, int64(v))
}

func (f *jsonEncoder) MarshalInt8(v int8) {
	f.appendSeparator()
	AppendInt(f.buf, int64(v))
}

func (f *jsonEncoder) MarshalUint64(v uint64) {
	f.appendSeparator()
	AppendUint(f.buf, uint64(v))
}

func (f *jsonEncoder) MarshalUint32(v uint32) {
	f.appendSeparator()
	AppendUint(f.buf, uint64(v))
}

func (f *jsonEncoder) MarshalUint16(v uint16) {
	f.appendSeparator()
	AppendUint(f.buf, uint64(v))
}

func (f *jsonEncoder) MarshalUint8(v uint8) {
	f.appendSeparator()
	AppendUint(f.buf, uint64(v))
}

func (f *jsonEncoder) MarshalFloat64(v float64) {
	f.appendSeparator()
	AppendFloat64(f.buf, v)
}

func (f *jsonEncoder) MarshalFloat32(v float32) {
	f.appendSeparator()
	AppendFloat32(f.buf, v)
}

func (f *jsonEncoder) MarshalDuration(v time.Duration) {
	f.appendSeparator()
	f.FormatDuration(v, f)
}

func (f *jsonEncoder) MarshalTime(v time.Time) {
	f.appendSeparator()
	f.FormatTime(v, f)
}

func (f *jsonEncoder) MarshalBytes(v []byte) {
	f.buf.AppendByte('"')
	base64.StdEncoding.Encode(f.buf.ExtendBytes(base64.StdEncoding.EncodedLen(len(v))), v)
	f.buf.AppendByte('"')
}

func (f *jsonEncoder) MarshalBools(v []bool) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalBool(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalInts64(v []int64) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalInt64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalInts32(v []int32) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalInt32(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalInts16(v []int16) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalInt16(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalInts8(v []int8) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalInt8(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalUints64(v []uint64) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalUint64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalUints32(v []uint32) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalUint32(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalUints16(v []uint16) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalUint16(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalUints8(v []uint8) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalUint8(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalFloats64(v []float64) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalFloat64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalFloats32(v []float32) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalFloat32(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalDurations(v []time.Duration) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalDuration(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalArray(v ArrayMarshaller) {
	f.buf.AppendByte('[')
	v.MarshalLogfArray(f)
	f.buf.AppendByte(']')
}

func (f *jsonEncoder) MarshalObject(v ObjectMarshaller) {
	f.buf.AppendByte('{')
	v.MarshalLogfObject(f)
	f.buf.AppendByte('}')
}

func (f *jsonEncoder) appendSeparator() {
	if f.buf.Len() == 0 {
		return
	}
	switch f.buf.Back() {
	case '{', '[', ':', ',':
		return
	}
	f.buf.AppendByte(',')
}

func (f *jsonEncoder) addKey(k string) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeString(f.buf, k)
	f.buf.AppendByte('"')
	f.buf.AppendByte(':')
}

func (f *jsonEncoder) Encode(buf *Buffer, e Entry) error {
	f.buf = buf

	buf.AppendByte('{')

	if f.FieldKeyLevel != "" {
		f.MarshalFieldString(f.FieldKeyLevel, e.Level.String())
	}
	if f.FieldKeyTime != "" {
		f.MarshalFieldTime(f.FieldKeyTime, e.Time)
	}
	if f.FieldKeyName != "" && e.LoggerName != "" {
		f.MarshalFieldString(f.FieldKeyName, e.LoggerName)
	}
	if f.FieldKeyMsg != "" {
		f.MarshalFieldString(f.FieldKeyMsg, e.Text)
	}
	if f.FieldKeyCaller != "" && e.Caller.Specified {
		f.addKey(f.FieldKeyCaller)
		f.FormatCaller(e.Caller, f)
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

const hex = "0123456789abcdef"

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

func KnownTypeToBuf(buf *Buffer, v interface{}) bool {
	switch rv := v.(type) {
	case string:
		EscapeString(buf, rv)
	case bool:
		AppendBool(buf, rv)
	case int:
		AppendInt(buf, int64(rv))
	case int8:
		AppendInt(buf, int64(rv))
	case int16:
		AppendInt(buf, int64(rv))
	case int32:
		AppendInt(buf, int64(rv))
	case int64:
		AppendInt(buf, int64(rv))
	case uint:
		AppendUint(buf, uint64(rv))
	case uint8:
		AppendUint(buf, uint64(rv))
	case uint16:
		AppendUint(buf, uint64(rv))
	case uint32:
		AppendUint(buf, uint64(rv))
	case uint64:
		AppendUint(buf, uint64(rv))
	case float32:
		AppendFloat32(buf, rv)
	case float64:
		AppendFloat64(buf, rv)
	case fmt.Stringer:
		EscapeString(buf, rv.String())
	case error:
		EscapeString(buf, rv.Error())
	default:
		if rv == nil {
			return false
		}
		switch reflect.TypeOf(rv).Kind() {
		case reflect.String:
			EscapeString(buf, reflect.ValueOf(rv).String())
		case reflect.Bool:
			AppendBool(buf, reflect.ValueOf(rv).Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			AppendInt(buf, reflect.ValueOf(rv).Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			AppendUint(buf, reflect.ValueOf(rv).Uint())
		case reflect.Float32, reflect.Float64:
			AppendFloat64(buf, reflect.ValueOf(rv).Float())
		default:
			return false
		}
	}

	return true
}
