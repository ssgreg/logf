package logf

import (
	"fmt"
	"reflect"
	"time"
)

type Snapshotter interface {
	TakeSnapshot() interface{}
}

func TakeSnapshot(v interface{}) interface{} {
	switch rv := v.(type) {
	// byte is uint8
	case bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		uintptr,
		float32, float64,
		string:
		return v
	case time.Duration, time.Time:
		return v
	case []bool:
		return snapshotBoolArray(rv)
	case []int:
		return snapshotIntArray(rv)
	case []int8:
		return snapshotInt8Array(rv)
	case []int16:
		return snapshotInt16Array(rv)
	case []int32:
		return snapshotInt32Array(rv)
	case []int64:
		return snapshotInt64Array(rv)
	case []uint:
		return snapshotUintArray(rv)
	case []uint8:
		return snapshotUint8Array(rv)
	case []uint16:
		return snapshotUint16Array(rv)
	case []uint32:
		return snapshotUint32Array(rv)
	case []uint64:
		return snapshotUint64Array(rv)
	case []float32:
		return snapshotFloat32Array(rv)
	case []float64:
		return snapshotFloat64Array(rv)
	case []string:
		return snapshotStringArray(rv)
	case []time.Duration:
		return snapshotDurationArray(rv)
	case []time.Time:
		return snapshotTimeArray(rv)
	case error:
		return v
	case Snapshotter:
		return rv.TakeSnapshot()
	case fmt.Stringer:
		return rv.String()
	default:
		switch reflect.TypeOf(v).Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Uintptr,
			reflect.Float32, reflect.Float64,
			reflect.String:
			return v
		}

		return fmt.Sprint(rv)
	}
}

func snapshotBoolArray(b []bool) []bool {
	buf := make([]bool, len(b))
	copy(buf, b)

	return buf
}

func snapshotIntArray(b []int) interface{} {
	buf := make([]int, len(b))
	copy(buf, b)

	return b
}

func snapshotInt8Array(b []int8) []int8 {
	buf := make([]int8, len(b))
	copy(buf, b)

	return buf
}

func snapshotInt16Array(b []int16) []int16 {
	buf := make([]int16, len(b))
	copy(buf, b)

	return buf
}

func snapshotInt32Array(b []int32) []int32 {
	buf := make([]int32, len(b))
	copy(buf, b)

	return buf
}

func snapshotInt64Array(b []int64) []int64 {
	buf := make([]int64, len(b))
	copy(buf, b)

	return buf
}

func snapshotUintArray(b []uint) []uint {
	buf := make([]uint, len(b))
	copy(buf, b)

	return buf
}

func snapshotUint8Array(b []uint8) []uint8 {
	buf := make([]uint8, len(b))
	copy(buf, b)

	return buf
}

func snapshotUint16Array(b []uint16) []uint16 {
	buf := make([]uint16, len(b))
	copy(buf, b)

	return buf
}

func snapshotUint32Array(b []uint32) []uint32 {
	buf := make([]uint32, len(b))
	copy(buf, b)

	return buf
}

func snapshotUint64Array(b []uint64) []uint64 {
	buf := make([]uint64, len(b))
	copy(buf, b)

	return buf
}

func snapshotFloat32Array(b []float32) []float32 {
	buf := make([]float32, len(b))
	copy(buf, b)

	return buf
}

func snapshotFloat64Array(b []float64) []float64 {
	buf := make([]float64, len(b))
	copy(buf, b)

	return buf
}

func snapshotStringArray(b []string) []string {
	buf := make([]string, len(b))
	copy(buf, b)

	return buf
}

func snapshotDurationArray(b []time.Duration) []time.Duration {
	buf := make([]time.Duration, len(b))
	copy(buf, b)

	return buf
}

func snapshotTimeArray(b []time.Time) interface{} {
	buf := make([]time.Time, len(b))
	copy(buf, b)

	return buf
}
