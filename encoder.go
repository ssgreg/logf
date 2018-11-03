package logf

import (
	"time"
	"unsafe"
)

// Encoder defines the interface to create your own log format.
//
// In case of error, Encoder must remove bytes related to the given Entry
// from the Buffer.
type Encoder interface {
	Encode(*Buffer, Entry) error
}

type ArrayEncoder interface {
	EncodeLogfArray(TypeEncoder) error
}

type ObjectEncoder interface {
	EncodeLogfObject(FieldEncoder) error
}

type TypeEncoder interface {
	EncodeTypeAny(interface{})

	EncodeTypeBool(bool)
	EncodeTypeInt64(int64)
	EncodeTypeInt32(int32)
	EncodeTypeInt16(int16)
	EncodeTypeInt8(int8)
	EncodeTypeUint64(uint64)
	EncodeTypeUint32(uint32)
	EncodeTypeUint16(uint16)
	EncodeTypeUint8(uint8)
	EncodeTypeFloat64(float64)
	EncodeTypeFloat32(float32)

	EncodeTypeDuration(time.Duration)
	EncodeTypeTime(time.Time)

	EncodeTypeString(string)
	EncodeTypeBytes([]byte)
	EncodeTypeBools([]bool)
	EncodeTypeInts64([]int64)
	EncodeTypeInts32([]int32)
	EncodeTypeInts16([]int16)
	EncodeTypeInts8([]int8)
	EncodeTypeUints64([]uint64)
	EncodeTypeUints32([]uint32)
	EncodeTypeUints16([]uint16)
	EncodeTypeUints8([]uint8)
	EncodeTypeFloats64([]float64)
	EncodeTypeFloats32([]float32)
	EncodeTypeDurations([]time.Duration)

	EncodeTypeArray(ArrayEncoder)
	EncodeTypeObject(ObjectEncoder)

	EncodeTypeUnsafeBytes(unsafe.Pointer)
}

type FieldEncoder interface {
	EncodeFieldAny(string, interface{})

	EncodeFieldBool(string, bool)
	EncodeFieldInt64(string, int64)
	EncodeFieldInt32(string, int32)
	EncodeFieldInt16(string, int16)
	EncodeFieldInt8(string, int8)
	EncodeFieldUint64(string, uint64)
	EncodeFieldUint32(string, uint32)
	EncodeFieldUint16(string, uint16)
	EncodeFieldUint8(string, uint8)
	EncodeFieldFloat64(string, float64)
	EncodeFieldFloat32(string, float32)

	EncodeFieldDuration(string, time.Duration)
	EncodeFieldError(string, error)
	EncodeFieldTime(string, time.Time)

	EncodeFieldString(string, string)
	EncodeFieldBytes(string, []byte)
	EncodeFieldBools(string, []bool)
	EncodeFieldInts64(string, []int64)
	EncodeFieldInts32(string, []int32)
	EncodeFieldInts16(string, []int16)
	EncodeFieldInts8(string, []int8)
	EncodeFieldUints64(string, []uint64)
	EncodeFieldUints32(string, []uint32)
	EncodeFieldUints16(string, []uint16)
	EncodeFieldUints8(string, []uint8)
	EncodeFieldFloats64(string, []float64)
	EncodeFieldFloats32(string, []float32)
	EncodeFieldDurations(string, []time.Duration)

	EncodeFieldArray(string, ArrayEncoder)
	EncodeFieldObject(string, ObjectEncoder)
}

type TypeEncoderFactory interface {
	TypeEncoder(*Buffer) TypeEncoder
}
