package logf

import (
	"errors"
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
			expected: int32(42),
		},
		{
			name:     "Int16",
			fn:       func(v interface{}) Field { return Int16("k", v.(int16)) },
			original: int16(42),
			expected: int16(42),
		},
		{
			name:     "Int8",
			fn:       func(v interface{}) Field { return Int8("k", v.(int8)) },
			original: int8(42),
			expected: int8(42),
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
			expected: uint32(42),
		},
		{
			name:     "Uint16",
			fn:       func(v interface{}) Field { return Uint16("k", v.(uint16)) },
			original: uint16(42),
			expected: uint16(42),
		},
		{
			name:     "Uint8",
			fn:       func(v interface{}) Field { return Uint8("k", v.(uint8)) },
			original: uint8(42),
			expected: uint8(42),
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
			expected: float32(42),
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
		{
			name:     "ConstBytes",
			fn:       func(v interface{}) Field { return ConstBytes("k", v.([]byte)) },
			original: []byte{42},
			expected: []byte{42},
		},
		{
			name:     "ConstBools",
			fn:       func(v interface{}) Field { return ConstBools("k", v.([]bool)) },
			original: []bool{true},
			expected: []bool{true},
		},
		{
			name:     "ConstInts",
			fn:       func(v interface{}) Field { return ConstInts("k", v.([]int)) },
			original: []int{42},
			expected: []int64{42},
		},
		{
			name:     "ConstInts64",
			fn:       func(v interface{}) Field { return ConstInts64("k", v.([]int64)) },
			original: []int64{42},
			expected: []int64{42},
		},
		{
			name:     "ConstInts32",
			fn:       func(v interface{}) Field { return ConstInts32("k", v.([]int32)) },
			original: []int32{42},
			expected: []int32{42},
		},
		{
			name:     "ConstInts16",
			fn:       func(v interface{}) Field { return ConstInts16("k", v.([]int16)) },
			original: []int16{42},
			expected: []int16{42},
		},
		{
			name:     "ConstInts8",
			fn:       func(v interface{}) Field { return ConstInts8("k", v.([]int8)) },
			original: []int8{42},
			expected: []int8{42},
		},
		{
			name:     "ConstUints",
			fn:       func(v interface{}) Field { return ConstUints("k", v.([]uint)) },
			original: []uint{42},
			expected: []uint64{42},
		},
		{
			name:     "ConstUints64",
			fn:       func(v interface{}) Field { return ConstUints64("k", v.([]uint64)) },
			original: []uint64{42},
			expected: []uint64{42},
		},
		{
			name:     "ConstUints32",
			fn:       func(v interface{}) Field { return ConstUints32("k", v.([]uint32)) },
			original: []uint32{42},
			expected: []uint32{42},
		},
		{
			name:     "ConstUints16",
			fn:       func(v interface{}) Field { return ConstUints16("k", v.([]uint16)) },
			original: []uint16{42},
			expected: []uint16{42},
		},
		{
			name:     "ConstUints8",
			fn:       func(v interface{}) Field { return ConstUints8("k", v.([]uint8)) },
			original: []uint8{42},
			expected: []uint8{42},
		},
		{
			name:     "ConstFloats64",
			fn:       func(v interface{}) Field { return ConstFloats64("k", v.([]float64)) },
			original: []float64{42},
			expected: []float64{42},
		},
		{
			name:     "ConstFloats32",
			fn:       func(v interface{}) Field { return ConstFloats32("k", v.([]float32)) },
			original: []float32{42},
			expected: []float32{42},
		},
		{
			name:     "ConstDurations",
			fn:       func(v interface{}) Field { return ConstDurations("k", v.([]time.Duration)) },
			original: []time.Duration{time.Second},
			expected: []time.Duration{time.Second},
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

			// Need to snapshot fields because Any could return raw byte
			// types that need to be copied.
			snapshotField(&f)
			e := newTestFieldEncoder()
			f.Accept(e)

			assert.Equal(t, c.expected, e.result["k"])
		})
	}
}

