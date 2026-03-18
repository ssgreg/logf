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

	return Field{Key: k, Type: FieldTypeBool, Val: tmp}
}

// Int returns a new Field with the given key and int.
func Int(k string, v int) Field {
	return Field{Key: k, Type: FieldTypeInt64, Val: int64(v)}
}

// Int64 returns a new Field with the given key and int64.
func Int64(k string, v int64) Field {
	return Field{Key: k, Type: FieldTypeInt64, Val: v}
}

// Int32 returns a new Field with the given key and int32.
func Int32(k string, v int32) Field {
	return Field{Key: k, Type: FieldTypeInt64, Val: int64(v)}
}

// Int16 returns a new Field with the given key and int16.
func Int16(k string, v int16) Field {
	return Field{Key: k, Type: FieldTypeInt64, Val: int64(v)}
}

// Int8 returns a new Field with the given key and int.8
func Int8(k string, v int8) Field {
	return Field{Key: k, Type: FieldTypeInt64, Val: int64(v)}
}

// Uint returns a new Field with the given key and uint.
func Uint(k string, v uint) Field {
	return Field{Key: k, Type: FieldTypeUint64, Val: int64(v)}
}

// Uint64 returns a new Field with the given key and uint64.
func Uint64(k string, v uint64) Field {
	return Field{Key: k, Type: FieldTypeUint64, Val: int64(v)}
}

// Uint32 returns a new Field with the given key and uint32.
func Uint32(k string, v uint32) Field {
	return Field{Key: k, Type: FieldTypeUint64, Val: int64(v)}
}

// Uint16 returns a new Field with the given key and uint16.
func Uint16(k string, v uint16) Field {
	return Field{Key: k, Type: FieldTypeUint64, Val: int64(v)}
}

// Uint8 returns a new Field with the given key and uint8.
func Uint8(k string, v uint8) Field {
	return Field{Key: k, Type: FieldTypeUint64, Val: int64(v)}
}

// Float64 returns a new Field with the given key and float64.
func Float64(k string, v float64) Field {
	return Field{Key: k, Type: FieldTypeFloat64, Val: int64(math.Float64bits(v))}
}

// Float32 returns a new Field with the given key and float32.
func Float32(k string, v float32) Field {
	return Field{Key: k, Type: FieldTypeFloat64, Val: int64(math.Float64bits(float64(v)))}
}

// Duration returns a new Field with the given key and time.Duration.
func Duration(k string, v time.Duration) Field {
	return Field{Key: k, Type: FieldTypeDuration, Val: int64(v)}
}

// Bytes returns a new Field with the given key and slice of bytes.
func Bytes(k string, v []byte) Field {
	return Field{Key: k, Type: FieldTypeBytes, Ptr: unsafe.Pointer(unsafe.SliceData(v)), Val: int64(len(v))}
}

// String returns a new Field with the given key and string.
func String(k string, v string) Field {
	return Field{Key: k, Type: FieldTypeBytesToString, Ptr: unsafe.Pointer(unsafe.StringData(v)), Val: int64(len(v))}
}

// Strings returns a new Field with the given key and slice of strings.
func Strings(k string, v []string) Field {
	return Field{Key: k, Type: FieldTypeBytesToStrings, Ptr: unsafe.Pointer(unsafe.SliceData(v)), Val: int64(len(v))}
}

// Ints returns a new Field with the given key and slice of ints.
func Ints(k string, v []int) Field {
	if unsafe.Sizeof(int(0)) == unsafe.Sizeof(int64(0)) {
		return Field{Key: k, Type: FieldTypeBytesToInts64, Ptr: unsafe.Pointer(unsafe.SliceData(v)), Val: int64(len(v))}
	}
	// 32-bit platform: int is 4 bytes, cannot reinterpret as []int64.
	s := make([]int64, len(v))
	for i, x := range v {
		s[i] = int64(x)
	}
	return Field{Key: k, Type: FieldTypeBytesToInts64, Ptr: unsafe.Pointer(unsafe.SliceData(s)), Val: int64(len(s))}
}

// Ints64 returns a new Field with the given key and slice of 64-bit ints.
func Ints64(k string, v []int64) Field {
	return Field{Key: k, Type: FieldTypeBytesToInts64, Ptr: unsafe.Pointer(unsafe.SliceData(v)), Val: int64(len(v))}
}

// Floats64 returns a new Field with the given key and slice of 64-bit floats.
func Floats64(k string, v []float64) Field {
	return Field{Key: k, Type: FieldTypeBytesToFloats64, Ptr: unsafe.Pointer(unsafe.SliceData(v)), Val: int64(len(v))}
}

