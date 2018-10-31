package logf

import (
	"time"
	"unsafe"
)

type ArrayMarshaller interface {
	MarshalLogfArray(TypeMarshaller) error
}

type ObjectMarshaller interface {
	MarshalLogfObject(FieldMarshaller) error
}

type TypeMarshaller interface {
	MarshalAny(interface{})

	MarshalBool(bool)
	MarshalInt64(int64)
	MarshalInt32(int32)
	MarshalInt16(int16)
	MarshalInt8(int8)
	MarshalUint64(uint64)
	MarshalUint32(uint32)
	MarshalUint16(uint16)
	MarshalUint8(uint8)
	MarshalFloat64(float64)
	MarshalFloat32(float32)

	MarshalDuration(time.Duration)
	MarshalTime(time.Time)

	MarshalString(string)
	MarshalBytes([]byte)
	MarshalBools([]bool)
	MarshalInts64([]int64)
	MarshalInts32([]int32)
	MarshalInts16([]int16)
	MarshalInts8([]int8)
	MarshalUints64([]uint64)
	MarshalUints32([]uint32)
	MarshalUints16([]uint16)
	MarshalUints8([]uint8)
	MarshalFloats64([]float64)
	MarshalFloats32([]float32)
	MarshalDurations([]time.Duration)

	MarshalArray(ArrayMarshaller)
	MarshalObject(ObjectMarshaller)

	MarshalUnsafeBytes(unsafe.Pointer)
}

type FieldMarshaller interface {
	MarshalFieldAny(string, interface{})

	MarshalFieldBool(string, bool)
	MarshalFieldInt64(string, int64)
	MarshalFieldInt32(string, int32)
	MarshalFieldInt16(string, int16)
	MarshalFieldInt8(string, int8)
	MarshalFieldUint64(string, uint64)
	MarshalFieldUint32(string, uint32)
	MarshalFieldUint16(string, uint16)
	MarshalFieldUint8(string, uint8)
	MarshalFieldFloat64(string, float64)
	MarshalFieldFloat32(string, float32)

	MarshalFieldDuration(string, time.Duration)
	MarshalFieldError(string, error)
	MarshalFieldTime(string, time.Time)

	MarshalFieldString(string, string)
	MarshalFieldBytes(string, []byte)
	MarshalFieldBools(string, []bool)
	MarshalFieldInts64(string, []int64)
	MarshalFieldInts32(string, []int32)
	MarshalFieldInts16(string, []int16)
	MarshalFieldInts8(string, []int8)
	MarshalFieldUints64(string, []uint64)
	MarshalFieldUints32(string, []uint32)
	MarshalFieldUints16(string, []uint16)
	MarshalFieldUints8(string, []uint8)
	MarshalFieldFloats64(string, []float64)
	MarshalFieldFloats32(string, []float32)
	MarshalFieldDurations(string, []time.Duration)

	MarshalFieldArray(string, ArrayMarshaller)
	MarshalFieldObject(string, ObjectMarshaller)
}

type TypeMarshallerFactory interface {
	TypeMarshaller(*Buffer) TypeMarshaller
}
