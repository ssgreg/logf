package logf

import (
	"fmt"
	"math"
	"reflect"
	"time"
	"unsafe"
)

// Bool returns a new Field with the given key and bool.
func Bool(k string, v bool) Field {
	var tmp int64
	if v {
		tmp = 1
	}

	return Field{Key: k, Type: FieldTypeBool, Int: tmp}
}

// Int returns a new Field with the given key and int.
func Int(k string, v int) Field {
	return Field{Key: k, Type: FieldTypeInt64, Int: int64(v)}
}

// Int64 returns a new Field with the given key and int64.
func Int64(k string, v int64) Field {
	return Field{Key: k, Type: FieldTypeInt64, Int: v}
}

// Int32 returns a new Field with the given key and int32.
func Int32(k string, v int32) Field {
	return Field{Key: k, Type: FieldTypeInt32, Int: int64(v)}
}

// Int16 returns a new Field with the given key and int16.
func Int16(k string, v int16) Field {
	return Field{Key: k, Type: FieldTypeInt16, Int: int64(v)}
}

// Int8 returns a new Field with the given key and int.8
func Int8(k string, v int8) Field {
	return Field{Key: k, Type: FieldTypeInt8, Int: int64(v)}
}

// Uint returns a new Field with the given key and uint.
func Uint(k string, v uint) Field {
	return Field{Key: k, Type: FieldTypeUint64, Int: int64(v)}
}

// Uint64 returns a new Field with the given key and uint64.
func Uint64(k string, v uint64) Field {
	return Field{Key: k, Type: FieldTypeUint64, Int: int64(v)}
}

// Uint32 returns a new Field with the given key and uint32.
func Uint32(k string, v uint32) Field {
	return Field{Key: k, Type: FieldTypeUint32, Int: int64(v)}
}

// Uint16 returns a new Field with the given key and uint16.
func Uint16(k string, v uint16) Field {
	return Field{Key: k, Type: FieldTypeUint16, Int: int64(v)}
}

// Uint8 returns a new Field with the given key and uint8.
func Uint8(k string, v uint8) Field {
	return Field{Key: k, Type: FieldTypeUint8, Int: int64(v)}
}

// Float64 returns a new Field with the given key and float64.
func Float64(k string, v float64) Field {
	return Field{Key: k, Type: FieldTypeFloat64, Int: int64(math.Float64bits(v))}
}

// Float32 returns a new Field with the given key and float32.
func Float32(k string, v float32) Field {
	return Field{Key: k, Type: FieldTypeFloat32, Int: int64(math.Float32bits(v))}
}

// Duration returns a new Field with the given key and time.Duration.
func Duration(k string, v time.Duration) Field {
	return Field{Key: k, Type: FieldTypeDuration, Int: int64(v)}
}

// Bytes returns a new Field with the given key and slice of bytes.
func Bytes(k string, v []byte) Field {
	return Field{Key: k, Type: FieldTypeRawBytes, Bytes: v}
}

