package logf

import (
	"fmt"
	"time"
	"unsafe"
)

// testAppender
type testAppender struct {
	Entries          []Entry
	FlushCallCounter int
	SyncCallCounter  int

	AppendError error
	FlushError  error
	SyncError   error
}

func (a *testAppender) Append(e Entry) error {
	if a.AppendError != nil {
		return a.AppendError
	}
	a.Entries = append(a.Entries, e)

	return nil
}

func (a *testAppender) Flush() error {
	if a.FlushError != nil {
		return a.FlushError
	}
	a.FlushCallCounter++

	return nil
}

func (a *testAppender) Sync() error {
	if a.SyncError != nil {
		return a.SyncError
	}
	a.SyncCallCounter++

	return nil
}

// testEntryWriter implements EntryWriter storing the last Entry.
type testEntryWriter struct {
	Entry *Entry
}

func (w *testEntryWriter) WriteEntry(e Entry) {
	w.Entry = &e
}

// testSnapshotter implements Snapshotter allowing to check whether
// TakeSnapshot was called or not. TakeSnapshot returns new object of
// this type.
type testSnapshotter struct {
	Called bool
}

func (s *testSnapshotter) TakeSnapshot() interface{} {
	s.Called = true

	return &testSnapshotter{}
}

// testLevelCheckerReturningFalse implements LevelCheckerGetter that always
// returns false.
type testLevelCheckerReturningFalse struct {
}

func (g testLevelCheckerReturningFalse) LevelChecker() LevelChecker {
	return func(Level) bool {
		return false
	}
}

// testTypeEncoder implements TypeEncoder storing the last encoding value.
type testTypeEncoder struct {
	result interface{}
}

func (e *testTypeEncoder) EncodeTypeAny(v interface{}) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeBool(v bool) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeInt64(v int64) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeInt32(v int32) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeInt16(v int16) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeInt8(v int8) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeUint64(v uint64) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeUint32(v uint32) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeUint16(v uint16) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeUint8(v uint8) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeFloat64(v float64) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeFloat32(v float32) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeDuration(v time.Duration) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeTime(v time.Time) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeString(v string) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeBytes(v []byte) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeBools(v []bool) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeInts64(v []int64) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeInts32(v []int32) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeInts16(v []int16) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeInts8(v []int8) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeUints64(v []uint64) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeUints32(v []uint32) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeUints16(v []uint16) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeUints8(v []uint8) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeFloats64(v []float64) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeFloats32(v []float32) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeDurations(v []time.Duration) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeArray(v ArrayEncoder) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeObject(v ObjectEncoder) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeUnsafeBytes(v unsafe.Pointer) {
	e.result = string(*(*[]byte)(v))
}

func newTestFieldEncoder() *testFieldEncoder {
	return &testFieldEncoder{make(map[string]interface{})}
}

// testFieldEncoder implements FieldEncoder storing all fields to be encoded.
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

func (e *testFieldEncoder) EncodeFieldStrings(k string, v []string) {
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

type testObjectEncoder struct{}

func (o testObjectEncoder) EncodeLogfObject(e FieldEncoder) error {
	e.EncodeFieldString("username", "username")
	e.EncodeFieldInt64("code", 42)

	return nil
}

type testArrayEncoder struct{}

func (o testArrayEncoder) EncodeLogfArray(e TypeEncoder) error {
	e.EncodeTypeInt64(42)

	return nil
}

type testStringer struct {
	result string
}

func (s testStringer) String() string {
	return s.result
}

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
