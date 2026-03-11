package logf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRFC3339TimeEncoder(t *testing.T) {
	tm := time.Unix(1542266559, 305941000).UTC()
	enc := testTypeEncoder{}
	RFC3339TimeEncoder(tm, &enc)

	assert.EqualValues(t, "2018-11-15T07:22:39Z", enc.result)
}

func TestRFC3339NanoTimeEncoder(t *testing.T) {
	tm := time.Unix(1542266559, 305941000).UTC()
	enc := testTypeEncoder{}
	RFC3339NanoTimeEncoder(tm, &enc)

	assert.EqualValues(t, "2018-11-15T07:22:39.305941Z", enc.result)
}

func TestLayoutTimeEncoder(t *testing.T) {
	tm := time.Unix(1542266559, 305941000).UTC()
	enc := testTypeEncoder{}
	LayoutTimeEncoder(time.StampNano)(tm, &enc)

	assert.EqualValues(t, "Nov 15 07:22:39.305941000", enc.result)
}

func TestUnixNanoTimeEncoder(t *testing.T) {
	tm := time.Unix(1542266559, 305941000).UTC()
	enc := testTypeEncoder{}
	UnixNanoTimeEncoder(tm, &enc)

	assert.EqualValues(t, 1542266559305941000, enc.result)
}

func TestNanoDurationEncoder(t *testing.T) {
	d := time.Duration(66559305941000)
	enc := testTypeEncoder{}
	NanoDurationEncoder(d, &enc)

	assert.EqualValues(t, 66559305941000, enc.result)
}

func TestFloatSecondsDurationEncoder(t *testing.T) {
	d := time.Duration(66559305941000)
	enc := testTypeEncoder{}
	FloatSecondsDurationEncoder(d, &enc)

	assert.InDelta(t, 66559.305941, enc.result, 0.0000005)
}

func TestStringDurationEncoder(t *testing.T) {
	d := time.Duration(66559305941000)
	enc := testTypeEncoder{}
	StringDurationEncoder(d, &enc)

	assert.EqualValues(t, "18h29m19.305941s", enc.result)
}

func TestAppendDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		// Zero.
		{0, "0s"},

		// Nanoseconds.
		{1 * time.Nanosecond, "1ns"},
		{999 * time.Nanosecond, "999ns"},

		// Microseconds.
		{1 * time.Microsecond, "1µs"},
		{1500 * time.Nanosecond, "1.5µs"},
		{999*time.Microsecond + 999*time.Nanosecond, "999.999µs"},

		// Milliseconds.
		{1 * time.Millisecond, "1ms"},
		{42 * time.Millisecond, "42ms"},
		{1500 * time.Microsecond, "1.5ms"},
		{999*time.Millisecond + 999*time.Microsecond + 999*time.Nanosecond, "999.999999ms"},

		// Seconds.
		{1 * time.Second, "1s"},
		{1500 * time.Millisecond, "1.5s"},
		{time.Second + 305941*time.Microsecond, "1.305941s"},

		// Minutes.
		{1 * time.Minute, "1m0s"},
		{time.Minute + 30*time.Second, "1m30s"},
		{time.Minute + 30*time.Second + 500*time.Millisecond, "1m30.5s"},

		// Hours.
		{1 * time.Hour, "1h0m0s"},
		{time.Hour + 2*time.Minute + 3*time.Second, "1h2m3s"},
		{18*time.Hour + 29*time.Minute + 19*time.Second + 305941*time.Microsecond, "18h29m19.305941s"},
		{24 * time.Hour, "24h0m0s"},
		{100*time.Hour + 500*time.Millisecond, "100h0m0.5s"},

		// Negative.
		{-1 * time.Nanosecond, "-1ns"},
		{-42 * time.Millisecond, "-42ms"},
		{-1*time.Hour - 30*time.Minute, "-1h30m0s"},

		// Large.
		{290*24*time.Hour + 23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond, "6983h59m59.999999999s"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			var buf [32]byte
			got := string(appendDuration(buf[:0], tt.d))
			// Verify against stdlib.
			assert.Equal(t, tt.d.String(), got, "must match time.Duration.String()")
			assert.Equal(t, tt.want, got)
		})
	}
}
