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
	v := []bool{true}
	f := Bools("", v)

	snapshotField(&f)
	v[0] = false
	assert.Equal(t, []bool{true}, []bool(f.Any.(boolArray)))
}

func TestSnapshotInts(t *testing.T) {
	v := []int{0}
	f := Ints("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []int{0}, []int(f.Any.(intArray)))
}

func TestSnapshotInts64(t *testing.T) {
	v := []int64{0}
	f := Ints64("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []int64{0}, []int64(f.Any.(int64Array)))
}

func TestSnapshotInts32(t *testing.T) {
	v := []int32{0}
	f := Ints32("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []int32{0}, []int32(f.Any.(int32Array)))
}

func TestSnapshotInts16(t *testing.T) {
	v := []int16{0}
	f := Ints16("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []int16{0}, []int16(f.Any.(int16Array)))
}

func TestSnapshotInts8(t *testing.T) {
	v := []int8{0}
	f := Ints8("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []int8{0}, []int8(f.Any.(int8Array)))
}

func TestSnapshotUints(t *testing.T) {
	v := []uint{0}
	f := Uints("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []uint{0}, []uint(f.Any.(uintArray)))
}

func TestSnapshotUints64(t *testing.T) {
	v := []uint64{0}
	f := Uints64("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []uint64{0}, []uint64(f.Any.(uint64Array)))
}

func TestSnapshotUints32(t *testing.T) {
	v := []uint32{0}
	f := Uints32("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []uint32{0}, []uint32(f.Any.(uint32Array)))
}

func TestSnapshotUints16(t *testing.T) {
	v := []uint16{0}
	f := Uints16("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []uint16{0}, []uint16(f.Any.(uint16Array)))
}

func TestSnapshotUints8(t *testing.T) {
	v := []uint8{0}
	f := Uints8("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []uint8{0}, []uint8(f.Any.(uint8Array)))
}

func TestSnapshotFloats64(t *testing.T) {
	v := []float64{0}
	f := Floats64("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []float64{0}, []float64(f.Any.(float64Array)))
}

func TestSnapshotFloats32(t *testing.T) {
	v := []float32{0}
	f := Floats32("", v)

	snapshotField(&f)
	v[0] = 1
	assert.Equal(t, []float32{0}, []float32(f.Any.(float32Array)))
}

func TestSnapshotDurations(t *testing.T) {
	v := []time.Duration{time.Second}
	f := Durations("", v)

	snapshotField(&f)
	v[0] = time.Minute
	assert.Equal(t, []time.Duration{time.Second}, []time.Duration(f.Any.(durationArray)))
}
