package logf

import (
	"errors"
	"fmt"
	"testing"

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
