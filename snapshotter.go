package logf

import (
	"time"
	"unsafe"
)

// Snapshotter is the interface that allows to do a custom copy of a logging
// object. If the object type implements TaskSnapshot function it will be
// called during the logging procedure in a caller's goroutine.
type Snapshotter interface {
	TakeSnapshot() interface{}
}

// snapshotField calls an appropriate function to snapshot a Field.
func snapshotField(f *Field) {
	if f.Type&FieldTypeRawMask != 0 {
		switch f.Type {
		case FieldTypeRawBytes:
			snapshotRawBytes(f)
		case FieldTypeRawBytesToBools:
			snapshotRawBytesToBools(f)
		case FieldTypeRawBytesToInts64:
			snapshotRawBytesToInts64(f)
		case FieldTypeRawBytesToInts32:
			snapshotRawBytesToInts32(f)
		case FieldTypeRawBytesToInts16:
			snapshotRawBytesToInts16(f)
		case FieldTypeRawBytesToInts8:
			snapshotRawBytesToInts8(f)
		case FieldTypeRawBytesToUints64:
			snapshotRawBytesToUints64(f)
		case FieldTypeRawBytesToUints32:
			snapshotRawBytesToUints32(f)
		case FieldTypeRawBytesToUints16:
			snapshotRawBytesToUints16(f)
		case FieldTypeRawBytesToUints8:
			snapshotRawBytesToUints8(f)
		case FieldTypeRawBytesToFloats64:
			snapshotRawBytesToFloats64(f)
		case FieldTypeRawBytesToFloats32:
			snapshotRawBytesToFloats32(f)
		case FieldTypeRawBytesToDurations:
			snapshotRawBytesToDurations(f)
		}
	}
	if f.Type == FieldTypeAny {
		if f.Any == nil {
			return
		}
		switch rv := f.Any.(type) {
		case Snapshotter:
			f.Any = rv.TakeSnapshot()
		}
	}
}

func snapshotRawBytes(f *Field) {
	cc := make([]byte, len(f.Bytes))
	copy(cc, f.Bytes)
	f.Bytes = cc
	f.Type = FieldTypeBytes
}

func snapshotRawBytesToBools(f *Field) {
	s := *(*[]bool)(unsafe.Pointer(&f.Bytes))
	cc := make([]bool, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToBools
}

func snapshotRawBytesToInts64(f *Field) {
	s := *(*[]int64)(unsafe.Pointer(&f.Bytes))
	cc := make([]int64, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToInts64
}

func snapshotRawBytesToInts32(f *Field) {
	s := *(*[]int32)(unsafe.Pointer(&f.Bytes))
	cc := make([]int32, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToInts32
}

func snapshotRawBytesToInts16(f *Field) {
	s := *(*[]int16)(unsafe.Pointer(&f.Bytes))
	cc := make([]int16, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToInts16
}

func snapshotRawBytesToInts8(f *Field) {
	s := *(*[]int8)(unsafe.Pointer(&f.Bytes))
	cc := make([]int8, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToInts8
}

func snapshotRawBytesToUints64(f *Field) {
	s := *(*[]uint64)(unsafe.Pointer(&f.Bytes))
	cc := make([]uint64, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToUints64
}

func snapshotRawBytesToUints32(f *Field) {
	s := *(*[]uint32)(unsafe.Pointer(&f.Bytes))
	cc := make([]uint32, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToUints32
}

func snapshotRawBytesToUints16(f *Field) {
	s := *(*[]uint16)(unsafe.Pointer(&f.Bytes))
	cc := make([]uint16, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToUints16
}

func snapshotRawBytesToUints8(f *Field) {
	s := *(*[]uint8)(unsafe.Pointer(&f.Bytes))
	cc := make([]uint8, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToUints8
}

func snapshotRawBytesToFloats64(f *Field) {
	s := *(*[]float64)(unsafe.Pointer(&f.Bytes))
	cc := make([]float64, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToFloats64
}

func snapshotRawBytesToFloats32(f *Field) {
	s := *(*[]float32)(unsafe.Pointer(&f.Bytes))
	cc := make([]float32, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToFloats32
}

func snapshotRawBytesToDurations(f *Field) {
	s := *(*[]time.Duration)(unsafe.Pointer(&f.Bytes))
	cc := make([]time.Duration, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToDurations
}