// String returns a new Field with the given key and string.
func String(k string, v string) Field {
	return Field{Key: k, Type: FieldTypeBytesToString, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Strings returns a new Field with the given key and slice of strings.
func Strings(k string, v []string) Field {
	return Field{Key: k, Type: FieldTypeArray, Any: stringArray(v)}
}

type stringArray []string

func (o stringArray) EncodeLogfArray(e TypeEncoder) error {
	for i := range o {
		e.EncodeTypeString(o[i])
	}

	return nil
}

// Bools returns a new Field with the given key and slice of bools.
func Bools(k string, v []bool) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToBools, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Ints returns a new Field with the given key and slice of ints.
func Ints(k string, v []int) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToInts64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Ints64 returns a new Field with the given key and slice of 64-bit ints.
func Ints64(k string, v []int64) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToInts64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Ints32 returns a new Field with the given key and slice of 32-bit ints.
func Ints32(k string, v []int32) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToInts32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Ints16 returns a new Field with the given key and slice of 16-bit ints.
func Ints16(k string, v []int16) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToInts16, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Ints8 returns a new Field with the given key and slice of 8-bit ints.
func Ints8(k string, v []int8) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToInts8, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Uints returns a new Field with the given key and slice of uints.
func Uints(k string, v []uint) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToUints64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Uints64 returns a new Field with the given key and slice of 64-bit uints.
func Uints64(k string, v []uint64) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToUints64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Uints32 returns a new Field with the given key and slice of 32-bit uints.
func Uints32(k string, v []uint32) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToUints32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Uints16 returns a new Field with the given key and slice of 16-bit uints.
func Uints16(k string, v []uint16) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToUints16, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Uints8 returns a new Field with the given key and slice of 8-bit uints.
func Uints8(k string, v []uint8) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToUints8, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Floats64 returns a new Field with the given key and slice of 64-biy floats.
func Floats64(k string, v []float64) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToFloats64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Floats32 returns a new Field with the given key and slice of 32-bit floats.
func Floats32(k string, v []float32) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToFloats32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// Durations returns a new Field with the given key and slice of time.Duration.
func Durations(k string, v []time.Duration) Field {
	return Field{Key: k, Type: FieldTypeRawBytesToDurations, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstBytes returns a new Field with the given key and slice of bytes.
//
// Call ConstBytes if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstBytes(k string, v []byte) Field {
	return Field{Key: k, Type: FieldTypeBytes, Bytes: v}
}

// ConstBools returns a new Field with the given key and slice of bools.
//
// Call ConstBools if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstBools(k string, v []bool) Field {
	return Field{Key: k, Type: FieldTypeBytesToBools, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstInts returns a new Field with the given key and slice of ints.
//
// Call ConstInts if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstInts(k string, v []int) Field {
	return Field{Key: k, Type: FieldTypeBytesToInts64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstInts64 returns a new Field with the given key and slice of 64-bit ints.
//
// Call ConstInts64 if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstInts64(k string, v []int64) Field {
	return Field{Key: k, Type: FieldTypeBytesToInts64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstInts32 returns a new Field with the given key and slice of 32-bit ints.
//
// Call ConstInts32 if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstInts32(k string, v []int32) Field {
	return Field{Key: k, Type: FieldTypeBytesToInts32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstInts16 returns a new Field with the given key and slice of 16-bit ints.
//
// Call ConstInts16 if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstInts16(k string, v []int16) Field {
	return Field{Key: k, Type: FieldTypeBytesToInts16, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstInts8 returns a new Field with the given key and slice of 8-bit ints.
//
// Call ConstInts8 if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstInts8(k string, v []int8) Field {
	return Field{Key: k, Type: FieldTypeBytesToInts8, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstUints returns a new Field with the given key and slice of uints.
//
// Call ConstUints if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstUints(k string, v []uint) Field {
	return Field{Key: k, Type: FieldTypeBytesToUints64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstUints64 returns a new Field with the given key and slice of 64-bit uints.
//
// Call ConstUints64 if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstUints64(k string, v []uint64) Field {
	return Field{Key: k, Type: FieldTypeBytesToUints64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstUints32 returns a new Field with the given key and slice of 32-bit uints.
//
// Call ConstUints32 if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstUints32(k string, v []uint32) Field {
	return Field{Key: k, Type: FieldTypeBytesToUints32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstUints16 returns a new Field with the given key and slice of 16-bit uints.
//
// Call ConstUints16 if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstUints16(k string, v []uint16) Field {
	return Field{Key: k, Type: FieldTypeBytesToUints16, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstUints8 returns a new Field with the given key and slice of 8-bit uints.
//
// Call ConstUints8 if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstUints8(k string, v []uint8) Field {
	return Field{Key: k, Type: FieldTypeBytesToUints8, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstFloats64 returns a new Field with the given key and slice of 64-bit floats.
//
// Call ConstFloats64 if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstFloats64(k string, v []float64) Field {
	return Field{Key: k, Type: FieldTypeBytesToFloats64, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstFloats32 returns a new Field with the given key and slice of 32-bit floats.
//
// Call ConstFloats32 if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstFloats32(k string, v []float32) Field {
	return Field{Key: k, Type: FieldTypeBytesToFloats32, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// ConstDurations returns a new Field with the given key and slice of time.Duration.
//
// Call ConstDurations if your array is const. It has significantly less impact
// on the calling goroutine.
func ConstDurations(k string, v []time.Duration) Field {
	return Field{Key: k, Type: FieldTypeBytesToDurations, Bytes: *(*[]byte)(unsafe.Pointer(&v))}
}

// NamedError returns a new Field with the given key and error.
func NamedError(k string, v error) Field {
	return Field{Key: k, Type: FieldTypeError, Any: v}
}

// Error returns a new Field with the given error. Key is 'error'.
func Error(v error) Field {
	return NamedError("error", v)
}

// Time returns a new Field with the given key and time.Time.
func Time(k string, v time.Time) Field {
	return Field{Key: k, Type: FieldTypeTime, Int: v.UnixNano(), Any: v.Location()}
}

// Array returns a new Field with the given key and ArrayEncoder.
func Array(k string, v ArrayEncoder) Field {
	return Field{Key: k, Type: FieldTypeArray, Any: v}
}

// Object returns a new Field with the given key and ObjectEncoder.
func Object(k string, v ObjectEncoder) Field {
	return Field{Key: k, Type: FieldTypeObject, Any: v}
}

// ConstStringer returns a new Field with the given key and Stringer.
// Call ConstStringer if your object is const. It has significantly less
// impact on the calling goroutine.
func ConstStringer(k string, v fmt.Stringer) Field {
	return Field{Key: k, Type: FieldTypeStringer, Any: v}
}

// Stringer returns a new Field with the given key and Stringer.
func Stringer(k string, v fmt.Stringer) Field {
	if v == nil {
		return String(k, "nil")
	}

	return String(k, v.String())
}

// ConstFormatter returns a new Field with the given key, verb and interface
// to format.
//
// Call ConstFormatter if your object is const. It has significantly less
// impact on the calling goroutine.
func ConstFormatter(k string, verb string, v interface{}) Field {
	return Field{Key: k, Type: FieldTypeFormatter, Bytes: *(*[]byte)(unsafe.Pointer(&verb)), Any: v}
}

// ConstFormatterV returns a new Field with the given key and interface to
// format. It uses the predefined verb "%#v" (a Go-syntax representation of
// the value).
//
// Call ConstFormatterV if your object is const. It has significantly less
// impact on the calling goroutine.
func ConstFormatterV(k string, v interface{}) Field {
	return ConstFormatter(k, "%#v", v)
}

// Formatter returns a new Field with the given key, verb and interface to
// format.
func Formatter(k string, verb string, v interface{}) Field {
	return String(k, fmt.Sprintf(verb, v))
}

// FormatterV returns a new Field with the given key and interface to format.
// It uses the predefined verb "%#v" (a Go-syntax representation of the value).
func FormatterV(k string, v interface{}) Field {
	return Formatter(k, "%#v", v)
}

// Any returns a new Filed with the given key and value of any type. Is tries
// to choose the best way to represent key-value pair as a Field.
//
// Note that Any is not possible to choose ConstX methods. Use specific Field
// methods for better performance.
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
	case string:
		return String(k, rv)
	case nil:
		break

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

// FieldType specifies how to handle Field data.
type FieldType byte

// Set of FileType values.
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
	FieldTypeStringer
	FieldTypeFormatter
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

// Field hold data of a specific field.
type Field struct {
	Key   string
	Type  FieldType
	Any   interface{}
	Int   int64
	Bytes []byte
}

// Accept interprets Field data according to FieldType and calls appropriate
// FieldEncoder function.
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
		if fd.Any != nil {
			v.EncodeFieldArray(fd.Key, fd.Any.(ArrayEncoder))
		} else {
			v.EncodeFieldString(fd.Key, "nil")
		}
	case FieldTypeObject:
		if fd.Any != nil {
			v.EncodeFieldObject(fd.Key, fd.Any.(ObjectEncoder))
		} else {
			v.EncodeFieldString(fd.Key, "nil")
		}
	case FieldTypeStringer:
		if fd.Any != nil {
			v.EncodeFieldString(fd.Key, (fd.Any.(fmt.Stringer)).String())
		} else {
			v.EncodeFieldString(fd.Key, "nil")
		}
	case FieldTypeFormatter:
		v.EncodeFieldString(fd.Key, fmt.Sprintf(*(*string)(unsafe.Pointer(&fd.Bytes)), fd.Any))
	case FieldTypeBytes, FieldTypeRawBytes:
		v.EncodeFieldBytes(fd.Key, fd.Bytes)
	case FieldTypeBytesToString:
		v.EncodeFieldString(fd.Key, *(*string)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToBools, FieldTypeRawBytesToBools:
		v.EncodeFieldBools(fd.Key, *(*[]bool)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToInts64, FieldTypeRawBytesToInts64:
		v.EncodeFieldInts64(fd.Key, *(*[]int64)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToInts32, FieldTypeRawBytesToInts32:
		v.EncodeFieldInts32(fd.Key, *(*[]int32)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToInts16, FieldTypeRawBytesToInts16:
		v.EncodeFieldInts16(fd.Key, *(*[]int16)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToInts8, FieldTypeRawBytesToInts8:
		v.EncodeFieldInts8(fd.Key, *(*[]int8)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToUints64, FieldTypeRawBytesToUints64:
		v.EncodeFieldUints64(fd.Key, *(*[]uint64)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToUints32, FieldTypeRawBytesToUints32:
		v.EncodeFieldUints32(fd.Key, *(*[]uint32)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToUints16, FieldTypeRawBytesToUints16:
		v.EncodeFieldUints16(fd.Key, *(*[]uint16)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToUints8, FieldTypeRawBytesToUints8:
		v.EncodeFieldUints8(fd.Key, *(*[]uint8)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToFloats64, FieldTypeRawBytesToFloats64:
		v.EncodeFieldFloats64(fd.Key, *(*[]float64)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToFloats32, FieldTypeRawBytesToFloats32:
		v.EncodeFieldFloats32(fd.Key, *(*[]float32)(unsafe.Pointer(&fd.Bytes)))
	case FieldTypeBytesToDurations, FieldTypeRawBytesToDurations:
		v.EncodeFieldDurations(fd.Key, *(*[]time.Duration)(unsafe.Pointer(&fd.Bytes)))
	}
}
