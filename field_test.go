package logf

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestField(t *testing.T) {
	cases := []struct {
		name     string
		fn       func(interface{}) Field
		original interface{}
		expected interface{}
	}{
		{
			name:     "Bool",
			fn:       func(v interface{}) Field { return Bool("k", v.(bool)) },
			original: true,
			expected: true,
		},
		{
			name:     "Int",
			fn:       func(v interface{}) Field { return Int("k", v.(int)) },
			original: 42,
			expected: int64(42),
		},
		{
			name:     "Int64",
			fn:       func(v interface{}) Field { return Int64("k", v.(int64)) },
			original: int64(42),
			expected: int64(42),
		},
		{
			name:     "Int32",
			fn:       func(v interface{}) Field { return Int32("k", v.(int32)) },
			original: int32(42),
			expected: int64(42),
		},
		{
			name:     "Int16",
			fn:       func(v interface{}) Field { return Int16("k", v.(int16)) },
			original: int16(42),
			expected: int64(42),
		},
		{
			name:     "Int8",
			fn:       func(v interface{}) Field { return Int8("k", v.(int8)) },
			original: int8(42),
			expected: int64(42),
		},
		{
			name:     "Uint",
			fn:       func(v interface{}) Field { return Uint("k", v.(uint)) },
			original: uint(42),
			expected: uint64(42),
		},
		{
			name:     "Uint64",
			fn:       func(v interface{}) Field { return Uint64("k", v.(uint64)) },
			original: uint64(42),
			expected: uint64(42),
		},
		{
			name:     "Uint32",
			fn:       func(v interface{}) Field { return Uint32("k", v.(uint32)) },
			original: uint32(42),
			expected: uint64(42),
		},
		{
			name:     "Uint16",
			fn:       func(v interface{}) Field { return Uint16("k", v.(uint16)) },
			original: uint16(42),
			expected: uint64(42),
		},
		{
			name:     "Uint8",
			fn:       func(v interface{}) Field { return Uint8("k", v.(uint8)) },
			original: uint8(42),
			expected: uint64(42),
		},
		{
			name:     "Float64",
			fn:       func(v interface{}) Field { return Float64("k", v.(float64)) },
			original: float64(42),
			expected: float64(42),
		},
		{
			name:     "Float32",
			fn:       func(v interface{}) Field { return Float32("k", v.(float32)) },
			original: float32(42),
			expected: float64(42),
		},
		{
			name:     "Duration",
			fn:       func(v interface{}) Field { return Duration("k", v.(time.Duration)) },
			original: time.Second,
			expected: time.Second,
		},
		{
			name:     "String",
			fn:       func(v interface{}) Field { return String("k", v.(string)) },
			original: "42",
			expected: "42",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := c.fn(c.original)
			e := newTestFieldEncoder()
			f.Accept(e)
			assert.Equal(t, c.expected, e.result["k"])
		})
		t.Run(c.name+"->Any", func(t *testing.T) {
			f := Any("k", c.original)
			e := newTestFieldEncoder()
			f.Accept(e)

			assert.Equal(t, c.expected, e.result["k"])
		})
	}
}

func TestFieldStrings(t *testing.T) {
	t.Run("Strings", func(t *testing.T) {
		f := Strings("k", []string{"42"})
		e := newTestFieldEncoder()
		f.Accept(e)

		assert.Equal(t, []string{"42"}, e.result["k"])
	})
	t.Run("Any", func(t *testing.T) {
		f := Any("k", []string{"42"})
		e := newTestFieldEncoder()
		f.Accept(e)

		assert.Equal(t, []string{"42"}, e.result["k"])
	})
}

func TestFieldArray(t *testing.T) {
	golden := &testArrayEncoder{}

	check := func(t *testing.T, f *Field) {
		e := newTestFieldEncoder()
		f.Accept(e)
		assert.Equal(t, golden, e.result["k"])
	}

	t.Run("Array", func(t *testing.T) {
		f := Array("k", &testArrayEncoder{})
		check(t, &f)
	})
	t.Run("Array->Any", func(t *testing.T) {
		f := Any("k", &testArrayEncoder{})
		check(t, &f)
	})
}

