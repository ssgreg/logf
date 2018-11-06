package logf

import (
	"math"
	"reflect"
	"time"
	"unsafe"
)

// TODO: special case that calls String() for the passed object

type FieldType byte

const (
	FieldTypeUnknown FieldType = iota
	FieldTypeAny
	FieldTypeBool
	FieldTypeInt64
	FieldTypeInt32
	FieldTypeInt16
	FieldTypeInt8
	FieldTypeUint64
	FieldTypeUint32
	FieldTypeUint16
	FieldTypeUint8
	FieldTypeFloat64
	FieldTypeFloat32
	FieldTypeDuration
	FieldTypeError
	FieldTypeTime

	FieldTypeBytes
	FieldTypeBytesToString
	FieldTypeBytesToBools
	FieldTypeBytesToInts64
	FieldTypeBytesToInts32
	FieldTypeBytesToInts16
	FieldTypeBytesToInts8
	FieldTypeBytesToUints64
	FieldTypeBytesToUints32
	FieldTypeBytesToUints16
	FieldTypeBytesToUints8
	FieldTypeBytesToFloats64
	FieldTypeBytesToFloats32
	FieldTypeBytesToDurations

	FieldTypeArray
	FieldTypeObject
)

// Special cases that are processed during snapshoting phase.
const (
	FieldTypeRawMask FieldType = 1<<7 + iota
	FieldTypeRawBytes
	FieldTypeRawBytesToBools
	FieldTypeRawBytesToInts64
	FieldTypeRawBytesToInts32
	FieldTypeRawBytesToInts16
	FieldTypeRawBytesToInts8
	FieldTypeRawBytesToUints64
	FieldTypeRawBytesToUints32
	FieldTypeRawBytesToUints16
	FieldTypeRawBytesToUints8
	FieldTypeRawBytesToFloats64
	FieldTypeRawBytesToFloats32
	FieldTypeRawBytesToDurations
)

type Field struct {
	Key   string
	Type  FieldType
	Any   interface{}
	Int   int64
	Bytes []byte
}

