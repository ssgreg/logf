package logf

import (
	"time"
	"unsafe"
)

type Snapshotter interface {
	TakeSnapshot() interface{}
}

// func TakeSnapshot(v interface{}) interface{} {
// 	switch rv := v.(type) {
// 	// byte is uint8
// 	case bool,
// 		int, int8, int16, int32, int64,
// 		uint, uint8, uint16, uint32, uint64,
// 		uintptr,
// 		float32, float64,
// 		string:
// 		return v
// 	case time.Duration, time.Time:
// 		return v
// 	case []bool:
// 		return snapshotBoolArray(rv)
// 	case []int:
// 		return snapshotIntArray(rv)
// 	case []int8:
// 		return snapshotInt8Array(rv)
// 	case []int16:
// 		return snapshotInt16Array(rv)
// 	case []int32:
// 		return snapshotInt32Array(rv)
// 	case []int64:
// 		return snapshotInt64Array(rv)
// 	case []uint:
// 		return snapshotUintArray(rv)
// 	case []uint8:
// 		return snapshotUint8Array(rv)
// 	case []uint16:
// 		return snapshotUint16Array(rv)
// 	case []uint32:
// 		return snapshotUint32Array(rv)
// 	case []uint64:
// 		return snapshotUint64Array(rv)
// 	case []float32:
// 		return snapshotFloat32Array(rv)
// 	case []float64:
// 		return snapshotFloat64Array(rv)
// 	case []string:
// 		return snapshotStringArray(rv)
// 	case []time.Duration:
// 		return snapshotDurationArray(rv)
// 	case []time.Time:
// 		return snapshotTimeArray(rv)
// 	case error:
// 		return v
// 	case Snapshotter:
// 		return rv.TakeSnapshot()
// 	case fmt.Stringer:
// 		return rv.String()
// 	default:
// 		switch reflect.TypeOf(v).Kind() {
// 		case reflect.Bool,
// 			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
// 			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
// 			reflect.Uintptr,
// 			reflect.Float32, reflect.Float64,
// 			reflect.String:
// 			return v
// 		}

// 		return fmt.Sprint(rv)
// 	}
// }

// func snapshotBoolArray(b []bool) []bool {
// 	buf := make([]bool, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotIntArray(b []int) interface{} {
// 	buf := make([]int, len(b))
// 	copy(buf, b)

// 	return b
// }

// func snapshotInt8Array(b []int8) []int8 {
// 	buf := make([]int8, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotInt16Array(b []int16) []int16 {
// 	buf := make([]int16, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotInt32Array(b []int32) []int32 {
// 	buf := make([]int32, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotInt64Array(b []int64) []int64 {
// 	buf := make([]int64, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotUintArray(b []uint) []uint {
// 	buf := make([]uint, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotUint8Array(b []uint8) []uint8 {
// 	buf := make([]uint8, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotUint16Array(b []uint16) []uint16 {
// 	buf := make([]uint16, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotUint32Array(b []uint32) []uint32 {
// 	buf := make([]uint32, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotUint64Array(b []uint64) []uint64 {
// 	buf := make([]uint64, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotFloat32Array(b []float32) []float32 {
// 	buf := make([]float32, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotFloat64Array(b []float64) []float64 {
// 	buf := make([]float64, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotStringArray(b []string) []string {
// 	buf := make([]string, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotDurationArray(b []time.Duration) []time.Duration {
// 	buf := make([]time.Duration, len(b))
// 	copy(buf, b)

// 	return buf
// }

// func snapshotTimeArray(b []time.Time) interface{} {
// 	buf := make([]time.Time, len(b))
// 	copy(buf, b)

// 	return buf
// }

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