func TestFieldNilArray(t *testing.T) {
	e := newTestFieldEncoder()
	f := Array("k", nil)
	assert.Panics(t, func() { f.Accept(e) })
}

func TestFieldObject(t *testing.T) {
	golden := &testObjectEncoder{}

	check := func(t *testing.T, f *Field) {
		e := newTestFieldEncoder()
		f.Accept(e)
		assert.Equal(t, golden, e.result["k"])
	}

	t.Run("Array", func(t *testing.T) {
		f := Object("k", &testObjectEncoder{})
		check(t, &f)
	})
	t.Run("Array->Any", func(t *testing.T) {
		f := Any("k", &testObjectEncoder{})
		check(t, &f)
	})
}

func TestFieldInline(t *testing.T) {
	e := newTestFieldEncoder()
	f := Inline(&testObjectEncoder{})
	f.Accept(e)
	assert.Equal(t, "username", e.result["username"])
	assert.Equal(t, int64(42), e.result["code"])
}

func TestFieldNilInline(t *testing.T) {
	e := newTestFieldEncoder()
	f := Inline(nil)
	assert.Panics(t, func() { f.Accept(e) })
}

func TestFieldNilObject(t *testing.T) {
	e := newTestFieldEncoder()
	f := Object("k", nil)
	assert.Panics(t, func() { f.Accept(e) })
}

func TestFieldTime(t *testing.T) {
	golden := time.Now()

	check := func(t *testing.T, f *Field) {
		e := newTestFieldEncoder()
		f.Accept(e)

		assert.Equal(t, golden.Format(time.RFC3339Nano), (e.result["k"].(time.Time)).Format(time.RFC3339Nano))
	}

	t.Run("Time", func(t *testing.T) {
		f := Time("k", golden)
		check(t, &f)
	})
	t.Run("Any->Time", func(t *testing.T) {
		f := Any("k", golden)
		check(t, &f)
	})
}

func TestFieldTimeWithoutLocation(t *testing.T) {
	golden := time.Unix(320836234, 0)

	e := newTestFieldEncoder()
	f := Field{Key: "k", Type: FieldTypeTime, Val: golden.UnixNano()}
	f.Accept(e)

	assert.Equal(t, golden.Format(time.RFC3339Nano), (e.result["k"].(time.Time)).Format(time.RFC3339Nano))
}

func TestFieldStringer(t *testing.T) {
	golden := "before"
	str := &testStringer{golden}

	e := newTestFieldEncoder()
	f := Stringer("k", str)

	// Change result returning by str. Stinger must call String() during a
	// call of Stringer().
	str.result = "after"

	f.Accept(e)
	assert.Equal(t, golden, e.result["k"])
}

func TestFieldNilStringer(t *testing.T) {
	e := newTestFieldEncoder()
	f := Stringer("k", nil)

	f.Accept(e)
	assert.Equal(t, "nil", e.result["k"])
}

func TestFieldFormatterV(t *testing.T) {
	type testFormatterV struct {
		str string
	}
	testing := testFormatterV{"42"}

	e := newTestFieldEncoder()
	f := FormatterV("k", &testing)

	// Change testing value to check that Formatter formats string
	// during its call.
	testing.str = "66"

	f.Accept(e)
	assert.Equal(t, `&logf.testFormatterV{str:"42"}`, e.result["k"])
}

func TestFieldByteString(t *testing.T) {
	e := newTestFieldEncoder()
	f := ByteString("k", []byte("hello"))
	f.Accept(e)
	assert.Equal(t, "hello", e.result["k"])
}

// secretToken implements fmt.Stringer to mask its value in logs.
type secretToken string

func (t secretToken) String() string {
	if len(t) > 4 {
		return string(t[:2]) + "***"
	}
	return "***"
}