func (fd Field) Accept(v FieldEncoder) {
	switch fd.Type {
	case FieldTypeAny:
		v.EncodeFieldAny(fd.Key, fd.Any)
	case FieldTypeBool:
		v.EncodeFieldBool(fd.Key, fd.Int != 0)
	case FieldTypeInt64:
		v.EncodeFieldInt64(fd.Key, fd.Int)
	case FieldTypeInt32:
		v.EncodeFieldInt32(fd.Key, int32(fd.Int))
	case FieldTypeInt16:
		v.EncodeFieldInt16(fd.Key, int16(fd.Int))
	case FieldTypeInt8:
		v.EncodeFieldInt8(fd.Key, int8(fd.Int))
	case FieldTypeUint64:
		v.EncodeFieldUint64(fd.Key, uint64(fd.Int))
	case FieldTypeUint32:
		v.EncodeFieldUint32(fd.Key, uint32(fd.Int))
	case FieldTypeUint16:
		v.EncodeFieldUint16(fd.Key, uint16(fd.Int))
	case FieldTypeUint8:
		v.EncodeFieldUint8(fd.Key, uint8(fd.Int))
	case FieldTypeFloat32:
		v.EncodeFieldFloat32(fd.Key, math.Float32frombits(uint32(fd.Int)))
	case FieldTypeFloat64:
		v.EncodeFieldFloat64(fd.Key, math.Float64frombits(uint64(fd.Int)))
	case FieldTypeDuration:
		v.EncodeFieldDuration(fd.Key, time.Duration(fd.Int))
	case FieldTypeError:
		if fd.Any != nil {
			v.EncodeFieldError(fd.Key, fd.Any.(error))
		} else {
			v.EncodeFieldError(fd.Key, nil)
		}
	case FieldTypeTime:
		if fd.Any != nil {
			v.EncodeFieldTime(fd.Key, time.Unix(0, fd.Int).In(fd.Any.(*time.Location)))
		} else {
			v.EncodeFieldTime(fd.Key, time.Unix(0, fd.Int))
		}
	case FieldTypeArray:
		v.EncodeFieldArray(fd.Key, fd.Any.(ArrayEncoder))
	case FieldTypeObject:
		v.EncodeFieldObject(fd.Key, fd.Any.(ObjectEncoder))
	case FieldTypeBytes:
		v.EncodeFieldBytes(fd.Key, fd.Bytes)
	case FieldTypeBytesToString:
		v.EncodeFieldString(fd.Key, *(*string)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToBools:
		v.EncodeFieldBools(fd.Key, *(*[]bool)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToInts64:
		v.EncodeFieldInts64(fd.Key, *(*[]int64)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToInts32:
		v.EncodeFieldInts32(fd.Key, *(*[]int32)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToInts16:
		v.EncodeFieldInts16(fd.Key, *(*[]int16)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToInts8:
		v.EncodeFieldInts8(fd.Key, *(*[]int8)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToUints64:
		v.EncodeFieldUints64(fd.Key, *(*[]uint64)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToUints32:
		v.EncodeFieldUints32(fd.Key, *(*[]uint32)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToUints16:
		v.EncodeFieldUints16(fd.Key, *(*[]uint16)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToUints8:
		v.EncodeFieldUints8(fd.Key, *(*[]uint8)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToFloats64:
		v.EncodeFieldFloats64(fd.Key, *(*[]float64)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToFloats32:
		v.EncodeFieldFloats32(fd.Key, *(*[]float32)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToDurations:
		v.EncodeFieldDurations(fd.Key, *(*[]time.Duration)(unsafe.Pointer(&fd.Bytes)))
	}
}

func Any(k string, v interface{}) Field {
	switch rv := v.(type) {
	case bool:
		return Bool(k, rv)
	case int:
		return Int(k, rv)
	case int64:
		return Int64(k, rv)
	case int32:
		return Int32(k, rv)
	case int16:
		return Int16(k, rv)
	case int8:
		return Int8(k, rv)
	case uint:
		return Uint(k, rv)
	case uint64:
		return Uint64(k, rv)
	case uint32:
		return Uint32(k, rv)
	case uint16:
		return Uint16(k, rv)
	case uint8:
		return Uint8(k, rv)
	case float64:
		return Float64(k, rv)
	case float32:
		return Float32(k, rv)
	case time.Time:
		return Time(k, rv)
	case time.Duration:
		return Duration(k, rv)
	case error:
		return NamedError(k, rv)
	case ArrayEncoder:
		return Array(k, rv)
	case ObjectEncoder:
		return Object(k, rv)
	case []byte:
		return Bytes(k, rv)
	case []string:
		return Strings(k, rv)
	case []bool:
		return Bools(k, rv)
	case []int:
		return Ints(k, rv)
	case []int64:
		return Ints64(k, rv)
	case []int32:
		return Ints32(k, rv)
	case []int16:
		return Ints16(k, rv)
	case []int8:
		return Ints8(k, rv)
	case []uint:
		return Uints(k, rv)
	case []uint64:
		return Uints64(k, rv)
	case []uint32:
		return Uints32(k, rv)
	case []uint16:
		return Uints16(k, rv)
	case []float64:
		return Floats64(k, rv)
	case []float32:
		return Floats32(k, rv)
	case []time.Duration:
		return Durations(k, rv)

	default:
		switch reflect.TypeOf(rv).Kind() {
		case reflect.String:
			return String(k, reflect.ValueOf(rv).String())
		case reflect.Bool:
			return Bool(k, reflect.ValueOf(rv).Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return Int64(k, reflect.ValueOf(rv).Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return Uint64(k, reflect.ValueOf(rv).Uint())
		case reflect.Float32, reflect.Float64:
			return Float64(k, reflect.ValueOf(rv).Float())
		}
	}

	return Field{Key: k, Type: FieldTypeAny, Any: v}
}

// func Any(k string, v interface{}) Field {
// 	return Field{Key: k, Type: FieldTypeAny, Any: v}
// }

func Bool(k string, v bool) Field {
	var tmp int64
	if v {
		tmp = 1
	}

	return Field{Key: k, Type: FieldTypeBool, Int: tmp}
}

func Int(k string, v int) Field {
	return Field{Key: k, Type: FieldTypeInt64, Int: int64(v)}
}

func Int64(k string, v int64) Field {
	return Field{Key: k, Type: FieldTypeInt64, Int: v}
}

func Int32(k string, v int32) Field {
	return Field{Key: k, Type: FieldTypeInt32, Int: int64(v)}
}

func Int16(k string, v int16) Field {
	return Field{Key: k, Type: FieldTypeInt16, Int: int64(v)}
}

func Int8(k string, v int8) Field {
	return Field{Key: k, Type: FieldTypeInt8, Int: int64(v)}
}

func Uint(k string, v uint) Field {
	return Field{Key: k, Type: FieldTypeUint64, Int: int64(v)}
}

func Uint64(k string, v uint64) Field {
	return Field{Key: k, Type: FieldTypeUint64, Int: int64(v)}
}

func Uint32(k string, v uint32) Field {
	return Field{Key: k, Type: FieldTypeUint32, Int: int64(v)}
}

func Uint16(k string, v uint16) Field {
	return Field{Key: k, Type: FieldTypeUint16, Int: int64(v)}
}

func Uint8(k string, v uint8) Field {
	return Field{Key: k, Type: FieldTypeUint8, Int: int64(v)}
}

func Float64(k string, v float64) Field {
	return Field{Key: k, Type: FieldTypeFloat64, Int: int64(math.Float64bits(v))}
}

func Float32(k string, v float32) Field {
	return Field{Key: k, Type: FieldTypeFloat32, Int: int64(math.Float32bits(v))}
}

func Duration(k string, v time.Duration) Field {
	return Field{Key: k, Type: FieldTypeDuration, Int: int64(v)}
}

func Bytes(k string, v []byte) Field {
	return Field{Key: k, Type: FieldTypeRawBytes, Bytes: v}
}

func String(k string, v string) Field {
	return Field{Key: k, Type: FieldTypeBytesToString, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Strings(k string, v []string) Field {
	return Field{Key: k, Type: FieldTypeArray, Any: stringArray(v)}
}

func Bools(k string, v []bool) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToBools, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Ints(k string, v []int) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToInts64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Ints64(k string, v []int64) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToInts64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Ints32(k string, v []int32) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToInts32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Ints16(k string, v []int16) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToInts16, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Ints8(k string, v []int8) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToInts8, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Uints(k string, v []uint) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToUints64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Uints64(k string, v []uint64) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToUints64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Uints32(k string, v []uint32) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToUints32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Uints16(k string, v []uint16) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToUints16, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Uints8(k string, v []uint8) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToUints8, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Floats64(k string, v []float64) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToFloats64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Floats32(k string, v []float32) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToFloats32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func Durations(k string, v []time.Duration) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToDurations, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstBytes(k string, v []byte) Field {
	return Field{Key: k, Type: FieldTypeBytes, Bytes: v}
}

func ConstBools(k string, v []bool) Field {
	return Field{Key: k, Type: FieldTypeBytesToBools, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstInts(k string, v []int) Field {
	return Field{Key: k, Type: FieldTypeBytesToInts64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstInts64(k string, v []int64) Field {
	return Field{Key: k, Type: FieldTypeBytesToInts64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstInts32(k string, v []int32) Field {
	return Field{Key: k, Type: FieldTypeBytesToInts32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstInts16(k string, v []int16) Field {
	return Field{Key: k, Type: FieldTypeBytesToInts16, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstInts8(k string, v []int8) Field {
	return Field{Key: k, Type: FieldTypeBytesToInts8, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstUints(k string, v []uint) Field {
	return Field{Key: k, Type: FieldTypeBytesToUints64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstUints64(k string, v []uint64) Field {
	return Field{Key: k, Type: FieldTypeBytesToUints64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstUints32(k string, v []uint32) Field {
	return Field{Key: k, Type: FieldTypeBytesToUints32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstUints16(k string, v []uint16) Field {
	return Field{Key: k, Type: FieldTypeBytesToUints16, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstUints8(k string, v []uint8) Field {
	return Field{Key: k, Type: FieldTypeBytesToUints8, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstFloats64(k string, v []float64) Field {
	return Field{Key: k, Type: FieldTypeBytesToFloats64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstFloats32(k string, v []float32) Field {
	return Field{Key: k, Type: FieldTypeBytesToFloats32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func ConstDurations(k string, v []time.Duration) Field {
	return Field{Key: k, Type: FieldTypeBytesToDurations, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

func NamedError(k string, v error) Field {
	return Field{Key: k, Type: FieldTypeError, Any: v}
}

func Error(v error) Field {
	return NamedError("error", v)
}

func Time(k string, v time.Time) Field {
	return Field{Key: k, Type: FieldTypeTime, Int: v.UnixNano(), Any: v.Location()}
}

func Array(k string, v ArrayEncoder) Field {
	return Field{Key: k, Type: FieldTypeArray, Any: v}
}

func Object(k string, v ObjectEncoder) Field {
	return Field{Key: k, Type: FieldTypeObject, Any: v}
}

type stringArray []string

func (o stringArray) EncodeLogfArray(e TypeEncoder) error {
	for i := range o {
		e.EncodeTypeString(o[i])
	}

	return nil
}