// Durations returns a new Field with the given key and slice of time.Duration.
func Durations(k string, v []time.Duration) Field {
	return Field{Key: k, Type: FieldTypeBytesToDurations, Ptr: unsafe.Pointer(unsafe.SliceData(v)), Val: int64(len(v))}
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
	if v.IsZero() {
		return Field{Key: k, Type: FieldTypeTime}
	}
	return Field{Key: k, Type: FieldTypeTime, Val: v.UnixNano(), Any: v.Location()}
}

// Array returns a new Field with the given key and ArrayEncoder.
func Array(k string, v ArrayEncoder) Field {
	return Field{Key: k, Type: FieldTypeArray, Any: v}
}

// Object returns a new Field with the given key and ObjectEncoder.
func Object(k string, v ObjectEncoder) Field {
	return Field{Key: k, Type: FieldTypeObject, Any: v}
}

// Inline returns a new Field that encodes the given ObjectEncoder's fields
// directly into the parent object, without a wrapping key.
//
// Example:
//
//	logger.Info(ctx, "request handled",
//	    logf.Inline(requestInfo),
//	    logf.Int("status", 200),
//	)
//	// → {"msg":"request handled", "trace_id":"abc", "method":"GET", "status":200}
func Inline(v ObjectEncoder) Field {
	return Object("", v)
}

// Group returns a new Field that encodes the given fields as a nested
// object under the given key.
//
// Example:
//
//	logger.Info(ctx, "done",
//	    logf.Group("request", logf.String("id", "abc"), logf.Int("status", 200)),
//	)
//	// → {"msg":"done", "request":{"id":"abc", "status":200}}
func Group(k string, fs ...Field) Field {
	return Field{Key: k, Type: FieldTypeGroup, Any: fs}
}

