package logf

import (
	"strconv"
)

type Buffer struct {
	Buf []byte
	err error
}

func NewBuffer(capacity int) *Buffer {
	return &Buffer{make([]byte, 0, capacity), nil}
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

func (b *Buffer) Reset() {
	b.collapse()
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
