package logf

import (
	"strconv"
)

// PageSize is the recommended buffer size.
const (
	PageSize = 4 * 1024
)

// NewBuffer creates the new instance of Buffer with default capacity.
func NewBuffer() *Buffer {
	return NewBufferWithCapacity(PageSize)
}

// NewBufferWithCapacity creates the new instance of Buffer with the given
// capacity.
func NewBufferWithCapacity(capacity int) *Buffer {
	return &Buffer{make([]byte, 0, capacity)}
}

// Buffer is a helping wrapper for byte slice.
type Buffer struct {
	Data []byte
}

// Write implements io.Writer.
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.Data = append(b.Data, p...)

	return len(p), nil
}

// String implements fmt.Stringer.
func (b *Buffer) String() string {
	return string(b.Bytes())
}

// EnsureSize ensures that the Buffer is able to append 's' bytes without
// a further realloc.
func (b *Buffer) EnsureSize(s int) {
	if cap(b.Data)-len(b.Data) < s {
		tmpLen := len(b.Data)
		tmp := make([]byte, tmpLen, tmpLen+s+(tmpLen>>1))
		copy(tmp, b.Data)
		b.Data = tmp
	}
}

// ExtendBytes extends the Buffer with the given size and returns a slice
// to the extended part of the Buffer.
func (b *Buffer) ExtendBytes(s int) []byte {
	b.EnsureSize(s)
	n := len(b.Data)
	b.Data = b.Data[:n+s]

	return b.Data[n:]
}

// AppendString appends a string to the Buffer.
func (b *Buffer) AppendString(data string) {
	b.Data = append(b.Data, data...)
}

// AppendBytes appends a byte slice to the Buffer.
func (b *Buffer) AppendBytes(data []byte) {
	b.Data = append(b.Data, data...)
}

// AppendByte appends a single byte to the Buffer.
func (b *Buffer) AppendByte(data byte) {
	b.Data = append(b.Data, data)
}

// Truncate shrinks the Buffer to the given length.
func (b *Buffer) Truncate(n int) {
	b.Data = b.Data[:n]
}

// Reset resets the underlying byte slice.
func (b *Buffer) Reset() {
	b.Data = b.Data[:0]
}

// Back returns the last byte of the underlying byte slice. A caller is in
// charge of checking that the Buffer is not empty.
func (b *Buffer) Back() byte {
	return b.Data[len(b.Data)-1]
}

// Bytes returns the underlying byte slice as is.
func (b *Buffer) Bytes() []byte {
	return b.Data
}

// Len returns the length of the underlying byte slice.
func (b *Buffer) Len() int {
	return len(b.Data)
}

// Cap returns the capacity of the underlying byte slice.
func (b *Buffer) Cap() int {
	return cap(b.Data)
}

// AppendUint appends the string form in the base 10 of the given unsigned
// integer.
func (b *Buffer) AppendUint(n uint64) {
	b.Data = strconv.AppendUint(b.Data, n, 10)
}

// AppendInt appends the string form in the base 10 of the given integer.
func (b *Buffer) AppendInt(n int64) {
	b.Data = strconv.AppendInt(b.Data, n, 10)
}

// AppendFloat32 appends the string form of the given float32.
func (b *Buffer) AppendFloat32(n float32) {
	b.Data = strconv.AppendFloat(b.Data, float64(n), 'g', -1, 32)
}

// AppendFloat64 appends the string form of the given float64.
func (b *Buffer) AppendFloat64(n float64) {
	b.Data = strconv.AppendFloat(b.Data, n, 'g', -1, 64)
}

// AppendBool appends "true" or "false" according to the given bool.
func (b *Buffer) AppendBool(n bool) {
	b.Data = strconv.AppendBool(b.Data, n)
}
