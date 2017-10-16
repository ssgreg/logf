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

func (b *Buffer) Error() error {
	return b.err
}

func (b *Buffer) EnsureSize(s int) []byte {
	if !b.ensureSizeInternal(s) {
		panic(fmt.Sprintf("logf not able to ensure size %d, max is %d", s, cap(b.Buf)))
	}
	return b.Buf
}

func (b *Buffer) AppendString(data string) {
	if !b.ensureSizeInternal(len(data)) {
		_, b.err = b.w.Write(([]byte)(data))
	}
	b.Buf = append(b.Buf, data...)
}

func (b *Buffer) AppendBytes(data []byte) {
	if !b.ensureSizeInternal(len(data)) {
		_, b.err = b.w.Write(data)
	}
	b.Buf = append(b.Buf, data...)
}

func (b *Buffer) AppendByte(data byte) {
	b.Buf = append(b.EnsureSize(1), data)
}

func (b *Buffer) Flush() {
	if len(b.Buf) != 0 {
		_, b.err = b.w.Write(b.Buf)
		b.collapse()
	}
}

func (b *Buffer) collapse() {
	b.Buf = b.Buf[:0]
}

func (b *Buffer) ensureSizeInternal(s int) bool {
	if cap(b.Buf)-len(b.Buf) < s {
		b.Flush()
	}
	return cap(b.Buf) >= s
}

const hex = "0123456789abcdef"

func EscapeString(buf *Buffer, s string) error {
	buf.AppendByte('"')
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
	buf.AppendByte('"')
	return nil
}

func KnownTypeToBuf(buf *Buffer, v interface{}) bool {
	switch rv := v.(type) {
	case string:
		EscapeString(buf, rv)
	case bool:
		boolToBuf(buf, rv)
	case int:
		intToBuf(buf, rv)
	case int8:
		int8ToBuf(buf, rv)
	case int16:
		int16ToBuf(buf, rv)
	case int32:
		int32ToBuf(buf, rv)
	case int64:
		int64ToBuf(buf, rv)
	case uint:
		uintToBuf(buf, rv)
	case uint8:
		uint8ToBuf(buf, rv)
	case uint16:
		uint16ToBuf(buf, rv)
	case uint32:
		uint32ToBuf(buf, rv)
	case uint64:
		uint64ToBuf(buf, rv)
	case float32:
		float32ToBuf(buf, rv)
	case float64:
		float64ToBuf(buf, rv)
	case fmt.Stringer:
		EscapeString(buf, rv.String())
	default:
		switch reflect.TypeOf(rv).Kind() {
		case reflect.String:
			EscapeString(buf, reflect.ValueOf(rv).String())
		case reflect.Bool:
			boolToBuf(buf, reflect.ValueOf(rv).Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			int64ToBuf(buf, reflect.ValueOf(rv).Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uint64ToBuf(buf, reflect.ValueOf(rv).Uint())
		case reflect.Float32, reflect.Float64:
			float64ToBuf(buf, reflect.ValueOf(rv).Float())
		default:
			return false
		}
	}
	return true
}

func uint8ToBuf(buf *Buffer, n uint8) {
	buf.Buf = strconv.AppendUint(buf.EnsureSize(3), uint64(n), 10)
}

func uint16ToBuf(buf *Buffer, n uint16) {
	buf.Buf = strconv.AppendUint(buf.EnsureSize(5), uint64(n), 10)
}

func uint32ToBuf(buf *Buffer, n uint32) {
	buf.Buf = strconv.AppendUint(buf.EnsureSize(10), uint64(n), 10)
}

func uintToBuf(buf *Buffer, n uint) {
	buf.Buf = strconv.AppendUint(buf.EnsureSize(20), uint64(n), 10)
}

func uint64ToBuf(buf *Buffer, n uint64) {
	buf.Buf = strconv.AppendUint(buf.EnsureSize(20), n, 10)
}

func int8ToBuf(buf *Buffer, n int8) {
	buf.Buf = strconv.AppendInt(buf.EnsureSize(4), int64(n), 10)
}

func int16ToBuf(buf *Buffer, n int16) {
	buf.Buf = strconv.AppendInt(buf.EnsureSize(6), int64(n), 10)
}

func int32ToBuf(buf *Buffer, n int32) {
	buf.Buf = strconv.AppendInt(buf.EnsureSize(11), int64(n), 10)
}

func intToBuf(buf *Buffer, n int) {
	buf.Buf = strconv.AppendInt(buf.EnsureSize(21), int64(n), 10)
}

func int64ToBuf(buf *Buffer, n int64) {
	buf.Buf = strconv.AppendInt(buf.EnsureSize(21), n, 10)
}

func float32ToBuf(buf *Buffer, n float32) {
	buf.Buf = strconv.AppendFloat(buf.EnsureSize(20), float64(n), 'g', -1, 32)
}

func float64ToBuf(buf *Buffer, n float64) {
	buf.Buf = strconv.AppendFloat(buf.EnsureSize(20), n, 'g', -1, 64)
}

func boolToBuf(buf *Buffer, n bool) {
	if n {
		buf.Buf = append(buf.EnsureSize(4), "true"...)
	} else {
		buf.Buf = append(buf.EnsureSize(5), "false"...)
	}
}
