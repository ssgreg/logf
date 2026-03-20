package logf

import (
	"time"
	"unsafe"
)

// Encoder is the interface that turns an Entry into bytes — it decides
// your log format (JSON, text, or whatever you dream up). The built-in
// JSON and Text encoders handle most needs, but implementing Encoder
// lets you go fully custom.
//
// Encode serializes the Entry and returns a pooled *Buffer. The caller
// must call Buffer.Free when done. Encode is safe for concurrent use —
// implementations handle internal cloning and buffer pooling.
//
// Clone returns an independent copy that shares immutable config but
// has its own mutable state, suitable for use in another goroutine.
type Encoder interface {
	Encode(Entry) (*Buffer, error)
	Clone() Encoder
}

// ArrayEncoder lets your custom types serialize themselves as JSON arrays
// (or whatever array representation the encoder uses). Implement
// EncodeLogfArray and pass your type to logf.Array().
//
// Example:
//
//	type stringArray []string
//
//	func (o stringArray) EncodeLogfArray(e TypeEncoder) error {
//		for i := range o {
//			e.EncodeTypeString(o[i])
//		}
//		return nil
//	}
type ArrayEncoder interface {
	EncodeLogfArray(TypeEncoder) error
}

// ObjectEncoder lets your custom types serialize themselves as structured
// objects with named fields. This is how you get zero-allocation logging
// for your domain types — no reflection, no fmt.Sprintf, just direct
// calls to the encoder.
//
// Example:
//
//	type user struct {
//		Username string
//		Password string
//	}
//
//	func (u user) EncodeLogfObject(e FieldEncoder) error {
//		e.EncodeFieldString("username", u.Username)
//		e.EncodeFieldString("password", u.Password)
//		return nil
//	}
type ObjectEncoder interface {
	EncodeLogfObject(FieldEncoder) error
}

// TypeEncoder provides methods for encoding individual values (scalars,
// slices, arrays, objects) without field names. It is the companion
// interface used by TimeEncoder, DurationEncoder, LevelEncoder, and
// CallerEncoder to write their output into the buffer.
type TypeEncoder interface {
	EncodeTypeAny(interface{})
	EncodeTypeBool(bool)
	EncodeTypeInt64(int64)
	EncodeTypeUint64(uint64)
	EncodeTypeFloat64(float64)
	EncodeTypeDuration(time.Duration)
	EncodeTypeTime(time.Time)
	EncodeTypeString(string)
	EncodeTypeStrings([]string)
	EncodeTypeBytes([]byte)
	EncodeTypeInts64([]int64)
	EncodeTypeFloats64([]float64)
	EncodeTypeDurations([]time.Duration)
	EncodeTypeArray(ArrayEncoder)
	EncodeTypeObject(ObjectEncoder)
	EncodeTypeUnsafeBytes(unsafe.Pointer)
}

// FieldEncoder provides methods for encoding key-value pairs. It is the
// interface that ObjectEncoder and ErrorEncoder receive to write named
// fields into the output. Each method encodes one field with the given
// key and typed value.
type FieldEncoder interface {
	EncodeFieldAny(string, interface{})
	EncodeFieldBool(string, bool)
	EncodeFieldInt64(string, int64)
	EncodeFieldUint64(string, uint64)
	EncodeFieldFloat64(string, float64)
	EncodeFieldDuration(string, time.Duration)
	EncodeFieldError(string, error)
	EncodeFieldTime(string, time.Time)
	EncodeFieldString(string, string)
	EncodeFieldStrings(string, []string)
	EncodeFieldBytes(string, []byte)
	EncodeFieldInts64(string, []int64)
	EncodeFieldFloats64(string, []float64)
	EncodeFieldDurations(string, []time.Duration)
	EncodeFieldArray(string, ArrayEncoder)
	EncodeFieldObject(string, ObjectEncoder)
	EncodeFieldGroup(string, []Field)
}

// TypeEncoderFactory creates a TypeEncoder that writes into the given Buffer.
// This lets one encoder borrow another encoder's formatting — for example,
// the text encoder uses the JSON encoder's TypeEncoderFactory to render
// nested objects and arrays in JSON syntax within otherwise plain-text output.
type TypeEncoderFactory interface {
	TypeEncoder(*Buffer) TypeEncoder
}

// EncoderBuilder builds an Encoder from accumulated configuration.
// Implemented by JSONEncoderBuilder and TextEncoderBuilder, and accepted
// by LoggerBuilder.EncoderFrom for composable builder chains.
type EncoderBuilder interface {
	Build() Encoder
}
