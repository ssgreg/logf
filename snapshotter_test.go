package logf

import (
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestSnapshotNilAny(t *testing.T) {
	f := Field{Type: FieldTypeAny}

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.Nil(t, f.Bytes)
}

type notImplementSnapshotter struct{}

func TestSnapshotAnyNotImplementSnapshotter(t *testing.T) {
	f := Field{Type: FieldTypeAny, Any: &notImplementSnapshotter{}}
	any := f.Any

	snapshotField(&f)
	assert.Equal(t, unsafe.Pointer(f.Any.(*notImplementSnapshotter)), unsafe.Pointer(any.(*notImplementSnapshotter)))
}

func TestSnapshotAnyImplementSnapshotter(t *testing.T) {
	obj := testSnapshotter{}
	f := Field{Type: FieldTypeAny, Any: &obj}

	snapshotField(&f)
	assert.True(t, obj.Called)
	assert.NotEqual(t, unsafe.Pointer(f.Any.(*testSnapshotter)), &obj)
}

func TestSnapshotBytes(t *testing.T) {
	f := Bytes("", []byte{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotBools(t *testing.T) {
	f := Bools("", []bool{true})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotInts(t *testing.T) {
	f := Ints("", []int{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotInts64(t *testing.T) {
	f := Ints64("", []int64{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotInts32(t *testing.T) {
	f := Ints32("", []int32{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotInts16(t *testing.T) {
	f := Ints16("", []int16{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotInts8(t *testing.T) {
	f := Ints8("", []int8{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotUints(t *testing.T) {
	f := Uints("", []uint{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotUints64(t *testing.T) {
	f := Uints64("", []uint64{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotUints32(t *testing.T) {
	f := Uints32("", []uint32{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotUints16(t *testing.T) {
	f := Uints16("", []uint16{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotUints8(t *testing.T) {
	f := Uints8("", []uint8{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotFloats64(t *testing.T) {
	f := Floats64("", []float64{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotFloats32(t *testing.T) {
	f := Floats32("", []float32{0})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}

func TestSnapshotDurations(t *testing.T) {
	f := Durations("", []time.Duration{time.Second})
	assert.True(t, f.Type&FieldTypeRawMask != 0)
	rawArray := f.Bytes

	snapshotField(&f)
	assert.True(t, f.Type&FieldTypeRawMask == 0)
	assert.NotEqual(t, unsafe.Pointer(&f.Bytes), unsafe.Pointer(&rawArray))
	assert.Equal(t, f.Bytes, rawArray)
}
