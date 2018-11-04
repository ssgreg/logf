package logf

import (
	"runtime"
	"time"
	"unsafe"
)

type TimeEncoder func(time.Time, TypeEncoder)
type DurationEncoder func(time.Duration, TypeEncoder)

func RFC3339TimeEncoder(t time.Time, m TypeEncoder) {
	var timeBuf [64]byte
	var b []byte
	b = timeBuf[:0]
	b = t.AppendFormat(b, time.RFC3339)
	m.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

func RFC3339NanoTimeEncoder(t time.Time, m TypeEncoder) {
	var timeBuf [64]byte
	var b []byte
	b = timeBuf[:0]
	b = t.AppendFormat(b, time.RFC3339Nano)
	m.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

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

func UnixNanoTimeEncoder(t time.Time, m TypeEncoder) {
	m.EncodeTypeInt64(t.UnixNano())
}

func NanoDurationEncoder(d time.Duration, m TypeEncoder) {
	m.EncodeTypeInt64(int64(d))
}

func StringDurationEncoder(d time.Duration, m TypeEncoder) {
	m.EncodeTypeString(d.String())
}
