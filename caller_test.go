package logf

import (
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestEntryCallerFileWithPackage(t *testing.T) {
	cases := []struct {
		caller EntryCaller
		golden string
	}{
		{
			caller: EntryCaller{0, "/a/b/c/d.go", 66, true},
			golden: "c/d.go",
		},
		{
			caller: EntryCaller{0, "c/d.go", 66, true},
			golden: "c/d.go",
		},
		{
			caller: EntryCaller{0, "d.go", 66, true},
			golden: "d.go",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.golden, c.caller.FileWithPackage())
	}
}

func TestEntryCaller(t *testing.T) {
	caller := NewEntryCaller(0)

	assert.True(t, caller.Specified)
	assert.True(t, caller.Line > 0 && caller.Line < 1000)
	assert.Equal(t, "logf/caller_test.go", caller.FileWithPackage())
	assert.Contains(t, caller.File, "/logf/caller_test.go")
}

func TestShortCallerEncoder(t *testing.T) {
	enc := testingTypeEncoder{}
	caller := EntryCaller{0, "/a/b/c/d.go", 66, true}
	ShortCallerEncoder(caller, &enc)

	assert.EqualValues(t, "c/d.go:66", enc.result)
}

func TestFullCallerEncoder(t *testing.T) {
	enc := testingTypeEncoder{}
	caller := EntryCaller{0, "/a/b/c/d.go", 66, true}
	FullCallerEncoder(caller, &enc)

	assert.EqualValues(t, "/a/b/c/d.go:66", enc.result)
}

type testingTypeEncoder struct {
	result interface{}
}

func (e *testingTypeEncoder) EncodeTypeAny(v interface{}) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeBool(v bool) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeInt64(v int64) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeInt32(v int32) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeInt16(v int16) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeInt8(v int8) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeUint64(v uint64) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeUint32(v uint32) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeUint16(v uint16) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeUint8(v uint8) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeFloat64(v float64) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeFloat32(v float32) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeDuration(v time.Duration) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeTime(v time.Time) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeString(v string) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeBytes(v []byte) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeBools(v []bool) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeInts64(v []int64) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeInts32(v []int32) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeInts16(v []int16) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeInts8(v []int8) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeUints64(v []uint64) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeUints32(v []uint32) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeUints16(v []uint16) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeUints8(v []uint8) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeFloats64(v []float64) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeFloats32(v []float32) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeDurations(v []time.Duration) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeArray(v ArrayEncoder) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeObject(v ObjectEncoder) {
	e.result = v
}

func (e *testingTypeEncoder) EncodeTypeUnsafeBytes(v unsafe.Pointer) {
	e.result = *(*string)(v)
}