func TestFieldMasking(t *testing.T) {
	raw := secretToken("eyJhbGciOiJSUzI1NiJ9.secret")

	t.Run("Stringer", func(t *testing.T) {
		e := newTestFieldEncoder()
		f := Stringer("token", raw)
		f.Accept(e)
		assert.Equal(t, "ey***", e.result["token"])
		assert.NotContains(t, e.result["token"], "secret")
	})
	t.Run("Any", func(t *testing.T) {
		e := newTestFieldEncoder()
		f := Any("token", raw)
		f.Accept(e)
		assert.Equal(t, "ey***", e.result["token"])
		assert.NotContains(t, e.result["token"], "secret")
	})
}

func TestFieldNamedError(t *testing.T) {
	golden := errors.New("named error")

	check := func(t *testing.T, f *Field) {
		e := newTestFieldEncoder()
		f.Accept(e)
		assert.Equal(t, golden, e.result["k"])
	}

	t.Run("NamedError", func(t *testing.T) {
		f := NamedError("k", golden)
		check(t, &f)
	})
	t.Run("NamedError->Any", func(t *testing.T) {
		f := Any("k", golden)
		check(t, &f)
	})
}

func TestFieldError(t *testing.T) {
	golden := errors.New("common error")

	e := newTestFieldEncoder()
	f := Error(golden)

	f.Accept(e)
	assert.Equal(t, golden, e.result["error"])
}

func TestFieldNilError(t *testing.T) {
	f := Error(nil)
	e := newTestFieldEncoder()
	f.Accept(e)
	assert.Equal(t, nil, e.result["error"])
}

func TestFieldAnyStringer(t *testing.T) {
	f := Any("k", net.IPv4(192, 168, 1, 1))
	e := newTestFieldEncoder()
	f.Accept(e)
	assert.Equal(t, "192.168.1.1", e.result["k"])
}

func TestFieldTimeZeroAccept(t *testing.T) {
	// Zero time (Val==0, Any==nil) should encode as time.Time{}.
	f := Time("k", time.Time{})
	e := newTestFieldEncoder()
	f.Accept(e)
	assert.Equal(t, time.Time{}, e.result["k"])
}

func TestFieldAnyWithCustomType(t *testing.T) {
	type customType struct{}
	customTypeValue := customType{}

	f := Any("k", &customTypeValue)
	e := newTestFieldEncoder()
	f.Accept(e)
	assert.Equal(t, &customTypeValue, e.result["k"])
}

