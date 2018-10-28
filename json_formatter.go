package logf

import (
	"encoding/base64"
	"encoding/json"
	"time"
	"unsafe"
)

func NewJSONFormatter(c *FormatterConfig) Formatter {
	f := &jsonFormatter{c, nil, NewCache(100)}
	if f.FormatDuration == nil {
		f.FormatDuration = StringDurationFormatter
	}
	if f.FormatTime == nil {
		f.FormatTime = RFC3339TimeFormatter
	}
	if f.FormatError == nil {
		f.FormatError = DefaultErrorFormatter
	}
	if f.FormatCaller == nil {
		f.FormatCaller = ShortCallerFormatter
	}

	return f
}

type jsonFormatter struct {
	*FormatterConfig

	buf   *Buffer
	cache *Cache
}

func (f *jsonFormatter) MarshalFieldAny(k string, v interface{}) {
	f.addKey(k)
	f.MarshalAny(v)
}

func (f *jsonFormatter) MarshalFieldBool(k string, v bool) {
	f.addKey(k)
	f.MarshalBool(v)
}

func (f *jsonFormatter) MarshalFieldInt64(k string, v int64) {
	f.addKey(k)
	f.MarshalInt64(v)
}

func (f *jsonFormatter) MarshalFieldInt32(k string, v int32) {
	f.addKey(k)
	f.MarshalInt32(v)
}

func (f *jsonFormatter) MarshalFieldInt16(k string, v int16) {
	f.addKey(k)
	f.MarshalInt16(v)
}

func (f *jsonFormatter) MarshalFieldInt8(k string, v int8) {
	f.addKey(k)
	f.MarshalInt8(v)
}

func (f *jsonFormatter) MarshalFieldUint64(k string, v uint64) {
	f.addKey(k)
	f.MarshalUint64(v)
}

func (f *jsonFormatter) MarshalFieldUint32(k string, v uint32) {
	f.addKey(k)
	f.MarshalUint32(v)
}

func (f *jsonFormatter) MarshalFieldUint16(k string, v uint16) {
	f.addKey(k)
	f.MarshalUint16(v)
}

func (f *jsonFormatter) MarshalFieldUint8(k string, v uint8) {
	f.addKey(k)
	f.MarshalUint8(v)
}

func (f *jsonFormatter) MarshalFieldFloat64(k string, v float64) {
	f.addKey(k)
	f.MarshalFloat64(v)
}

func (f *jsonFormatter) MarshalFieldFloat32(k string, v float32) {
	f.addKey(k)
	f.MarshalFloat32(v)
}

func (f *jsonFormatter) MarshalFieldString(k string, v string) {
	f.addKey(k)
	f.MarshalString(v)
}

func (f *jsonFormatter) MarshalFieldDuration(k string, v time.Duration) {
	f.addKey(k)
	f.MarshalDuration(v)
}

func (f *jsonFormatter) MarshalFieldError(k string, v error) {
	// The only exception that has no MarshalX function. FormatError can add
	// new fields by itself.
	f.FormatError(k, v, f)
}

func (f *jsonFormatter) MarshalFieldTime(k string, v time.Time) {
	f.addKey(k)
	f.MarshalTime(v)
}

func (f *jsonFormatter) MarshalFieldArray(k string, v ArrayMarshaller) {
	f.addKey(k)
	f.MarshalArray(v)
}

func (f *jsonFormatter) MarshalFieldObject(k string, v ObjectMarshaller) {
	f.addKey(k)
	f.MarshalObject(v)
}

func (f *jsonFormatter) MarshalFieldBytes(k string, v []byte) {
	f.addKey(k)
	f.MarshalBytes(v)
}

func (f *jsonFormatter) MarshalFieldBools(k string, v []bool) {
	f.addKey(k)
	f.MarshalBools(v)
}

func (f *jsonFormatter) MarshalFieldInts64(k string, v []int64) {
	f.addKey(k)
	f.MarshalInts64(v)
}

func (f *jsonFormatter) MarshalFieldInts32(k string, v []int32) {
	f.addKey(k)
	f.MarshalInts32(v)
}

func (f *jsonFormatter) MarshalFieldInts16(k string, v []int16) {
	f.addKey(k)
	f.MarshalInts16(v)
}

func (f *jsonFormatter) MarshalFieldInts8(k string, v []int8) {
	f.addKey(k)
	f.MarshalInts8(v)
}

func (f *jsonFormatter) MarshalFieldUints64(k string, v []uint64) {
	f.addKey(k)
	f.MarshalUints64(v)
}

func (f *jsonFormatter) MarshalFieldUints32(k string, v []uint32) {
	f.addKey(k)
	f.MarshalUints32(v)
}

func (f *jsonFormatter) MarshalFieldUints16(k string, v []uint16) {
	f.addKey(k)
	f.MarshalUints16(v)
}

func (f *jsonFormatter) MarshalFieldUints8(k string, v []uint8) {
	f.addKey(k)
	f.MarshalUints8(v)
}

func (f *jsonFormatter) MarshalFieldFloats64(k string, v []float64) {
	f.addKey(k)
	f.MarshalFloats64(v)
}

func (f *jsonFormatter) MarshalFieldFloats32(k string, v []float32) {
	f.addKey(k)
	f.MarshalFloats32(v)
}

func (f *jsonFormatter) MarshalFieldDurations(k string, v []time.Duration) {
	f.addKey(k)
	f.MarshalDurations(v)
}

