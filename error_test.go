package logf

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type verboseError struct {
	short string
	full  string
}

func (e *verboseError) Error() string {
	return e.short
}

func (e *verboseError) Format(f fmt.State, c rune) {
	f.Write([]byte(e.full))
}

func TestDefaultErrorEncoderWithPlainError(t *testing.T) {
	e := errors.New("simple error")
	enc := newTestFieldEncoder()
	DefaultErrorEncoder("error", e, enc)

	assert.EqualValues(t, e.Error(), enc.result["error"])
	assert.EqualValues(t, 1, len(enc.result))
}

func TestDefaultErrorEncoderWithVerboseError(t *testing.T) {
	e := &verboseError{"short", "verbose"}
	enc := newTestFieldEncoder()
	DefaultErrorEncoder("error", e, enc)

	assert.EqualValues(t, 2, len(enc.result))
	assert.EqualValues(t, e.short, enc.result["error"])
	assert.EqualValues(t, e.full, enc.result["error.verbose"])
}

func TestDefaultErrorEncoderWithNil(t *testing.T) {
	enc := newTestFieldEncoder()
	DefaultErrorEncoder("error", nil, enc)

	assert.EqualValues(t, 1, len(enc.result))
	assert.EqualValues(t, "<nil>", enc.result["error"])
}

func newTestFieldEncoder() *testFieldEncoder {
	return &testFieldEncoder{make(map[string]interface{})}
}

type testFieldEncoder struct {
	result map[string]interface{}
}

func (e *testFieldEncoder) EncodeFieldAny(k string, v interface{}) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldBool(k string, v bool) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldInt64(k string, v int64) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldInt32(k string, v int32) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldInt16(k string, v int16) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldInt8(k string, v int8) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldUint64(k string, v uint64) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldUint32(k string, v uint32) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldUint16(k string, v uint16) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldUint8(k string, v uint8) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldFloat64(k string, v float64) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldFloat32(k string, v float32) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldDuration(k string, v time.Duration) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldError(k string, v error) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldTime(k string, v time.Time) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldString(k string, v string) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldBytes(k string, v []byte) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldBools(k string, v []bool) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldInts64(k string, v []int64) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldInts32(k string, v []int32) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldInts16(k string, v []int16) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldInts8(k string, v []int8) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldUints64(k string, v []uint64) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldUints32(k string, v []uint32) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldUints16(k string, v []uint16) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldUints8(k string, v []uint8) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldFloats64(k string, v []float64) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldFloats32(k string, v []float32) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldDurations(k string, v []time.Duration) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldArray(k string, v ArrayEncoder) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldObject(k string, v ObjectEncoder) {
	e.result[k] = v
}