func TestFieldAnyReflect(t *testing.T) {
	type customStringType string
	type customBoolType bool
	type customIntType int
	type customInt64Type int64
	type customInt32Type int32
	type customInt16Type int16
	type customInt8Type int8
	type customUintType uint
	type customUint64Type uint64
	type customUint32Type uint32
	type customUint16Type uint16
	type customUint8Type uint8
	type customFloat64Type float64
	type customFloat32Type float32

	cases := []struct {
		name     string
		original interface{}
		expected interface{}
	}{
		{
			name:     "String",
			original: customStringType("42"),
			expected: "42",
		},
		{
			name:     "Bool",
			original: customBoolType(true),
			expected: true,
		},
		{
			name:     "Int",
			original: customIntType(42),
			expected: int64(42),
		},
		{
			name:     "Int64",
			original: customInt64Type(42),
			expected: int64(42),
		},
		{
			name:     "Int32",
			original: customInt32Type(42),
			expected: int64(42),
		},
		{
			name:     "Int16",
			original: customInt16Type(42),
			expected: int64(42),
		},
		{
			name:     "Int8",
			original: customInt8Type(42),
			expected: int64(42),
		},
		{
			name:     "Uint",
			original: customUintType(42),
			expected: uint64(42),
		},
		{
			name:     "Uint64",
			original: customUint64Type(42),
			expected: uint64(42),
		},
		{
			name:     "Uint32",
			original: customUint32Type(42),
			expected: uint64(42),
		},
		{
			name:     "Uint16",
			original: customUint16Type(42),
			expected: uint64(42),
		},
		{
			name:     "Uint8",
			original: customUint8Type(42),
			expected: uint64(42),
		},
		{
			name:     "Float64",
			original: customFloat64Type(42),
			expected: float64(42),
		},
		{
			name:     "Float32",
			original: customFloat32Type(42),
			expected: float64(42),
		},
		{
			name:     "nil",
			original: nil,
			expected: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := Any("k", c.original)
			e := newTestFieldEncoder()
			f.Accept(e)
			assert.Equal(t, c.expected, e.result["k"])
		})
	}

	// Pointer types: non-nil → typed value, nil → FieldTypeAny(nil).
	t.Run("*string", func(t *testing.T) {
		s := "hello"
		f := Any("k", &s)
		e := newTestFieldEncoder()
		f.Accept(e)
		assert.Equal(t, "hello", e.result["k"])
	})
	t.Run("*string/nil", func(t *testing.T) {
		f := Any("k", (*string)(nil))
		assert.Equal(t, FieldTypeAny, f.Type)
	})
	t.Run("*int64", func(t *testing.T) {
		v := int64(42)
		f := Any("k", &v)
		e := newTestFieldEncoder()
		f.Accept(e)
		assert.Equal(t, int64(42), e.result["k"])
	})
	t.Run("*int64/nil", func(t *testing.T) {
		f := Any("k", (*int64)(nil))
		assert.Equal(t, FieldTypeAny, f.Type)
	})
	t.Run("*bool", func(t *testing.T) {
		v := true
		f := Any("k", &v)
		e := newTestFieldEncoder()
		f.Accept(e)
		assert.Equal(t, true, e.result["k"])
	})
	t.Run("*float64", func(t *testing.T) {
		v := float64(3.14)
		f := Any("k", &v)
		e := newTestFieldEncoder()
		f.Accept(e)
		assert.Equal(t, float64(3.14), e.result["k"])
	})
	t.Run("*time.Duration", func(t *testing.T) {
		v := time.Minute
		f := Any("k", &v)
		e := newTestFieldEncoder()
		f.Accept(e)
		assert.Equal(t, time.Minute, e.result["k"])
	})
}

func TestFieldAccept(t *testing.T) {
	type customStruct struct {
		v int
	}

	cases := []struct {
		name     string
		original Field
		expected interface{}
	}{
		{
			name:     "String",
			original: String("k", "42"),
			expected: "42",
		},
		{
			name:     "Bool",
			original: Bool("k", true),
			expected: true,
		},
		{
			name:     "Int",
			original: Int("k", 42),
			expected: int64(42),
		},
		{
			name:     "Int64",
			original: Int64("k", 42),
			expected: int64(42),
		},
		{
			name:     "Int32",
			original: Int32("k", 42),
			expected: int64(42),
		},
		{
			name:     "Int16",
			original: Int16("k", 42),
			expected: int64(42),
		},
		{
			name:     "Int8",
			original: Int8("k", 42),
			expected: int64(42),
		},
		{
			name:     "Uint",
			original: Uint("k", 42),
			expected: uint64(42),
		},
		{
			name:     "Uint64",
			original: Uint64("k", 42),
			expected: uint64(42),
		},
		{
			name:     "Uint32",
			original: Uint32("k", 42),
			expected: uint64(42),
		},
		{
			name:     "Uint16",
			original: Uint16("k", 42),
			expected: uint64(42),
		},
		{
			name:     "Uint8",
			original: Uint8("k", 42),
			expected: uint64(42),
		},
		{
			name:     "Float64",
			original: Float64("k", 42),
			expected: float64(42),
		},
		{
			name:     "Float32",
			original: Float32("k", 42),
			expected: float64(42),
		},
		{
			name:     "Strings",
			original: Strings("k", []string{"42", "43"}),
			expected: []string{"42", "43"},
		},
		{
			name:     "Ints",
			original: Ints("k", []int{42, 43}),
			expected: []int64{42, 43},
		},
		{
			name:     "Ints64",
			original: Ints64("k", []int64{42, 43}),
			expected: []int64{42, 43},
		},
		{
			name:     "Floats64",
			original: Floats64("k", []float64{42}),
			expected: []float64{42},
		},
		{
			name:     "nil",
			original: Any("k", nil),
			expected: nil,
		},
		{
			name:     "Bytes",
			original: Bytes("k", []byte{42, 43}),
			expected: []byte{42, 43},
		},
		{
			name:     "Error",
			original: NamedError("k", errors.New("test")),
			expected: errors.New("test"),
		},
		{
			name:     "Time",
			original: Time("k", time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC)),
			expected: time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
		},
		{
			name:     "Duration",
			original: Duration("k", time.Minute),
			expected: time.Minute,
		},
		{
			name:     "Stringer",
			original: Stringer("k", &testStringer{"stringer"}),
			expected: "stringer",
		},
		{
			name:     "Stringer/Nil",
			original: Stringer("k", nil),
			expected: "nil", // FIXME
		},
		{
			name:     "Array",
			original: Array("k", &testArrayEncoder{}),
			expected: &testArrayEncoder{},
		},
		{
			name:     "Object",
			original: Object("k", &testObjectEncoder{}),
			expected: &testObjectEncoder{},
		},
		{
			name:     "Any",
			original: Any("k", customStruct{42}),
			expected: customStruct{42},
		},
		{
			name:     "Group",
			original: Group("k", String("a", "hello"), Int("b", 42)),
			expected: []Field{String("a", "hello"), Int("b", 42)},
		},
		{
			name:     "Group/empty",
			original: Group("k"),
			expected: []Field(nil),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := newTestFieldEncoder()
			c.original.Accept(e)
			assert.Equal(t, c.expected, e.result["k"])
		})
	}

}

