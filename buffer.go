package logf

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"unicode/utf8"
)

type Buffer struct {
	Buf []byte
	w   io.Writer
	err error
}

func NewBuffer(w io.Writer) *Buffer {
	return &Buffer{make([]byte, 0, 4096*16), w, nil}
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.Buf = append(b.Buf, p...)

	return len(p), nil
}

func (b *Buffer) Error() error {
	return b.err
}

func (b *Buffer) AppendString(data string) {
	b.Buf = append(b.Buf, data...)
}

func (b *Buffer) AppendBytes(data []byte) {
	b.Buf = append(b.Buf, data...)
}

func (b *Buffer) AppendByte(data byte) {
	b.Buf = append(b.Buf, data)
}

func (b *Buffer) Flush() {
	if len(b.Buf) != 0 {
		_, b.err = b.w.Write(b.Buf)
		b.collapse()
	}
}

func (b *Buffer) Back() byte {
	return b.Buf[len(b.Buf)-1]
}

func (b *Buffer) Len() int {
	return len(b.Buf)
}

func (b *Buffer) collapse() {
	b.Buf = b.Buf[:0]
}

func (b *Buffer) AppendUint8(n uint8) {
	b.Buf = strconv.AppendUint(b.Buf, uint64(n), 10)
}

func (b *Buffer) AppendUint16(n uint16) {
	b.Buf = strconv.AppendUint(b.Buf, uint64(n), 10)
}

func (b *Buffer) AppendUint32(n uint32) {
	b.Buf = strconv.AppendUint(b.Buf, uint64(n), 10)
}

func (b *Buffer) AppendUint(n uint) {
	b.Buf = strconv.AppendUint(b.Buf, uint64(n), 10)
}

func (b *Buffer) AppendUint64(n uint64) {
	b.Buf = strconv.AppendUint(b.Buf, n, 10)
}

func (b *Buffer) AppendInt8(n int8) {
	b.Buf = strconv.AppendInt(b.Buf, int64(n), 10)
}

func (b *Buffer) AppendInt16(n int16) {
	b.Buf = strconv.AppendInt(b.Buf, int64(n), 10)
}

func (b *Buffer) AppendInt32(n int32) {
	b.Buf = strconv.AppendInt(b.Buf, int64(n), 10)
}

func (b *Buffer) AppendInt(n int) {
	b.Buf = strconv.AppendInt(b.Buf, int64(n), 10)
}

func (b *Buffer) AppendInt64(n int64) {
	b.Buf = strconv.AppendInt(b.Buf, n, 10)
}

func (b *Buffer) AppendFloat32(n float32) {
	b.Buf = strconv.AppendFloat(b.Buf, float64(n), 'g', -1, 32)
}

func (b *Buffer) AppendFloat64(n float64) {
	b.Buf = strconv.AppendFloat(b.Buf, n, 'g', -1, 64)
}

func (b *Buffer) AppendBool(n bool) {
	if n {
		b.Buf = append(b.Buf, "true"...)
	} else {
		b.Buf = append(b.Buf, "false"...)
	}
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

func EscapeStringBytes(buf *Buffer, s []byte) error {
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
		buf.AppendBool(rv)
	case int:
		buf.AppendInt(rv)
	case int8:
		buf.AppendInt8(rv)
	case int16:
		buf.AppendInt16(rv)
	case int32:
		buf.AppendInt32(rv)
	case int64:
		buf.AppendInt64(rv)
	case uint:
		buf.AppendUint(rv)
	case uint8:
		buf.AppendUint8(rv)
	case uint16:
		buf.AppendUint16(rv)
	case uint32:
		buf.AppendUint32(rv)
	case uint64:
		buf.AppendUint64(rv)
	case float32:
		buf.AppendFloat32(rv)
	case float64:
		buf.AppendFloat64(rv)
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
			buf.AppendBool(reflect.ValueOf(rv).Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			buf.AppendInt64(reflect.ValueOf(rv).Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			buf.AppendUint64(reflect.ValueOf(rv).Uint())
		case reflect.Float32, reflect.Float64:
			buf.AppendFloat64(reflect.ValueOf(rv).Float())
		default:
			return false
		}
	}

	return true
}
