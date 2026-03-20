package logf

import (
	"bytes"
	"encoding/json"
	"math"
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
			`{"level":"error","msg":"m"}` + "\n",
		},
		{
			"LevelDebug",
			Entry{
				Level: LevelDebug,
			},
			`{"level":"debug","msg":""}` + "\n",
		},
		{
			"LevelInfo",
			Entry{
				Level: LevelInfo,
			},
			`{"level":"info","msg":""}` + "\n",
		},
		{
			"LevelWarn",
			Entry{
				Level: LevelWarn,
			},
			`{"level":"warn","msg":""}` + "\n",
		},
		{
			"LevelError",
			Entry{
				Level: LevelError,
			},
			`{"level":"error","msg":""}` + "\n",
		},
		{
			"LoggerName",
			Entry{
				LoggerName: "logger.name",
			},
			`{"level":"error","logger":"logger.name","msg":""}` + "\n",
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
			`{"level":"error","msg":"","bool":true,"int":42,"int64":42,"int32":42,"int16":42,"int8":42,"uint":42,"uint64":42,"uint32":42,"uint16":42,"uint8":42,"float64":4.2,"float32":4.199999809265137}` + "\n",
		},
		{
			"FieldsSlices",
			Entry{
				Fields: []Field{
					Ints("ints", []int{42}), Ints64("ints64", []int64{42}),
					Floats64("floats64", []float64{4.2}),
				},
			},
			`{"level":"error","msg":"","ints":[42],"ints64":[42],"floats64":[4.2]}` + "\n",
		},
		{
			"FieldsDuration",
			Entry{
				Fields: []Field{
					Duration("duration", time.Second),
					Durations("durations", []time.Duration{time.Second}),
				},
			},
			`{"level":"error","msg":"","duration":"1s","durations":["1s"]}` + "\n",
		},
		{
			"FieldsTime",
			Entry{
				Fields: []Field{
					Time("time", time.Unix(320836234, 0).UTC()),
				},
			},
			`{"level":"error","msg":"","time":"1980-03-02T09:10:34Z"}` + "\n",
		},
		{
			"FieldsArray",
			Entry{
				Fields: []Field{
					Array("array", &testArrayEncoder{}),
				},
			},
			`{"level":"error","msg":"","array":[42]}` + "\n",
		},
		{
			"FieldsObject",
			Entry{
				Fields: []Field{
					Object("object", &testObjectEncoder{}),
				},
			},
			`{"level":"error","msg":"","object":{"username":"username","code":42}}` + "\n",
		},
		{
			"FieldsInline",
			Entry{
				Fields: []Field{
					Inline(&testObjectEncoder{}),
					String("extra", "value"),
				},
			},
			`{"level":"error","msg":"","username":"username","code":42,"extra":"value"}` + "\n",
		},
		{
			"FieldsError",
			Entry{
				Fields: []Field{
					Error(&verboseError{"short", "verbose"}),
				},
			},
			`{"level":"error","msg":"","error":"short","error.verbose":"verbose"}` + "\n",
		},
		{
			"FieldsNilError",
			Entry{
				Fields: []Field{
					NamedError("error", nil),
				},
			},
			`{"level":"error","msg":"","error":"<nil>"}` + "\n",
		},
		{
			"FieldsBytes",
			Entry{
				Fields: []Field{
					Bytes("bytes", []byte{0x42}),
				},
			},
			`{"level":"error","msg":"","bytes":"Qg=="}` + "\n",
		},
		{
			"FieldsStrings",
			Entry{
				Fields: []Field{
					Strings("strings", []string{"a", "b"}),
				},
			},
			`{"level":"error","msg":"","strings":["a","b"]}` + "\n",
		},
		{
			"FieldsStringer",
			Entry{
				Fields: []Field{
					Stringer("stringer", time.Second),
				},
			},
			`{"level":"error","msg":"","stringer":"1s"}` + "\n",
		},
		{
			"FieldsStringerAlt",
			Entry{
				Fields: []Field{
					Stringer("stringer", time.Second),
				},
			},
			`{"level":"error","msg":"","stringer":"1s"}` + "\n",
		},
		{
			"FieldsNilStringer",
			Entry{
				Fields: []Field{
					Stringer("stringer", nil),
				},
			},
			`{"level":"error","msg":"","stringer":"nil"}` + "\n",
		},
		{
			"FieldsFormatter",
			Entry{
				Fields: []Field{
					Formatter("fmt", "%d", 42),
				},
			},
			`{"level":"error","msg":"","fmt":"42"}` + "\n",
		},
		{
			"FieldsAny",
			Entry{
				Fields: []Field{
					Any("any", &struct{ Field string }{Field: "42"}),
				},
			},
			`{"level":"error","msg":"","any":{"Field":"42"}}` + "\n",
		},
		{
			"FieldsGroup",
			Entry{
				Fields: []Field{
					Group("request", String("id", "abc"), Int("status", 200)),
				},
			},
			`{"level":"error","msg":"","request":{"id":"abc","status":200}}` + "\n",
		},
		{
			"FieldsGroupEmpty",
			Entry{
				Fields: []Field{
					Group("empty"),
				},
			},
			`{"level":"error","msg":"","empty":{}}` + "\n",
		},
		{
			"FieldsGroupNested",
			Entry{
				Fields: []Field{
					Group("outer", String("a", "1"), Group("inner", Int("b", 2))),
				},
			},
			`{"level":"error","msg":"","outer":{"a":"1","inner":{"b":2}}}` + "\n",
		},
		{
			"FieldsLoggerBag",
			Entry{
				LoggerBag: NewBag(
					Int("int", 42),
				),
			},
			`{"level":"error","msg":"","int":42}` + "\n",
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
			`{"level":"error","msg":"","int":42,"string":"42"}` + "\n",
		},
		{
			"WithGroup",
			Entry{
				LoggerBag: NewBag(String("a", "1")).WithGroup("http").With(String("method", "GET")),
				Fields:    []Field{Int("status", 200)},
			},
			`{"level":"error","msg":"","a":"1","http":{"method":"GET","status":200}}` + "\n",
		},
		{
			"WithGroupNested",
			Entry{
				LoggerBag: NewBag().WithGroup("http").WithGroup("request").With(String("path", "/api")),
				Fields:    []Field{Int("status", 200)},
			},
			`{"level":"error","msg":"","http":{"request":{"path":"/api","status":200}}}` + "\n",
		},
		{
			"WithGroupNoFields",
			Entry{
				LoggerBag: NewBag().WithGroup("http"),
				Fields:    []Field{Int("status", 200)},
			},
			`{"level":"error","msg":"","http":{"status":200}}` + "\n",
		},
		{
			"WithGroupAndWith",
			Entry{
				LoggerBag: NewBag().WithGroup("http").With(String("method", "GET")).With(String("path", "/api")),
				Fields:    []Field{Int("status", 200)},
			},
			`{"level":"error","msg":"","http":{"method":"GET","path":"/api","status":200}}` + "\n",
		},
		{
			"WithGroupEmpty",
			Entry{
				LoggerBag: NewBag().WithGroup("http"),
			},
			`{"level":"error","msg":""}` + "\n",
		},
		{
			"WithGroupNestedEmpty",
			Entry{
				LoggerBag: NewBag().WithGroup("http").WithGroup("request"),
			},
			`{"level":"error","msg":""}` + "\n",
		},
		{
			"WithGroupPartiallyEmpty",
			Entry{
				LoggerBag: NewBag().WithGroup("http").With(String("host", "localhost")).WithGroup("request"),
			},
			`{"level":"error","msg":"","http":{"host":"localhost"}}` + "\n",
		},
	}

	enc := JSON().Build()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := enc.Encode(tc.entry)

			if tc.golden != "" {
				require.EqualValues(t, tc.golden, b.String())
			} else {
				// CallerPC: line number is dynamic, just check key presence.
				assert.Contains(t, b.String(), `"caller":"logf/jsonencoder_test.go:`)
			}

			var a map[string]interface{}
			require.NoError(t, json.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(&a))
			b.Free()
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

