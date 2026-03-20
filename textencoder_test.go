package logf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextEncoder(t *testing.T) {
	testCases := []encoderTestCase{
		{
			"Message",
			Entry{Text: "m"},
			"[ERR] m\n",
		},
		{
			"LevelDebug",
			Entry{Level: LevelDebug},
			"[DBG]\n",
		},
		{
			"LevelInfo",
			Entry{Level: LevelInfo},
			"[INF]\n",
		},
		{
			"LevelWarn",
			Entry{Level: LevelWarn},
			"[WRN]\n",
		},
		{
			"LevelError",
			Entry{Level: LevelError},
			"[ERR]\n",
		},
		{
			"LoggerName",
			Entry{LoggerName: "logger.name"},
			"[ERR] logger.name:\n",
		},
		{
			"CallerPC",
			Entry{CallerPC: CallerPC(0)},
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
			"[ERR] › bool=true int=42 int64=42 int32=42 int16=42 int8=42 uint=42 uint64=42 uint32=42 uint16=42 uint8=42 float64=4.2 float32=4.199999809265137\n",
		},
		{
			"FieldsSlices",
			Entry{
				Fields: []Field{
					Ints("ints", []int{42}), Ints64("ints64", []int64{42}),
					Floats64("floats64", []float64{4.2}),
				},
			},
			"[ERR] › ints=[42] ints64=[42] floats64=[4.2]\n",
		},
		{
			"FieldsDuration",
			Entry{
				Fields: []Field{
					Duration("duration", time.Second),
					Durations("durations", []time.Duration{time.Second}),
				},
			},
			"[ERR] › duration=1s durations=[1s]\n",
		},
		{
			"FieldsTime",
			Entry{
				Fields: []Field{
					Time("time", time.Unix(320836234, 0).UTC()),
				},
			},
			"[ERR] › time=\"Mar  2 09:10:34.000\"\n",
		},
		{
			"FieldsArray",
			Entry{
				Fields: []Field{
					Array("array", &testArrayEncoder{}),
				},
			},
			"[ERR] › array=[42]\n",
		},
		{
			"FieldsObject",
			Entry{
				Fields: []Field{
					Object("object", &testObjectEncoder{}),
				},
			},
			"[ERR] › object={\"username\":\"username\",\"code\":42}\n",
		},
		{
			"FieldsInline",
			Entry{
				Fields: []Field{
					Inline(&testObjectEncoder{}),
					String("extra", "value"),
				},
			},
			"[ERR] › username=username code=42 extra=value\n",
		},
		{
			"FieldsError",
			Entry{
				Fields: []Field{
					Error(&verboseError{"short", "verbose"}),
				},
			},
			"[ERR] › error=short error.verbose=verbose\n",
		},
		{
			"FieldsNilError",
			Entry{
				Fields: []Field{
					NamedError("error", nil),
				},
			},
			"[ERR] › error=<nil>\n",
		},
		{
			"FieldsBytes",
			Entry{
				Fields: []Field{
					Bytes("bytes", []byte{0x42}),
				},
			},
			"[ERR] › bytes=\"Qg==\"\n",
		},
		{
			"FieldsStrings",
			Entry{
				Fields: []Field{
					Strings("strings", []string{"a", "b"}),
				},
			},
			"[ERR] › strings=[a,b]\n",
		},
		{
			"FieldsStringer",
			Entry{
				Fields: []Field{
					Stringer("stringer", time.Second),
				},
			},
			"[ERR] › stringer=1s\n",
		},
		{
			"FieldsNilStringer",
			Entry{
				Fields: []Field{
					Stringer("stringer", nil),
				},
			},
			"[ERR] › stringer=nil\n",
		},
		{
			"FieldsFormatter",
			Entry{
				Fields: []Field{
					Formatter("fmt", "%d", 42),
				},
			},
			"[ERR] › fmt=42\n",
		},
		{
			"FieldsAny",
			Entry{
				Fields: []Field{
					Any("any", &struct{ Field string }{Field: "42"}),
				},
			},
			"[ERR] › any={\"Field\":\"42\"}\n",
		},
		{
			"FieldsGroup",
			Entry{
				Fields: []Field{
					Group("request", String("id", "abc"), Int("status", 200)),
				},
			},
			"[ERR] › request.id=abc request.status=200\n",
		},
		{
			"FieldsGroupEmpty",
			Entry{
				Fields: []Field{
					Group("empty"),
				},
			},
			"[ERR]\n",
		},
		{
			"FieldsGroupNested",
			Entry{
				Fields: []Field{
					Group("outer", String("a", "1"), Group("inner", Int("b", 2))),
				},
			},
			"[ERR] › outer.a=1 outer.inner.b=2\n",
		},
		{
			"FieldsLoggerBag",
			Entry{
				LoggerBag: NewBag(Int("int", 42)),
			},
			"[ERR] › int=42\n",
		},
		{
			"FieldsLoggerBagFirst",
			Entry{
				LoggerBag: NewBag(Int("int", 42)),
				Fields:    []Field{String("string", "42")},
			},
			"[ERR] › int=42 string=42\n",
		},
		{
			"WithGroup",
			Entry{
				LoggerBag: NewBag(String("a", "1")).WithGroup("http").With(String("method", "GET")),
				Fields:    []Field{Int("status", 200)},
			},
			"[ERR] › a=1 http.method=GET http.status=200\n",
		},
		{
			"WithGroupNested",
			Entry{
				LoggerBag: NewBag().WithGroup("http").WithGroup("request").With(String("path", "/api")),
				Fields:    []Field{Int("status", 200)},
			},
			"[ERR] › http.request.path=/api http.request.status=200\n",
		},
		{
			"WithGroupNoFields",
			Entry{
				LoggerBag: NewBag().WithGroup("http"),
				Fields:    []Field{Int("status", 200)},
			},
			"[ERR] › http.status=200\n",
		},
		{
			"WithGroupAndWith",
			Entry{
				LoggerBag: NewBag().WithGroup("http").With(String("method", "GET")).With(String("path", "/api")),
				Fields:    []Field{Int("status", 200)},
			},
			"[ERR] › http.method=GET http.path=/api http.status=200\n",
		},
		{
			"WithGroupEmpty",
			Entry{
				LoggerBag: NewBag().WithGroup("http"),
			},
			"[ERR]\n",
		},
		{
			"WithGroupNestedEmpty",
			Entry{
				LoggerBag: NewBag().WithGroup("http").WithGroup("request"),
			},
			"[ERR]\n",
		},
		{
			"WithGroupPartiallyEmpty",
			Entry{
				LoggerBag: NewBag().WithGroup("http").With(String("host", "localhost")).WithGroup("request"),
			},
			"[ERR] › http.host=localhost\n",
		},
		{
			"StringWithSpaces",
			Entry{
				Fields: []Field{String("msg", "hello world")},
			},
			"[ERR] › msg=\"hello world\"\n",
		},
		{
			"MessageAndFields",
			Entry{
				Text:   "request handled",
				Fields: []Field{String("method", "GET"), Int("status", 200)},
			},
			"[ERR] request handled › method=GET status=200\n",
		},
	}

	// Color test: full message with all elements.
	t.Run("FullColorOutput", func(t *testing.T) {
		colorEnc := NewTextEncoder(TextEncoderConfig{
			DisableFieldTime:   true,
			DisableFieldCaller: true,
		})
		e := Entry{
			Level:      LevelInfo,
			LoggerName: "runvm",
			Text:       "started processing task",
			LoggerBag: NewBag(
				String("dc-name", "za01-cloud"),
			),
			Fields: []Field{
				String("task-type", "runvm_vm_finalize"),
				Int("status", 200),
				Bool("ok", true),
				Duration("elapsed", 42*time.Millisecond),
			},
		}
		b, err := colorEnc.Encode(e)
		require.NoError(t, err)
		got := b.String()
		b.Free()

		// Verify structure: [level] name: message › fields
		assert.Contains(t, got, "[")          // level bracket
		assert.Contains(t, got, "INF")        // level text
		assert.Contains(t, got, "]")          // level bracket
		assert.Contains(t, got, "runvm:")     // logger name
		assert.Contains(t, got, "started processing task") // message
		assert.Contains(t, got, "›")          // field separator
		assert.Contains(t, got, "dc-name")    // bag field key
		assert.Contains(t, got, "za01-cloud") // bag field value
		assert.Contains(t, got, "task-type")  // entry field key
		assert.Contains(t, got, "200")        // numeric value
		assert.Contains(t, got, "true")       // bool value
		assert.Contains(t, got, "42ms")       // duration value

		// Verify ANSI codes present (not NoColor):
		assert.Contains(t, got, "\x1b[")      // has escape sequences
		assert.Contains(t, got, "\x1b[0;2m")  // dim (brackets, separators)
		assert.Contains(t, got, "\x1b[1;36m") // bold cyan (INF)
		assert.Contains(t, got, "\x1b[1m")    // bold (message)
		assert.Contains(t, got, "\x1b[0;2;3m") // dim italic (logger name)
		assert.Contains(t, got, "\x1b[94;3m") // bright blue italic (field keys)
		assert.Contains(t, got, "\x1b[32m")   // green (numeric values)

		t.Logf("output:\n%s", got)
	})

	enc := NewTextEncoder(TextEncoderConfig{
		NoColor:          true,
		DisableFieldTime: true,
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := enc.Encode(tc.entry)
			require.NoError(t, err)

			if tc.golden != "" {
				assert.Equal(t, tc.golden, b.String())
			} else {
				// CallerPC: line number is dynamic, just check key presence.
				assert.Contains(t, b.String(), "→ logf/textencoder_test.go:")
			}
			b.Free()
		})
	}
}

