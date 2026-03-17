package logf

import (
	"context"
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

// testHandler implements Handler storing all entries.
type testHandler struct {
	Entry      *Entry
	Entries    []Entry
	level      Level
	checkLevel bool
}

func newLeveledTestHandler(level Level) *testHandler {
	return &testHandler{level: level, checkLevel: true}
}

func (w *testHandler) Handle(_ context.Context, e Entry) error {
	w.Entry = &e
	w.Entries = append(w.Entries, e)
	return nil
}

func (w *testHandler) Enabled(_ context.Context, lvl Level) bool {
	if w.checkLevel {
		return w.level.Enabled(lvl)
	}
	return true
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

func (e *testTypeEncoder) EncodeTypeUint64(v uint64) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeFloat64(v float64) {
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

func (e *testTypeEncoder) EncodeTypeStrings(v []string) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeBytes(v []byte) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeInts64(v []int64) {
	e.result = v
}

func (e *testTypeEncoder) EncodeTypeFloats64(v []float64) {
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

func (e *testFieldEncoder) EncodeFieldUint64(k string, v uint64) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldFloat64(k string, v float64) {
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

func (e *testFieldEncoder) EncodeFieldInts64(k string, v []int64) {
	e.result[k] = v
}

func (e *testFieldEncoder) EncodeFieldFloats64(k string, v []float64) {
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

func (e *testFieldEncoder) EncodeFieldGroup(k string, fs []Field) {
	e.result[k] = fs
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
	_, _ = f.Write([]byte(e.full))
}
