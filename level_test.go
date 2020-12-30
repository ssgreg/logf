package logf

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelChecker(t *testing.T) {
	type LevelCheck struct {
		level   Level
		enabled bool
	}

	cases := []struct {
		level   Level
		goldens []LevelCheck
	}{
		{
			LevelError,
			[]LevelCheck{{LevelError, true}, {LevelWarn, false}, {LevelInfo, false}, {LevelDebug, false}},
		},
		{
			LevelWarn,
			[]LevelCheck{{LevelError, true}, {LevelWarn, true}, {LevelInfo, false}, {LevelDebug, false}},
		},
		{
			LevelInfo,
			[]LevelCheck{{LevelError, true}, {LevelWarn, true}, {LevelInfo, true}, {LevelDebug, false}},
		},
		{
			LevelDebug,
			[]LevelCheck{{LevelError, true}, {LevelWarn, true}, {LevelInfo, true}, {LevelDebug, true}},
		},
	}

	for _, cs := range cases {
		checker := cs.level.LevelChecker()
		for _, golden := range cs.goldens {
			assert.Equal(t, golden.enabled, checker(golden.level), "%q checks with %q", cs.level, golden.level)
		}
	}
}

func TestLevelString(t *testing.T) {
	cases := []struct {
		level        Level
		strGolden    string
		capStrGolden string
	}{
		{LevelError, "error", "ERROR"},
		{LevelWarn, "warn", "WARN"},
		{LevelInfo, "info", "INFO"},
		{LevelDebug, "debug", "DEBUG"},
		{Level(42), "unknown", "UNKNOWN"},
	}

	for _, cs := range cases {
		assert.Equal(t, cs.strGolden, cs.level.String(), "check level %d String", int(cs.level))
		assert.Equal(t, cs.capStrGolden, cs.level.UpperCaseString(), "check level %d UpperCaseString", int(cs.level))
	}
}

func TestLevelFromString(t *testing.T) {
	cases := []struct {
		checking []string
		golden   Level
	}{
		{[]string{"error", "ERROR"}, LevelError},
		{[]string{"warn", "WARN", "warning", "WARNING"}, LevelWarn},
		{[]string{"info", "INFO", "information", "INFORMATION"}, LevelInfo},
		{[]string{"debug", "DEBUG"}, LevelDebug},
	}

	for _, cs := range cases {
		for _, lvl := range cs.checking {
			level, ok := LevelFromString(lvl)
			assert.True(t, ok, "validate %q", lvl)
			assert.Equal(t, cs.golden, level, "compare golden %q with %q", cs.golden, level)
		}
	}

	_, ok := LevelFromString("unknown")
	assert.False(t, ok, "fail is expected")
}

func TestDefaultLevelEncoder(t *testing.T) {
	enc := testTypeEncoder{}
	DefaultLevelEncoder(LevelError, &enc)

	assert.EqualValues(t, "error", enc.result)
}

func TestUpperCaseLevelEncoder(t *testing.T) {
	enc := testTypeEncoder{}
	UpperCaseLevelEncoder(LevelError, &enc)

	assert.EqualValues(t, "ERROR", enc.result)
}

func TestMutableLevelChecker(t *testing.T) {
	type LevelCheck struct {
		level   Level
		enabled bool
	}

	cases := []struct {
		checker LevelChecker
		goldens []LevelCheck
	}{
		{
			NewMutableLevel(LevelError).LevelChecker(),
			[]LevelCheck{{LevelError, true}, {LevelWarn, false}, {LevelInfo, false}, {LevelDebug, false}},
		},
		{
			NewMutableLevel(LevelWarn).LevelChecker(),
			[]LevelCheck{{LevelError, true}, {LevelWarn, true}, {LevelInfo, false}, {LevelDebug, false}},
		},
		{
			NewMutableLevel(LevelInfo).LevelChecker(),
			[]LevelCheck{{LevelError, true}, {LevelWarn, true}, {LevelInfo, true}, {LevelDebug, false}},
		},
		{
			NewMutableLevel(LevelDebug).LevelChecker(),
			[]LevelCheck{{LevelError, true}, {LevelWarn, true}, {LevelInfo, true}, {LevelDebug, true}},
		},
	}

	for i, cs := range cases {
		for _, golden := range cs.goldens {
			assert.Equal(t, golden.enabled, cs.checker(golden.level), "%d checks with %q", i, golden.level)
		}
	}
}

func TestMutableLevel(t *testing.T) {
	level := NewMutableLevel(LevelError)
	assert.Equal(t, LevelError, level.Level())

	level.Set(LevelDebug)
	assert.Equal(t, LevelDebug, level.Level())
}

func TestLevelUnmarshal(t *testing.T) {
	v := struct {
		Level Level `json:"level"`
	}{}

	err := json.Unmarshal([]byte(`{"level": "warn"}`), &v)
	assert.NoError(t, err)
	assert.Equal(t, LevelWarn, v.Level)
}

func TestLevelUnmarshalInvalid(t *testing.T) {
	v := struct {
		Level Level `json:"level"`
	}{}

	err := json.Unmarshal([]byte(`{"level": "some-invalid-value"}`), &v)
	assert.EqualError(t, err, `invalid logging level "some-invalid-value"`)
	assert.Equal(t, LevelError, v.Level)
}