func (f *jsonFormatter) MarshalAny(v interface{}) {
	if !KnownTypeToBuf(f.buf, v) {
		e := json.NewEncoder(f.buf)
		e.Encode(v)
	}
}

func (f *jsonFormatter) MarshalByte(v byte) {
	// TODO: fix as json default marhaller do
	f.appendSeparator()
	f.buf.AppendByte(v)
}

func (f *jsonFormatter) MarshalUnsafeBytes(v unsafe.Pointer) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeStringBytes(f.buf, *(*[]byte)(v))
	f.buf.AppendByte('"')
}

func (f *jsonFormatter) MarshalBool(v bool) {
	f.appendSeparator()
	f.buf.AppendBool(v)
}

func (f *jsonFormatter) MarshalString(v string) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeString(f.buf, v)
	f.buf.AppendByte('"')
}

func (f *jsonFormatter) MarshalInt64(v int64) {
	f.appendSeparator()
	f.buf.AppendInt64(v)
}

func (f *jsonFormatter) MarshalInt32(v int32) {
	f.appendSeparator()
	f.buf.AppendInt32(v)
}

func (f *jsonFormatter) MarshalInt16(v int16) {
	f.appendSeparator()
	f.buf.AppendInt16(v)
}

func (f *jsonFormatter) MarshalInt8(v int8) {
	f.appendSeparator()
	f.buf.AppendInt8(v)
}

func (f *jsonFormatter) MarshalUint64(v uint64) {
	f.appendSeparator()
	f.buf.AppendUint64(v)
}

func (f *jsonFormatter) MarshalUint32(v uint32) {
	f.appendSeparator()
	f.buf.AppendUint32(v)
}

func (f *jsonFormatter) MarshalUint16(v uint16) {
	f.appendSeparator()
	f.buf.AppendUint16(v)
}

func (f *jsonFormatter) MarshalUint8(v uint8) {
	f.appendSeparator()
	f.buf.AppendUint8(v)
}

func (f *jsonFormatter) MarshalFloat64(v float64) {
	f.appendSeparator()
	f.buf.AppendFloat64(v)
}

func (f *jsonFormatter) MarshalFloat32(v float32) {
	f.appendSeparator()
	f.buf.AppendFloat32(v)
}

func (f *jsonFormatter) MarshalDuration(v time.Duration) {
	f.appendSeparator()
	f.FormatDuration(v, f)
}

func (f *jsonFormatter) MarshalTime(v time.Time) {
	f.appendSeparator()
	f.FormatTime(v, f)
}

func (f *jsonFormatter) MarshalBytes(v []byte) {
	f.buf.AppendByte('"')
	// f.buf.AppendString(base64.StdEncoding.EncodeToString(v))

	// TODO: add ensure size by enc.EncodedLen(len(src))
	base64.StdEncoding.Encode(f.buf.Buf, v)
	// e := base64.NewEncoder(base64.StdEncoding, f.buf)
	// e.Write(v)
	f.buf.AppendByte('"')
}

func (f *jsonFormatter) MarshalBools(v []bool) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalBool(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalInts64(v []int64) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalInt64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalInts32(v []int32) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalInt32(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalInts16(v []int16) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalInt16(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalInts8(v []int8) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalInt8(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalUints64(v []uint64) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalUint64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalUints32(v []uint32) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalUint32(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalUints16(v []uint16) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalUint16(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalUints8(v []uint8) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalUint8(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalFloats64(v []float64) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalFloat64(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalFloats32(v []float32) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalFloat32(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalDurations(v []time.Duration) {
	f.buf.AppendByte('[')
	for i := range v {
		f.MarshalDuration(v[i])
	}
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalArray(v ArrayMarshaller) {
	f.buf.AppendByte('[')
	v.MarshalLogfArray(f)
	f.buf.AppendByte(']')
}

func (f *jsonFormatter) MarshalObject(v ObjectMarshaller) {
	f.buf.AppendByte('{')
	v.MarshalLogfObject(f)
	f.buf.AppendByte('}')
}

func (f *jsonFormatter) appendSeparator() {
	if f.buf.Len() == 0 {
		return
	}
	switch f.buf.Back() {
	case '{', '[', ':', ',':
		return
	}
	f.buf.AppendByte(',')
}

func (f *jsonFormatter) addKey(k string) {
	f.appendSeparator()
	f.buf.AppendByte('"')
	EscapeString(f.buf, k)
	f.buf.AppendByte('"')
	f.buf.AppendByte(':')
}

func (f *jsonFormatter) Format(buf *Buffer, e Entry) error {
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

	// for _, field := range e.DerivedFields {
	// 	buf.AppendString(",")
	// 	EscapeString(buf, field.Key)
	// 	buf.AppendString(":")
	// 	field.Accept(f)
	// }

	if bytes, ok := f.cache.Get(e.LoggerID); ok {
		buf.AppendBytes(bytes)
	} else {
		le := buf.Len()
		for _, field := range e.DerivedFields {
			field.Accept(f)
		}

		bf := make([]byte, buf.Len()-le)
		copy(bf, buf.Buf[le:])
		f.cache.Set(e.LoggerID, bf)
	}

	buf.AppendByte('}')
	buf.AppendByte('\n')

	// fmt.Println(string(buf.Buf))
	// panic(string(buf.Buf))

	buf.Flush()

	return nil
}