func TestFieldIf(t *testing.T) {
	// If(true) returns the field unchanged.
	f := String("k", "v").If(true)
	assert.Equal(t, FieldTypeBytesToString, f.Type)
	assert.Equal(t, "k", f.Key)

	// If(false) returns an empty field.
	f = String("k", "v").If(false)
	assert.Equal(t, FieldTypeUnknown, f.Type)
	assert.Equal(t, "", f.Key)

	// Empty field is silently skipped by Accept.
	e := newTestFieldEncoder()
	f.Accept(e)
	assert.Empty(t, e.result)
}

func TestFieldOptional(t *testing.T) {
	// Non-zero values pass through.
	assert.Equal(t, FieldTypeBytesToString, String("k", "v").Optional().Type)
	assert.Equal(t, FieldTypeInt64, Int("k", 42).Optional().Type)
	assert.Equal(t, FieldTypeError, NamedError("k", errors.New("x")).Optional().Type)
	assert.Equal(t, FieldTypeDuration, Duration("k", time.Second).Optional().Type)
	assert.Equal(t, FieldTypeBool, Bool("k", true).Optional().Type)
	assert.Equal(t, FieldTypeFloat64, Float64("k", 3.14).Optional().Type)
	assert.Equal(t, FieldTypeTime, Time("k", time.Now()).Optional().Type)
	assert.Equal(t, FieldTypeBytes, Bytes("k", []byte{1}).Optional().Type)
	assert.Equal(t, FieldTypeBytesToStrings, Strings("k", []string{"a"}).Optional().Type)
	assert.Equal(t, FieldTypeBytesToInts64, Ints64("k", []int64{1}).Optional().Type)

	// Non-zero: remaining types.
	assert.Equal(t, FieldTypeUint64, Uint64("k", 1).Optional().Type)
	assert.Equal(t, FieldTypeFloat64, Float32("k", 1.5).Optional().Type)
	assert.Equal(t, FieldTypeBytesToFloats64, Floats64("k", []float64{1}).Optional().Type)
	assert.Equal(t, FieldTypeBytesToDurations, Durations("k", []time.Duration{time.Second}).Optional().Type)
	assert.Equal(t, FieldTypeBytesToString, Any("k", "val").Optional().Type)                            // Any(string) → String
	assert.Equal(t, FieldTypeBytesToString, Stringer("k", bytes.NewBufferString("x")).Optional().Type) // Stringer → String
	assert.Equal(t, FieldTypeAny, Any("k", struct{}{}).Optional().Type)                                // Any(struct) stays FieldTypeAny
	assert.Equal(t, FieldTypeObject, Object("k", &testObjectEncoder{}).Optional().Type)
	assert.Equal(t, FieldTypeGroup, Group("k", String("a", "b")).Optional().Type)

	// Zero values are skipped.
	assert.Equal(t, FieldTypeUnknown, String("k", "").Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Int("k", 0).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Uint64("k", 0).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Float32("k", 0).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, NamedError("k", nil).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Duration("k", 0).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Bool("k", false).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Float64("k", 0).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Time("k", time.Time{}).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Bytes("k", nil).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Bytes("k", []byte{}).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Strings("k", nil).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Strings("k", []string{}).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Ints64("k", nil).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Ints64("k", []int64{}).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Floats64("k", nil).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Floats64("k", []float64{}).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Durations("k", nil).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Durations("k", []time.Duration{}).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Any("k", nil).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Any("k", (*int)(nil)).Optional().Type)       // typed nil pointer
	assert.Equal(t, FieldTypeUnknown, Any("k", (error)(nil)).Optional().Type)      // typed nil interface
	assert.Equal(t, FieldTypeUnknown, Any("k", ([]string)(nil)).Optional().Type)   // nil slice via Any
	assert.Equal(t, FieldTypeUnknown, Any("k", (map[string]int)(nil)).Optional().Type) // nil map
	assert.Equal(t, FieldTypeUnknown, Group("k").Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Group("k", []Field{}...).Optional().Type)
	assert.Equal(t, FieldTypeUnknown, Field{Type: FieldTypeGroup, Key: "k"}.Optional().Type) // nil Any in group

	// Empty field is skipped by encoder.
	e := newTestFieldEncoder()
	String("k", "").Optional().Accept(e)
	assert.Empty(t, e.result)
}