func TestTextEncoderBuilder(t *testing.T) {
	enc := Text().NoColor().DisableTime().DisableLevel().DisableMsg().DisableName().DisableCaller().Build()
	require.NotNil(t, enc)

	e := Entry{
		Text:       "hello",
		Level:      LevelInfo,
		LoggerName: "test",
		Fields:     []Field{String("k", "v")},
	}
	b, err := enc.Encode(e)
	require.NoError(t, err)
	got := b.String()
	b.Free()

	// All standard fields disabled, only entry fields should appear.
	assert.NotContains(t, got, "[INF]")
	assert.NotContains(t, got, "hello")
	assert.NotContains(t, got, "test:")
	assert.Contains(t, got, "k=v")
}

func TestTextEncoderBuilderDefaults(t *testing.T) {
	// Build with no options — should produce valid output with color.
	enc := Text().Build()
	b, err := enc.Encode(Entry{Text: "msg", Level: LevelWarn})
	require.NoError(t, err)
	got := b.String()
	b.Free()

	assert.Contains(t, got, "WRN")
	assert.Contains(t, got, "msg")
	assert.Contains(t, got, "\x1b[") // ANSI codes present
}

func TestTextEncoderBuilderCustomEncoders(t *testing.T) {
	customTime := func(t time.Time, te TypeEncoder) {
		te.EncodeTypeString("CUSTOM_TIME")
	}
	customDuration := func(d time.Duration, te TypeEncoder) {
		te.EncodeTypeString("CUSTOM_DUR")
	}
	customLevel := func(l Level, te TypeEncoder) {
		te.EncodeTypeString("LVL")
	}
	customCaller := func(pc uintptr, te TypeEncoder) {
		te.EncodeTypeString("CUSTOM_CALLER")
	}
	customError := func(k string, err error, fe FieldEncoder) {
		fe.EncodeFieldString(k, "CUSTOM_ERR:"+err.Error())
	}

	enc := Text().
		NoColor().
		EncodeTime(customTime).
		EncodeDuration(customDuration).
		EncodeLevel(customLevel).
		EncodeCaller(customCaller).
		EncodeError(customError).
		Build()

	e := Entry{
		Text:     "test",
		Level:    LevelError,
		Time:     time.Now(),
		CallerPC: CallerPC(0),
		Fields: []Field{
			Duration("d", time.Second),
			NamedError("err", &verboseError{"oops", "detail"}),
		},
	}
	b, err := enc.Encode(e)
	require.NoError(t, err)
	got := b.String()
	b.Free()

	assert.Contains(t, got, "CUSTOM_TIME")
	assert.Contains(t, got, "LVL")
	assert.Contains(t, got, "CUSTOM_DUR")
	assert.Contains(t, got, "CUSTOM_CALLER")
	assert.Contains(t, got, "CUSTOM_ERR:oops")
}
