package logf

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
	"unsafe"
)

type TimeFormatter func(time.Time, TypeMarshaller)

func RFC3339TimeFormatter(t time.Time, m TypeMarshaller) {
	var timeBuf [64]byte
	var b []byte
	b = timeBuf[:0]
	b = t.AppendFormat(b, time.RFC3339)
	m.MarshalUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

func RFC3339NanoTimeFormatter(t time.Time, m TypeMarshaller) {
	var timeBuf [64]byte
	var b []byte
	b = timeBuf[:0]
	b = t.AppendFormat(b, time.RFC3339Nano)
	m.MarshalUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

func UnixNanoTimeFormatter(t time.Time, m TypeMarshaller) {
	m.MarshalInt64(t.UnixNano())
}

type DurationFormatter func(time.Duration, TypeMarshaller)

func NanoDurationFormatter(d time.Duration, m TypeMarshaller) {
	m.MarshalInt64(int64(d))
}

func StringDurationFormatter(d time.Duration, m TypeMarshaller) {
	m.MarshalString(d.String())
}

type ErrorFormatter func(string, error, FieldMarshaller)

func DefaultErrorFormatter(k string, e error, m FieldMarshaller) {
	msg := e.Error()
	m.MarshalFieldString(k, msg)

	switch e.(type) {
	case fmt.Formatter:
		verbose := fmt.Sprintf("%+v", e)
		if verbose != msg {
			m.MarshalFieldString(k+".verbose", verbose)
		}
	}
}

type CallerFormatter func(EntryCaller, TypeMarshaller)

func ShortCallerFormatter(c EntryCaller, m TypeMarshaller) {
	var callerBuf [64]byte
	var b []byte
	b = callerBuf[:0]
	b = append(b, c.FileWithPackage()...)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(c.Line), 10)

	m.MarshalUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

func FullCallerFormatter(c EntryCaller, m TypeMarshaller) {
	var callerBuf [256]byte
	var b []byte
	b = callerBuf[:0]
	b = append(b, c.File...)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(c.Line), 10)

	m.MarshalUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}
