package logf

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestField(t *testing.T) {
	cases := []struct {
		name   string
		fn     func(interface{}) Field
		golden interface{}
	}{
		{
			name:   "Bool",
			fn:     func(v interface{}) Field { return Bool("k", v.(bool)) },
			golden: true,
		},
		{
			name:   "Int",
			fn:     func(v interface{}) Field { return Int("k", int(v.(int64))) },
			golden: int64(42),
		},
		{
			name:   "Int64",
			fn:     func(v interface{}) Field { return Int64("k", v.(int64)) },
			golden: int64(42),
		},
		{
			name:   "Int32",
			fn:     func(v interface{}) Field { return Int32("k", v.(int32)) },
			golden: int32(42),
		},
		{
			name:   "Int16",
			fn:     func(v interface{}) Field { return Int16("k", v.(int16)) },
			golden: int16(42),
		},
		{
			name:   "Int8",
			fn:     func(v interface{}) Field { return Int8("k", v.(int8)) },
			golden: int8(42),
		},
		{
			name:   "Uint",
			fn:     func(v interface{}) Field { return Uint("k", uint(v.(uint64))) },
			golden: uint64(42),
		},
		{
			name:   "Uint64",
			fn:     func(v interface{}) Field { return Uint64("k", v.(uint64)) },
			golden: uint64(42),
		},
		{
			name:   "Uint32",
			fn:     func(v interface{}) Field { return Uint32("k", v.(uint32)) },
			golden: uint32(42),
		},
		{
			name:   "Uint16",
			fn:     func(v interface{}) Field { return Uint16("k", v.(uint16)) },
			golden: uint16(42),
		},
		{
			name:   "Uint8",
			fn:     func(v interface{}) Field { return Uint8("k", v.(uint8)) },
			golden: uint8(42),
		},
		{
			name:   "Float64",
			fn:     func(v interface{}) Field { return Float64("k", v.(float64)) },
			golden: float64(42),
		},
		{
			name:   "Float32",
			fn:     func(v interface{}) Field { return Float32("k", v.(float32)) },
			golden: float32(42),
		},
		{
			name:   "Duration",
			fn:     func(v interface{}) Field { return Duration("k", v.(time.Duration)) },
			golden: time.Second,
		},
		{
			name:   "Duration",
			fn:     func(v interface{}) Field { return Duration("k", v.(time.Duration)) },
			golden: time.Second,
		},
		{
			name:   "String",
			fn:     func(v interface{}) Field { return String("k", v.(string)) },
			golden: "42",
		},
		{
			name:   "ConstBytes",
			fn:     func(v interface{}) Field { return ConstBytes("k", v.([]byte)) },
			golden: []byte{42},
		},
		{
			name:   "ConstBools",
			fn:     func(v interface{}) Field { return ConstBools("k", v.([]bool)) },
			golden: []bool{true},
		},
		{
			name:   "ConstInts",
			fn:     func(v interface{}) Field { return ConstInts("k", []int{int((v.([]int64))[0])}) },
			golden: []int64{42},
		},
		{
			name:   "ConstInts64",
			fn:     func(v interface{}) Field { return ConstInts64("k", v.([]int64)) },
			golden: []int64{42},
		},
		{
			name:   "ConstInts32",
			fn:     func(v interface{}) Field { return ConstInts32("k", v.([]int32)) },
			golden: []int32{42},
		},
		{
			name:   "ConstInts16",
			fn:     func(v interface{}) Field { return ConstInts16("k", v.([]int16)) },
			golden: []int16{42},
		},
		{
			name:   "ConstInts8",
			fn:     func(v interface{}) Field { return ConstInts8("k", v.([]int8)) },
			golden: []int8{42},
		},
		{
			name:   "ConstUints",
			fn:     func(v interface{}) Field { return ConstUints("k", []uint{uint((v.([]uint64))[0])}) },
			golden: []uint64{42},
		},
		{
			name:   "ConstUints64",
			fn:     func(v interface{}) Field { return ConstUints64("k", v.([]uint64)) },
			golden: []uint64{42},
		},
		{
			name:   "ConstUints32",
			fn:     func(v interface{}) Field { return ConstUints32("k", v.([]uint32)) },
			golden: []uint32{42},
		},
		{
			name:   "ConstUints16",
			fn:     func(v interface{}) Field { return ConstUints16("k", v.([]uint16)) },
			golden: []uint16{42},
		},
		{
			name:   "ConstUints8",
			fn:     func(v interface{}) Field { return ConstUints8("k", v.([]uint8)) },
			golden: []uint8{42},
		},
		{
			name:   "ConstFloats64",
			fn:     func(v interface{}) Field { return ConstFloats64("k", v.([]float64)) },
			golden: []float64{42},
		},
		{
			name:   "ConstFloats32",
			fn:     func(v interface{}) Field { return ConstFloats32("k", v.([]float32)) },
			golden: []float32{42},
		},
		{
			name:   "ConstDurations",
			fn:     func(v interface{}) Field { return ConstDurations("k", v.([]time.Duration)) },
			golden: []time.Duration{time.Second},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := newTestFieldEncoder()
			f := c.fn(c.golden)
			f.Accept(e)

			assert.Equal(t, c.golden, e.result["k"])
		})
	}
}

func TestFieldStrings(t *testing.T) {
	e := newTestFieldEncoder()
	f := Strings("k", []string{"42"})
	f.Accept(e)

	ae := e.result["k"].(ArrayEncoder)
	te := testTypeEncoder{}
	ae.EncodeLogfArray(&te)

	assert.Equal(t, "42", te.result)
}

type testArrayEncoder struct{}

func (o testArrayEncoder) EncodeLogfArray(TypeEncoder) error {
	return nil
}

func TestFieldArray(t *testing.T) {
	golden := &testArrayEncoder{}

	e := newTestFieldEncoder()
	f := Array("k", &testArrayEncoder{})
	f.Accept(e)

	assert.Equal(t, golden, e.result["k"])
}

func TestFieldObject(t *testing.T) {
	golden := &testObjectEncoder{}

	e := newTestFieldEncoder()
	f := Object("k", &testObjectEncoder{})
	f.Accept(e)

	assert.Equal(t, golden, e.result["k"])
}

func TestFieldTime(t *testing.T) {
	golden := time.Now()

	e := newTestFieldEncoder()
	f := Time("k", golden)
	f.Accept(e)

	assert.Equal(t, golden.Format(time.RFC3339Nano), (e.result["k"].(time.Time)).Format(time.RFC3339Nano))
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

func TestFieldConstStringer(t *testing.T) {
	golden := "before"
	str := &testStringer{golden}

	e := newTestFieldEncoder()
	f := ConstStringer("k", str)

	f.Accept(e)
	assert.Equal(t, golden, e.result["k"])
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

	e := newTestFieldEncoder()
	f := NamedError("k", golden)

	f.Accept(e)
	assert.Equal(t, golden, e.result["k"])
}

func TestFieldError(t *testing.T) {
	golden := errors.New("common error")

	e := newTestFieldEncoder()
	f := Error(golden)

	f.Accept(e)
	assert.Equal(t, golden, e.result["error"])
}
