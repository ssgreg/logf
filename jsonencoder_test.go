package logf

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type encoderTestCase struct {
	name   string
	entry  Entry
	golden string
}


func TestEncoder(t *testing.T) {
	testCases := []encoderTestCase{
		{
			"Message",
			Entry{
				Text: "m",
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"m"}` + "\n",
		},
		{
			"LevelDebug",
			Entry{
				Level: LevelDebug,
			},
			`{"level":"debug","ts":"0001-01-01T00:00:00Z","msg":""}` + "\n",
		},
		{
			"LevelInfo",
			Entry{
				Level: LevelInfo,
			},
			`{"level":"info","ts":"0001-01-01T00:00:00Z","msg":""}` + "\n",
		},
		{
			"LevelWarn",
			Entry{
				Level: LevelWarn,
			},
			`{"level":"warn","ts":"0001-01-01T00:00:00Z","msg":""}` + "\n",
		},
		{
			"LevelError",
			Entry{
				Level: LevelError,
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":""}` + "\n",
		},
		{
			"LoggerName",
			Entry{
				LoggerName: "logger.name",
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","logger":"logger.name","msg":""}` + "\n",
		},
		{
			"CallerPC",
			Entry{
				CallerPC: CallerPC(0),
			},
			"", // checked separately below
		},
		{
			"FieldsNumbers",
			Entry{
				Fields: []Field{
					Bool("bool", true),
					Int("int", 42), Int64("int64", 42), Int32("int32", 42), Int16("int16", 42), Int8("int8", 42),
					Uint("uint", 42), Uint64("uint64", 42), Uint32("uint32", 42), Uint16("uint16", 42), Uint8("uint8", 42),
					Float64("float64", 4.2), Float32("float32", 4.2),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","bool":true,"int":42,"int64":42,"int32":42,"int16":42,"int8":42,"uint":42,"uint64":42,"uint32":42,"uint16":42,"uint8":42,"float64":4.2,"float32":4.2}` + "\n",
		},
		{
			"FieldsConstSlices",
			Entry{
				Fields: []Field{
					ConstBools("bools", []bool{true}),
					ConstInts("ints", []int{42}), ConstInts64("ints64", []int64{42}), ConstInts32("ints32", []int32{42}), ConstInts16("ints16", []int16{42}), ConstInts8("ints8", []int8{42}),
					ConstUints("uints", []uint{42}), ConstUints64("uints64", []uint64{42}), ConstUints32("uints32", []uint32{42}), ConstUints16("uints16", []uint16{42}), ConstUints8("uints8", []uint8{42}),
					ConstFloats64("floats64", []float64{4.2}), ConstFloats32("floats32", []float32{4.2}),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","bools":[true],"ints":[42],"ints64":[42],"ints32":[42],"ints16":[42],"ints8":[42],"uints":[42],"uints64":[42],"uints32":[42],"uints16":[42],"uints8":[42],"floats64":[4.2],"floats32":[4.2]}` + "\n",
		},
		{
			"FieldsSlices",
			Entry{
				Fields: []Field{
					Bools("bools", []bool{true}),
					Ints("ints", []int{42}), Ints64("ints64", []int64{42}), Ints32("ints32", []int32{42}), Ints16("ints16", []int16{42}), Ints8("ints8", []int8{42}),
					Uints("uints", []uint{42}), Uints64("uints64", []uint64{42}), Uints32("uints32", []uint32{42}), Uints16("uints16", []uint16{42}), Uints8("uints8", []uint8{42}),
					Floats64("floats64", []float64{4.2}), Floats32("floats32", []float32{4.2}),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","bools":[true],"ints":[42],"ints64":[42],"ints32":[42],"ints16":[42],"ints8":[42],"uints":[42],"uints64":[42],"uints32":[42],"uints16":[42],"uints8":[42],"floats64":[4.2],"floats32":[4.2]}` + "\n",
		},
		{
			"FieldsDuration",
			Entry{
				Fields: []Field{
					Duration("duration", time.Second),
					ConstDurations("durations", []time.Duration{time.Second}),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","duration":"1s","durations":["1s"]}` + "\n",
		},
		{
			"FieldsTime",
			Entry{
				Fields: []Field{
					Time("time", time.Unix(320836234, 0).UTC()),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","time":"1980-03-02T09:10:34Z"}` + "\n",
		},
		{
			"FieldsArray",
			Entry{
				Fields: []Field{
					Array("array", &testArrayEncoder{}),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","array":[42]}` + "\n",
		},
		{
			"FieldsObject",
			Entry{
				Fields: []Field{
					Object("object", &testObjectEncoder{}),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","object":{"username":"username","code":42}}` + "\n",
		},
		{
			"FieldsError",
			Entry{
				Fields: []Field{
					Error(&verboseError{"short", "verbose"}),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","error":"short","error.verbose":"verbose"}` + "\n",
		},
		{
			"FieldsNilError",
			Entry{
				Fields: []Field{
					NamedError("error", nil),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","error":"<nil>"}` + "\n",
		},
		{
			"FieldsBytes",
			Entry{
				Fields: []Field{
					ConstBytes("bytes", []byte{0x42}),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","bytes":"Qg=="}` + "\n",
		},
		{
			"FieldsStrings",
			Entry{
				Fields: []Field{
					Strings("strings", []string{"a", "b"}),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","strings":["a","b"]}` + "\n",
		},
		{
			"FieldsStringer",
			Entry{
				Fields: []Field{
					Stringer("stringer", time.Second),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","stringer":"1s"}` + "\n",
		},
		{
			"FieldsConstStringer",
			Entry{
				Fields: []Field{
					ConstStringer("stringer", time.Second),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","stringer":"1s"}` + "\n",
		},
		{
			"FieldsNilStringer",
			Entry{
				Fields: []Field{
					Stringer("stringer", nil),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","stringer":"nil"}` + "\n",
		},
		{
			"FieldsFormatter",
			Entry{
				Fields: []Field{
					Formatter("fmt", "%d", 42),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","fmt":"42"}` + "\n",
		},
		{
			"FieldsAny",
			Entry{
				Fields: []Field{
					Any("any", &struct{ Field string }{Field: "42"}),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","any":{"Field":"42"}}` + "\n",
		},
		{
			"FieldsGroup",
			Entry{
				Fields: []Field{
					Group("request", String("id", "abc"), Int("status", 200)),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","request":{"id":"abc","status":200}}` + "\n",
		},
		{
			"FieldsGroupEmpty",
			Entry{
				Fields: []Field{
					Group("empty"),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","empty":{}}` + "\n",
		},
		{
			"FieldsGroupNested",
			Entry{
				Fields: []Field{
					Group("outer", String("a", "1"), Group("inner", Int("b", 2))),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","outer":{"a":"1","inner":{"b":2}}}` + "\n",
		},
		{
			"FieldsLoggerBag",
			Entry{
				LoggerBag: NewBag(
					Int("int", 42),
				),
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","int":42}` + "\n",
		},
		{
			"FieldsLoggerBagFirst",
			Entry{
				LoggerBag: NewBag(
					Int("int", 42),
				),
				Fields: []Field{
					String("string", "42"),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","int":42,"string":"42"}` + "\n",
		},
	}

	enc := NewJSONEncoder.Default()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBuffer()
			_ = enc.Encode(b, tc.entry)

			if tc.golden != "" {
				require.EqualValues(t, tc.golden, b.String())
			} else {
				// CallerPC: line number is dynamic, just check key presence.
				assert.Contains(t, b.String(), `"caller":"logf/jsonencoder_test.go:`)
			}

			var a map[string]interface{}
			require.NoError(t, json.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(&a))
		})
	}
}

func TestEscapeString(t *testing.T) {
	testCases := []struct {
		golden string
		source string
	}{
		{`кириллица`, "кириллица"},
		{`not<escape>html`, `not<escape>html`},
		{`badtext\ufffd`, "badtext\xc5"},
		{`ошибка\ufffdошибка`, "ошибка\xc5ошибка"},
		{`测试`, "测试"},
		{`测\ufffd试`, "测\xc5试"},
		{`\u0008\\\r\n\t\"`, "\b\\\r\n\t\""},
	}

	for _, tc := range testCases {
		b := NewBuffer()
		assert.NoError(t, EscapeString(b, tc.source))
		assert.Equal(t, tc.golden, b.String())
	}
}

func TestEscapeByteString(t *testing.T) {
	testCases := []struct {
		golden string
		source string
	}{
		{`кириллица`, "кириллица"},
		{`not<escape>html`, `not<escape>html`},
		{`测试`, "测试"},
		{`\u0008\\\r\n\t\"`, "\b\\\r\n\t\""},
	}

	for _, tc := range testCases {
		b := NewBuffer()
		assert.NoError(t, EscapeByteString(b, []byte(tc.source)))
		assert.Equal(t, tc.golden, b.String())
	}
}

func TestEncoderFactory(t *testing.T) {
	b := NewBuffer()
	ef := NewJSONTypeEncoderFactory.Default()
	te := ef.TypeEncoder(b)

	te.EncodeTypeString("42")
	assert.Equal(t, `"42"`, b.String())
}
