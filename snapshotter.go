package logf

import (
	"fmt"
	"reflect"
	"time"
)

type Snapshotter interface {
	TakeSnapshot() interface{}
}

type Greger interface {
	MyGreg()
}

func TakeSnapshot(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch rv := v.(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64, complex64, complex128, string:
		return v
	case time.Time:
		return v
	case []byte:
		return snapshotByteArray(rv)
	case []int:
		return snapshotIntArray(rv)
	case []string:
		return snapshotStringArray(rv)
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
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
			reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64,
			reflect.Complex128, reflect.String:
			return v
		}
		return fmt.Sprint(rv)
	}
}

func snapshotByteArray(b []byte) interface{} {
	buf := make([]byte, len(b))
	copy(buf, b)
	return buf
}

func snapshotIntArray(b []int) interface{} {
	buf := make([]int, len(b))
	copy(buf, b)
	return buf
}

func snapshotStringArray(b []string) interface{} {
	buf := make([]string, len(b))
	copy(buf, b)
	return buf
}

func snapshotTimeArray(b []time.Time) interface{} {
	buf := make([]time.Time, len(b))
	copy(buf, b)
	return buf
}
