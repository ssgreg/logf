package logf

import (
	"bytes"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type encoderTestCase struct {
	Name   string
	Entry  Entry
	Golden string
}

var loggerID = int32(0)

func newLoggerID() int32 {
	atomic.AddInt32(&loggerID, 1)

	return loggerID
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
				Level: LevelInfo,
			},
			`{"level":"info","ts":"0001-01-01T00:00:00Z","msg":""}` + "\n",
		},
		{
			"LevelInfo",
			Entry{
				Level: LevelDebug,
			},
			`{"level":"debug","ts":"0001-01-01T00:00:00Z","msg":""}` + "\n",
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
			"LoggerName",
			Entry{
				Caller: EntryCaller{
					File:      "/a/b/c/f.go",
					Line:      6,
					Specified: true,
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","caller":"c/f.go:6"}` + "\n",
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
			"FieldsSlicesWithNumbers",
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
			"FieldsBytes",
			Entry{
				Fields: []Field{
					ConstBytes("bytes", []byte{0x42}),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","bytes":"Qg=="}` + "\n",
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
			"FieldsDerivedFields",
			Entry{
				DerivedFields: []Field{
					Int("int", 42),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","int":42}` + "\n",
		},
		{
			"FieldsDerivedFieldsFirst",
			Entry{
				DerivedFields: []Field{
					Int("int", 42),
				},
				Fields: []Field{
					String("string", "42"),
				},
			},
			`{"level":"error","ts":"0001-01-01T00:00:00Z","msg":"","int":42,"string":"42"}` + "\n",
		},
	}

	enc := NewJSONEncoder.Default()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Setup skipped fields (skipped for short test case description).
			tc.Entry.LoggerID = newLoggerID()

			b := NewBuffer()
			enc.Encode(b, tc.Entry)
			require.EqualValues(t, tc.Golden, b.String())

			// Check for correct json.
			var a map[string]interface{}
			require.NoError(t, json.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(&a), "generated json expected to be parsed by native golang json encoder")
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