func TestFieldStrings(t *testing.T) {
	check := func(t *testing.T, f *Field) {
		e := newTestFieldEncoder()
		f.Accept(e)

		ae := e.result["k"].(ArrayEncoder)
		te := testTypeEncoder{}
		ae.EncodeLogfArray(&te)

		assert.Equal(t, "42", te.result)
	}

	t.Run("Time", func(t *testing.T) {
		f := Strings("k", []string{"42"})
		check(t, &f)
	})
	t.Run("Time->Any", func(t *testing.T) {
		f := Any("k", []string{"42"})
		check(t, &f)
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
	f.Accept(e)
	assert.Equal(t, "nil", e.result["k"])
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

func TestFieldNilObject(t *testing.T) {
	e := newTestFieldEncoder()
	f := Object("k", nil)
	f.Accept(e)
	assert.Equal(t, "nil", e.result["k"])
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
	f := Field{Key: "k", Type: FieldTypeTime, Int: golden.UnixNano()}
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

func TestFieldConstStringer(t *testing.T) {
	golden := "before"
	str := &testStringer{golden}

	e := newTestFieldEncoder()
	f := ConstStringer("k", str)

	f.Accept(e)
	assert.Equal(t, golden, e.result["k"])
}

func TestFieldNilConstStringer(t *testing.T) {
	e := newTestFieldEncoder()
	f := ConstStringer("k", nil)

	f.Accept(e)
	assert.Equal(t, "nil", e.result["k"])
}
func TestFieldConstFormatter(t *testing.T) {
	golden := "42"
	e := newTestFieldEncoder()
	f := ConstFormatter("k", "%d", 42)

	f.Accept(e)
	assert.Equal(t, golden, e.result["k"])
}

func TestFieldConstFormatterV(t *testing.T) {
	type testFormatterV struct {
		str string
	}

	e := newTestFieldEncoder()
	f := ConstFormatterV("k", testFormatterV{"42"})

	f.Accept(e)
	assert.Equal(t, `logf.testFormatterV{str:"42"}`, e.result["k"])
}

func TestFieldFormatter(t *testing.T) {
	type testFormatterV struct {
		str string
	}
	testing := testFormatterV{"42"}

	e := newTestFieldEncoder()
	f := Formatter("k", "%s", &testing)

	// Change testing value to check that ConstFormatter formats string
	// during it's call.
	testing.str = "66"

	f.Accept(e)
	assert.Equal(t, "&{42}", e.result["k"])
}

func TestFieldFormatterV(t *testing.T) {
	type testFormatterV struct {
		str string
	}
	testing := testFormatterV{"42"}

	e := newTestFieldEncoder()
	f := FormatterV("k", &testing)

	// Change testing value to check that ConstFormatter formats string
	// during it's call.
	testing.str = "66"

	f.Accept(e)
	assert.Equal(t, `&logf.testFormatterV{str:"42"}`, e.result["k"])
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
			expected: int32(42),
		},
		{
			name:     "Int16",
			original: Int16("k", 42),
			expected: int16(42),
		},
		{
			name:     "Int8",
			original: Int8("k", 42),
			expected: int8(42),
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
			expected: uint32(42),
		},
		{
			name:     "Uint16",
			original: Uint16("k", 42),
			expected: uint16(42),
		},
		{
			name:     "Uint8",
			original: Uint8("k", 42),
			expected: uint8(42),
		},
		{
			name:     "Float64",
			original: Float64("k", 42),
			expected: float64(42),
		},
		{
			name:     "Float32",
			original: Float32("k", 42),
			expected: float32(42),
		},
		{
			name:     "Strings",
			original: Strings("k", []string{"42", "43"}),
			expected: stringArray{"42", "43"}, // FIXME
		},
		{
			name:     "Bools",
			original: Bools("k", []bool{true, false}),
			expected: []bool{true, false},
		},
		{
			name:     "ConstBools",
			original: ConstBools("k", []bool{true, false}),
			expected: []bool{true, false},
		},
		{
			name:     "Ints",
			original: Ints("k", []int{42, 43}),
			expected: []int64{42, 43},
		},
		{
			name:     "ConstInts",
			original: ConstInts("k", []int{42, 43}),
			expected: []int64{42, 43},
		},
		{
			name:     "Ints64",
			original: Ints64("k", []int64{42, 43}),
			expected: []int64{42, 43},
		},
		{
			name:     "ConstInts64",
			original: ConstInts64("k", []int64{42, 43}),
			expected: []int64{42, 43},
		},
		{
			name:     "Ints32",
			original: Ints32("k", []int32{42, 43}),
			expected: []int32{42, 43},
		},
		{
			name:     "ConstInts32",
			original: ConstInts32("k", []int32{42, 43}),
			expected: []int32{42, 43},
		},
		{
			name:     "Ints16",
			original: Ints16("k", []int16{42, 43}),
			expected: []int16{42, 43},
		},
		{
			name:     "ConstInts16",
			original: ConstInts16("k", []int16{42, 43}),
			expected: []int16{42, 43},
		},
		{
			name:     "Ints8",
			original: Ints8("k", []int8{42}),
			expected: []int8{42},
		},
		{
			name:     "ConstInts8",
			original: ConstInts8("k", []int8{42}),
			expected: []int8{42},
		},
		{
			name:     "Uints",
			original: Uints("k", []uint{42, 43}),
			expected: []uint64{42, 43},
		},
		{
			name:     "ConstUints",
			original: ConstUints("k", []uint{42, 43}),
			expected: []uint64{42, 43},
		},
		{
			name:     "Uints64",
			original: Uints64("k", []uint64{42, 43}),
			expected: []uint64{42, 43},
		},
		{
			name:     "ConstUints64",
			original: ConstUints64("k", []uint64{42, 43}),
			expected: []uint64{42, 43},
		},
		{
			name:     "Uints32",
			original: Uints32("k", []uint32{42, 43}),
			expected: []uint32{42, 43},
		},
		{
			name:     "ConstUints32",
			original: ConstUints32("k", []uint32{42, 43}),
			expected: []uint32{42, 43},
		},
		{
			name:     "Uints16",
			original: Uints16("k", []uint16{42, 43}),
			expected: []uint16{42, 43},
		},
		{
			name:     "ConstUints16",
			original: ConstUints16("k", []uint16{42, 43}),
			expected: []uint16{42, 43},
		},
		{
			name:     "Uints8",
			original: Uints8("k", []uint8{42}),
			expected: []uint8{42},
		},
		{
			name:     "ConstUints8",
			original: ConstUints8("k", []uint8{42}),
			expected: []uint8{42},
		},
		{
			name:     "Floats64",
			original: Floats64("k", []float64{42}),
			expected: []float64{42},
		},
		{
			name:     "ConstFloats64",
			original: ConstFloats64("k", []float64{42}),
			expected: []float64{42},
		},
		{
			name:     "Floats32",
			original: Floats32("k", []float32{42}),
			expected: []float32{42},
		},
		{
			name:     "ConstFloats32",
			original: ConstFloats32("k", []float32{42}),
			expected: []float32{42},
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
			name:     "ConstBytes",
			original: ConstBytes("k", []byte{42, 43}),
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
			name:     "ConstStringer",
			original: ConstStringer("k", &testStringer{"stringer"}),
			expected: "stringer",
		},
		{
			name:     "Stringer/Nil",
			original: Stringer("k", nil),
			expected: "nil", // FIXME
		},
		{
			name:     "ConstStringer/Nil",
			original: ConstStringer("k", nil),
			expected: "nil", // FIXME
		},
		{
			name:     "Array",
			original: Array("k", &testArrayEncoder{}),
			expected: &testArrayEncoder{},
		},
		{
			name:     "Array/Nil",
			original: Array("k", nil),
			expected: "nil", // FIXME
		},
		{
			name:     "Object",
			original: Object("k", &testObjectEncoder{}),
			expected: &testObjectEncoder{},
		},
		{
			name:     "Object/Nil",
			original: Object("k", nil),
			expected: "nil", // FIXME
		},
		{
			name:     "Object/Nil",
			original: Object("k", nil),
			expected: "nil", // FIXME
		},
		{
			name:     "Any",
			original: Any("k", customStruct{42}),
			expected: customStruct{42},
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