// Stringer returns a new Field with the given key and Stringer.
func Stringer(k string, v fmt.Stringer) Field {
	if v == nil {
		return String(k, "nil")
	}

	return String(k, v.String())
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

// ByteString returns a new Field with the given key and []byte that is
// interpreted as a UTF-8 string (not base64-encoded like Bytes).
func ByteString(k string, v []byte) Field {
	return String(k, unsafe.String(unsafe.SliceData(v), len(v)))
}

// Any returns a new Filed with the given key and value of any type. Is tries
// to choose the best way to represent key-value pair as a Field.
//
// Note that Any may not choose the most efficient typed method for every type.
// Use specific Field methods for better performance.
func Any(k string, v interface{}) Field {
	switch rv := v.(type) {
	// Scalars.
	case bool:
		return Bool(k, rv)
	case string:
		return String(k, rv)
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

	// Scalar pointers (nil → FieldTypeAny → null).
	case *bool:
		if rv != nil {
			return Bool(k, *rv)
		}
	case *string:
		if rv != nil {
			return String(k, *rv)
		}
	case *int:
		if rv != nil {
			return Int(k, *rv)
		}
	case *int64:
		if rv != nil {
			return Int64(k, *rv)
		}
	case *int32:
		if rv != nil {
			return Int32(k, *rv)
		}
	case *uint:
		if rv != nil {
			return Uint(k, *rv)
		}
	case *uint64:
		if rv != nil {
			return Uint64(k, *rv)
		}
	case *uint32:
		if rv != nil {
			return Uint32(k, *rv)
		}
	case *float64:
		if rv != nil {
			return Float64(k, *rv)
		}
	case *float32:
		if rv != nil {
			return Float32(k, *rv)
		}
	case *time.Time:
		if rv != nil {
			return Time(k, *rv)
		}
	case *time.Duration:
		if rv != nil {
			return Duration(k, *rv)
		}

	// Slices.
	case []byte:
		return Bytes(k, rv)
	case []string:
		return Strings(k, rv)
	case []int:
		return Ints(k, rv)
	case []int64:
		return Ints64(k, rv)
	case []float64:
		return Floats64(k, rv)
	case []time.Duration:
		return Durations(k, rv)

	// Interface-based.
	case ArrayEncoder:
		return Array(k, rv)
	case ObjectEncoder:
		return Object(k, rv)
	// fmt.Stringer MUST stay after concrete types (time.Duration, error, etc.)
	// and before default. Moving it earlier would shadow typed handling;
	// moving it into default would let reflect.String leak raw values
	// from types that implement Stringer for masking (e.g. secrets).
	case fmt.Stringer:
		return String(k, rv.String())

	case nil:
		break

	default:
		val := reflect.ValueOf(rv)
		switch val.Kind() {
		case reflect.String:
			return String(k, val.String())
		case reflect.Bool:
			return Bool(k, val.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if val.Type() == reflect.TypeOf(time.Duration(0)) {
				return Duration(k, time.Duration(val.Int()))
			}
			return Int64(k, val.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return Uint64(k, val.Uint())
		case reflect.Float32, reflect.Float64:
			return Float64(k, val.Float())
		}
	}

	return Field{Key: k, Type: FieldTypeAny, Any: v}
}

// FieldType specifies how to handle Field data.
type FieldType byte

// Set of FileType values.
const (
	FieldTypeUnknown FieldType = iota

	// Scalars (value stored in Val/Any).
	FieldTypeAny
	FieldTypeBool
	FieldTypeInt64
	FieldTypeUint64
	FieldTypeFloat64
	FieldTypeDuration
	FieldTypeError
	FieldTypeTime

	// Unsafe pointer slices (data in Ptr, length in Val).
	FieldTypeBytes
	FieldTypeBytesToString
	FieldTypeBytesToInts64
	FieldTypeBytesToFloats64
	FieldTypeBytesToDurations
	FieldTypeBytesToStrings

	// Interface-based (encoder callback in Any).
	FieldTypeArray
	FieldTypeObject
	FieldTypeGroup
)

// Field hold data of a specific field.
//
// Layout (56 bytes):
//
//	Key  string         // 16  field name
//	Type FieldType      //  8  (1 byte + 7 padding)
//	Any  interface{}    // 16  error, object, array, stringer, any
//	Ptr  unsafe.Pointer //  8  slice/string data pointer
//	Val  int64          //  8  scalar value OR slice/string length
type Field struct {
	Key  string
	Type FieldType
	Any  interface{}
	Ptr  unsafe.Pointer
	Val  int64
}

// Accept interprets Field data according to FieldType and calls appropriate
// FieldEncoder function.
func (fd Field) Accept(v FieldEncoder) {
	switch fd.Type {
	case FieldTypeAny:
		v.EncodeFieldAny(fd.Key, fd.Any)
	case FieldTypeBool:
		v.EncodeFieldBool(fd.Key, fd.Val != 0)
	case FieldTypeInt64:
		v.EncodeFieldInt64(fd.Key, fd.Val)
	case FieldTypeUint64:
		v.EncodeFieldUint64(fd.Key, uint64(fd.Val))
	case FieldTypeFloat64:
		v.EncodeFieldFloat64(fd.Key, math.Float64frombits(uint64(fd.Val)))
	case FieldTypeDuration:
		v.EncodeFieldDuration(fd.Key, time.Duration(fd.Val))
	case FieldTypeError:
		if fd.Any != nil {
			v.EncodeFieldError(fd.Key, fd.Any.(error))
		} else {
			v.EncodeFieldError(fd.Key, nil)
		}
	case FieldTypeTime:
		if fd.Any != nil {
			v.EncodeFieldTime(fd.Key, time.Unix(0, fd.Val).In(fd.Any.(*time.Location)))
		} else if fd.Val != 0 {
			v.EncodeFieldTime(fd.Key, time.Unix(0, fd.Val))
		} else {
			v.EncodeFieldTime(fd.Key, time.Time{})
		}
	case FieldTypeBytes:
		v.EncodeFieldBytes(fd.Key, unsafe.Slice((*byte)(fd.Ptr), int(fd.Val)))
	case FieldTypeBytesToString:
		v.EncodeFieldString(fd.Key, unsafe.String((*byte)(fd.Ptr), int(fd.Val)))
	case FieldTypeBytesToInts64:
		v.EncodeFieldInts64(fd.Key, unsafe.Slice((*int64)(fd.Ptr), int(fd.Val)))
	case FieldTypeBytesToFloats64:
		v.EncodeFieldFloats64(fd.Key, unsafe.Slice((*float64)(fd.Ptr), int(fd.Val)))
	case FieldTypeBytesToDurations:
		v.EncodeFieldDurations(fd.Key, unsafe.Slice((*time.Duration)(fd.Ptr), int(fd.Val)))
	case FieldTypeBytesToStrings:
		v.EncodeFieldStrings(fd.Key, unsafe.Slice((*string)(fd.Ptr), int(fd.Val)))
	case FieldTypeArray:
		v.EncodeFieldArray(fd.Key, fd.Any.(ArrayEncoder))
	case FieldTypeObject:
		if fd.Key == "" {
			_ = (fd.Any.(ObjectEncoder)).EncodeLogfObject(v)
		} else {
			v.EncodeFieldObject(fd.Key, fd.Any.(ObjectEncoder))
		}
	case FieldTypeGroup:
		if fd.Any != nil {
			v.EncodeFieldGroup(fd.Key, fd.Any.([]Field))
		}
	default:
		panic(fmt.Sprintf("logf: unknown FieldType %d", fd.Type))
	}
}