func TestEscapeStringBytes(t *testing.T) {
	testCases := []struct {
		golden string
		source string
	}{
		{`кириллица`, "кириллица"},
		{`not<escape>html`, `not<escape>html`},
		{`测试`, "测试"},
		{`\u0008\\\r\n\t\"`, "\b\\\r\n\t\""},
		// Invalid UTF-8 — handled by generic EscapeString[[]byte].
		{`badtext\ufffd`, "badtext\xc5"},
		{`ошибка\ufffdошибка`, "ошибка\xc5ошибка"},
		{`测\ufffd试`, "测\xc5试"},
	}

	for _, tc := range testCases {
		b := NewBuffer()
		assert.NoError(t, EscapeString(b, []byte(tc.source)))
		assert.Equal(t, tc.golden, b.String())
	}
}

func TestEncodeFloatNaNInf(t *testing.T) {
	enc := JSON().Build()

	testCases := []struct {
		name   string
		fields []Field
		want   string
	}{
		{"Float64_NaN", []Field{Float64("v", math.NaN())}, `"NaN"`},
		{"Float64_PosInf", []Field{Float64("v", math.Inf(1))}, `"+Inf"`},
		{"Float64_NegInf", []Field{Float64("v", math.Inf(-1))}, `"-Inf"`},
		{"Float32_NaN", []Field{Float32("v", float32(math.NaN()))}, `"NaN"`},
		{"Float32_PosInf", []Field{Float32("v", float32(math.Inf(1)))}, `"+Inf"`},
		{"Float32_NegInf", []Field{Float32("v", float32(math.Inf(-1)))}, `"-Inf"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := enc.Encode(Entry{Fields: tc.fields})
			require.NoError(t, err)
			assert.Contains(t, b.String(), `"v":`+tc.want)
			// Must be valid JSON.
			var m map[string]interface{}
			require.NoError(t, json.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(&m))
			b.Free()
		})
	}
}

func TestEncodeNoopEncoderFallback(t *testing.T) {
	// No-op LevelEncoder: writes nothing.
	noopLevel := func(l Level, m TypeEncoder) {}
	// No-op TimeEncoder: writes nothing.
	noopTime := func(t time.Time, m TypeEncoder) {}
	// No-op CallerEncoder: writes nothing.
	noopCaller := func(pc uintptr, m TypeEncoder) {}

	enc := JSON().
		EncodeLevel(LevelEncoder(noopLevel)).
		EncodeTime(TimeEncoder(noopTime)).
		EncodeCaller(CallerEncoder(noopCaller)).
		Build()

	e := Entry{
		Level:    LevelInfo,
		Time:     time.Unix(1234567890, 123456789),
		Text:     "hello",
		CallerPC: CallerPC(0),
	}

	b, err := enc.Encode(e)
	require.NoError(t, err)
	s := b.String()

	// Level fallback: should contain the level string.
	assert.Contains(t, s, `"level":"info"`)
	// Time fallback: should contain UnixNano int.
	assert.Contains(t, s, `"ts":1234567890123456789`)
	// Caller fallback: should contain "unknown".
	assert.Contains(t, s, `"caller":"unknown"`)
	// Must be valid JSON.
	var m map[string]interface{}
	require.NoError(t, json.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(&m))
	b.Free()
}

func TestJSONEncoderBuilderKeys(t *testing.T) {
	enc := JSON().
		NameKey("logger_name").
		CallerKey("source").
		TimeKey("timestamp").
		LevelKey("severity").
		MsgKey("message").
		Build()

	e := Entry{
		Text:       "hello",
		Level:      LevelInfo,
		Time:       time.Unix(1234567890, 0),
		LoggerName: "myapp",
		CallerPC:   CallerPC(0),
	}
	b, err := enc.Encode(e)
	require.NoError(t, err)
	s := b.String()

	assert.Contains(t, s, `"severity":`)
	assert.Contains(t, s, `"timestamp":`)
	assert.Contains(t, s, `"logger_name":"myapp"`)
	assert.Contains(t, s, `"message":"hello"`)
	assert.Contains(t, s, `"source":`)

	// Must be valid JSON.
	var m map[string]interface{}
	require.NoError(t, json.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(&m))
	b.Free()
}

func TestJSONEncoderBuilderDisable(t *testing.T) {
	enc := JSON().
		DisableLevel().
		DisableMsg().
		DisableName().
		DisableTime().
		DisableCaller().
		Build()

	e := Entry{
		Text:       "hello",
		Level:      LevelInfo,
		Time:       time.Now(),
		LoggerName: "myapp",
		CallerPC:   CallerPC(0),
		Fields:     []Field{String("k", "v")},
	}
	b, err := enc.Encode(e)
	require.NoError(t, err)
	s := b.String()

	assert.NotContains(t, s, `"level":`)
	assert.NotContains(t, s, `"msg":`)
	assert.NotContains(t, s, `"logger":`)
	assert.NotContains(t, s, `"ts":`)
	assert.NotContains(t, s, `"caller":`)
	assert.Contains(t, s, `"k":"v"`)

	var m map[string]interface{}
	require.NoError(t, json.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(&m))
	b.Free()
}

func TestJSONEncoderBuilderEncoders(t *testing.T) {
	customDuration := func(d time.Duration, te TypeEncoder) {
		te.EncodeTypeFloat64(d.Seconds())
	}
	customError := func(k string, err error, fe FieldEncoder) {
		fe.EncodeFieldString(k+"_custom", err.Error())
	}

	enc := JSON().
		EncodeDuration(customDuration).
		EncodeError(customError).
		Build()

	e := Entry{
		Fields: []Field{
			Duration("dur", 2*time.Second),
			NamedError("err", &verboseError{"bad", "detail"}),
		},
	}
	b, err := enc.Encode(e)
	require.NoError(t, err)
	s := b.String()

	assert.Contains(t, s, `"dur":2`)
	assert.Contains(t, s, `"err_custom":"bad"`)
	b.Free()
}

func TestJSONEncoderClone(t *testing.T) {
	enc := JSON().Build()
	clone := enc.Clone()
	require.NotNil(t, clone)

	// Both should produce identical output.
	e := Entry{Text: "clone-test", Level: LevelInfo}
	b1, err := enc.Encode(e)
	require.NoError(t, err)
	b2, err := clone.Encode(e)
	require.NoError(t, err)
	assert.Equal(t, b1.String(), b2.String())
	b1.Free()
	b2.Free()
}

func TestNewJSONEncoder(t *testing.T) {
	enc := NewJSONEncoder(JSONEncoderConfig{
		FieldKeyMsg:   "message",
		FieldKeyLevel: "severity",
	})
	require.NotNil(t, enc)

	e := Entry{Text: "direct", Level: LevelWarn}
	b, err := enc.Encode(e)
	require.NoError(t, err)
	s := b.String()

	assert.Contains(t, s, `"severity":"warn"`)
	assert.Contains(t, s, `"message":"direct"`)

	var m map[string]interface{}
	require.NoError(t, json.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(&m))
	b.Free()
}

