package logf

import (
	"runtime"
	"time"
	"unsafe"
)

// TimeEncoder is a function that formats a time.Time into the log output
// via the TypeEncoder. Swap it out to control timestamp format globally.
type TimeEncoder func(time.Time, TypeEncoder)

// DurationEncoder is a function that formats a time.Duration into the log
// output via the TypeEncoder.
type DurationEncoder func(time.Duration, TypeEncoder)

// RFC3339TimeEncoder formats timestamps as RFC3339 strings (e.g.
// "2006-01-02T15:04:05Z07:00"). This is the default for JSON output.
func RFC3339TimeEncoder(t time.Time, e TypeEncoder) {
	var timeBuf [64]byte
	var b []byte
	b = timeBuf[:0]
	b = t.AppendFormat(b, time.RFC3339)
	e.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

// RFC3339NanoTimeEncoder formats timestamps as RFC3339 strings with
// nanosecond precision (e.g. "2006-01-02T15:04:05.999999999Z07:00").
func RFC3339NanoTimeEncoder(t time.Time, e TypeEncoder) {
	var timeBuf [64]byte
	var b []byte
	b = timeBuf[:0]
	b = t.AppendFormat(b, time.RFC3339Nano)
	e.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

// LayoutTimeEncoder returns a TimeEncoder that formats timestamps using the
// given Go time layout string (same format as time.Format).
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

// UnixNanoTimeEncoder formats timestamps as integer nanoseconds since the
// Unix epoch. Compact and machine-friendly, but not human-readable.
func UnixNanoTimeEncoder(t time.Time, e TypeEncoder) {
	e.EncodeTypeInt64(t.UnixNano())
}

// NanoDurationEncoder formats durations as integer nanoseconds.
func NanoDurationEncoder(d time.Duration, e TypeEncoder) {
	e.EncodeTypeInt64(int64(d))
}

// FloatSecondsDurationEncoder formats durations as floating-point seconds
// (e.g. 1.5 for one and a half seconds).
func FloatSecondsDurationEncoder(d time.Duration, e TypeEncoder) {
	e.EncodeTypeFloat64(float64(d) / float64(time.Second))
}

// StringDurationEncoder formats durations as human-readable strings like
// "4.5s", "300ms", or "1h2m3s" — the same format as time.Duration.String()
// but without allocating. This is the default.
func StringDurationEncoder(d time.Duration, m TypeEncoder) {
	var buf [32]byte
	var b []byte
	b = buf[:0]
	b = appendDuration(b, d)
	m.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

// appendDuration formats d identically to time.Duration.String but
// appends into the provided buffer instead of allocating a new string.
func appendDuration(b []byte, d time.Duration) []byte {
	if d == 0 {
		return append(b, '0', 's')
	}

	neg := d < 0
	if neg {
		b = append(b, '-')
		d = -d
	}

	u := uint64(d)

	// Hours.
	hasHours := u >= uint64(time.Hour)
	if hasHours {
		h := u / uint64(time.Hour)
		u -= h * uint64(time.Hour)
		b = appendUint(b, h)
		b = append(b, 'h')
	}
	// Minutes — always printed if hours were printed.
	hasMinutes := u >= uint64(time.Minute)
	if hasHours || hasMinutes {
		m := u / uint64(time.Minute)
		u -= m * uint64(time.Minute)
		b = appendUint(b, m)
		b = append(b, 'm')
	}
	// Seconds — always printed if hours or minutes were printed.
	if hasHours || hasMinutes {
		b = appendFrac(b, u, 1_000_000_000, 9)
		b = append(b, 's')
		return b
	}
	// Sub-second only (no hours/minutes prefix).
	if u >= uint64(time.Second) {
		b = appendFrac(b, u, 1_000_000_000, 9)
		b = append(b, 's')
		return b
	}
	if u >= uint64(time.Millisecond) {
		b = appendFrac(b, u, 1_000_000, 6)
		b = append(b, 'm', 's')
		return b
	}
	if u >= uint64(time.Microsecond) {
		b = appendFrac(b, u, 1_000, 3)
		b = append(b, '\xc2', '\xb5', 's') // µs UTF-8
		return b
	}
	b = appendUint(b, u)
	b = append(b, 'n', 's')
	return b
}

// appendFrac appends ns/unit as "whole.frac" with trailing zeros trimmed.
// fracDigits is the number of fractional digits (e.g. 9 for seconds).
func appendFrac(b []byte, ns, unit uint64, fracDigits int) []byte {
	whole := ns / unit
	frac := ns % unit

	b = appendUint(b, whole)
	if frac == 0 {
		return b
	}
	b = append(b, '.')

	// Write fractional digits, then trim trailing zeros.
	var tmp [9]byte
	for i := fracDigits - 1; i >= 0; i-- {
		tmp[i] = byte('0' + frac%10)
		frac /= 10
	}
	end := fracDigits
	for end > 0 && tmp[end-1] == '0' {
		end--
	}
	b = append(b, tmp[:end]...)
	return b
}

// appendUint appends the decimal representation of v.
func appendUint(b []byte, v uint64) []byte {
	if v == 0 {
		return append(b, '0')
	}
	var tmp [20]byte
	i := len(tmp)
	for v > 0 {
		i--
		tmp[i] = byte('0' + v%10)
		v /= 10
	}
	return append(b, tmp[i:]...)
}

// noescape hides a pointer from escape analysis.  noescape is
// the identity function but escape analysis doesn't think the
// output depends on the input.  noescape is inlined and currently
// compiles down to zero instructions.
// USE CAREFULLY!
//
//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)

	return unsafe.Pointer(x ^ 0) //nolint:staticcheck // intentional no-op to fool escape analysis
}
