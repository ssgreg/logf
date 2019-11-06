package logf

import (
	"runtime"
	"time"
	"unsafe"
)

// TimeEncoder is the function type to encode the given Time.
type TimeEncoder func(time.Time, TypeEncoder)

// DurationEncoder is the function type to encode the given Duration.
type DurationEncoder func(time.Duration, TypeEncoder)

// RFC3339TimeEncoder encodes the given Time as a string using RFC3339 layout.
func RFC3339TimeEncoder(t time.Time, e TypeEncoder) {
	var timeBuf [64]byte
	var b []byte
	b = timeBuf[:0]
	b = t.AppendFormat(b, time.RFC3339)
	e.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

// RFC3339NanoTimeEncoder encodes the given Time as a string using
// RFC3339Nano layout.
func RFC3339NanoTimeEncoder(t time.Time, e TypeEncoder) {
	var timeBuf [64]byte
	var b []byte
	b = timeBuf[:0]
	b = t.AppendFormat(b, time.RFC3339Nano)
	e.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

// LayoutTimeEncoder encodes the given Time as a string using custom layout.
func LayoutTimeEncoder(layout string) TimeEncoder {
	return func(t time.Time, m TypeEncoder) {
		var timeBuf [64]byte
		var b []byte
		b = timeBuf[:0]
		b = t.AppendFormat(b, layout)
		m.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
		runtime.KeepAlive(&b)
	}
}

// UnixNanoTimeEncoder encodes the given Time as a Unix time, the number of
// of nanoseconds elapsed since January 1, 1970 UTC.
func UnixNanoTimeEncoder(t time.Time, e TypeEncoder) {
	e.EncodeTypeInt64(t.UnixNano())
}

// NanoDurationEncoder encodes the given Duration as a number of nanoseconds.
func NanoDurationEncoder(d time.Duration, e TypeEncoder) {
	e.EncodeTypeInt64(int64(d))
}

// FloatSecondsDurationEncoder encodes the given Duration to a floating-point
// number of seconds elapsed.
func FloatSecondsDurationEncoder(d time.Duration, e TypeEncoder) {
	e.EncodeTypeFloat64(float64(d) / float64(time.Second))
}

// StringDurationEncoder encodes the given Duration as a string using
// Stringer interface.
func StringDurationEncoder(d time.Duration, m TypeEncoder) {
	m.EncodeTypeString(d.String())
}