func TestFieldIfWithEncoder(t *testing.T) {
	// Verify that If(false) fields are skipped in a real encoding scenario.
	e := newTestFieldEncoder()
	fields := []Field{
		String("always", "yes"),
		String("maybe", "no").If(false),
		Int("count", 42).If(true),
	}
	for _, f := range fields {
		f.Accept(e)
	}
	assert.Equal(t, "yes", e.result["always"])
	assert.Equal(t, int64(42), e.result["count"])
	assert.NotContains(t, e.result, "maybe")
}

func TestFieldIfAndOptionalJSON(t *testing.T) {
	// End-to-end: verify skipped fields produce no JSON artifacts.
	var buf bytes.Buffer
	logger := NewLogger().Output(&buf).Build().WithCaller(false)

	logger.Info(context.Background(), "test",
		String("a", "1"),
		String("b", "").Optional(),
		Int("c", 0).Optional(),
		String("d", "skip").If(false),
		Error(nil).Optional(),
		Int("e", 5),
	)

	got := buf.String()
	assert.Contains(t, got, `"a":"1"`)
	assert.Contains(t, got, `"e":5`)
	assert.NotContains(t, got, `"b"`)
	assert.NotContains(t, got, `"c"`)
	assert.NotContains(t, got, `"d"`)
	assert.NotContains(t, got, `"error"`)
	// No double commas or trailing commas.
	assert.NotContains(t, got, ",,")
	assert.NotContains(t, got, ",}")

	// Text encoder: same check.
	buf.Reset()
	loggerText := NewLogger().Output(&buf).
		EncoderFrom(Text().NoColor().DisableTime().DisableCaller()).
		Build()

	loggerText.Info(context.Background(), "test",
		String("a", "1"),
		String("b", "").Optional(),
		Int("c", 0).Optional(),
		String("d", "skip").If(false),
		Error(nil).Optional(),
		Int("e", 5),
	)

	gotText := buf.String()
	assert.Contains(t, gotText, "a=1")
	assert.Contains(t, gotText, "e=5")
	assert.NotContains(t, gotText, "b=")
	assert.NotContains(t, gotText, "c=")
	assert.NotContains(t, gotText, "d=")
	assert.NotContains(t, gotText, "error=")
}
