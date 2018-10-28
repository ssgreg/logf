package logf

import (
	"sync/atomic"
	"time"
	"unsafe"
)

var nextID int32

func NewDisabled() *Logger {
	return NewLogger(Discard, nil)
}

func NewLogger(level Level, w ChannelWriter) *Logger {
	return &Logger{
		level: level,
		ch:    w,
	}
}

type Logger struct {
	level      Level
	name       string
	id         int32
	fields     []Field
	ch         ChannelWriter
	hasCaller  bool
	callerSkip int
}

func (l *Logger) WithName(n string) *Logger {
	if n == "" {
		return l
	}

	cc := l.clone()
	if cc.name == "" {
		cc.name = n
	} else {
		cc.name += "." + n
	}

	return cc
}

func (l *Logger) WithCaller() *Logger {
	cc := l.clone()
	cc.hasCaller = true

	return cc
}

func (l *Logger) WithCallerSkip(skip int) *Logger {
	cc := l.clone()
	cc.callerSkip = skip

	return cc
}

func (l *Logger) With(fs ...Field) *Logger {
	// This code attempts to archive optimum performance with minimum
	// allocations count. Do not change it unless the folowing benchmarks
	// will show a better performance:
	// - BenchmarkCreateContextLogger
	// - BenchmarkCreateContextWithAccumulatedContextLogger

	var cc *Logger
	if len(l.fields) == 0 {
		// The fastest way. Use passed 'fs' as is.
		cc = l.clone()
		cc.fields = fs
	} else {
		// The less efficient path forces us to copy parent's fields.
		c := make([]Field, 0, len(l.fields)+len(fs))
		// TODO: change order
		c = append(c, l.fields...)
		c = append(c, fs...)

		cc = l.clone()
		cc.fields = c
	}

	for i := range cc.fields[len(l.fields):] {
		snapshotField(&cc.fields[i])
	}

	return cc
}

func (l *Logger) Debug(text string, fs ...Field) {
	if l.level < Debug {
		return
	}

	l.log(Debug, text, fs)
}

func (l *Logger) Info(text string, fs ...Field) {
	if l.level < Info {
		return
	}

	l.log(Info, text, fs)
}

func (l *Logger) log(lv Level, text string, fs []Field) {
	for i := range fs {
		snapshotField(&fs[i])
	}

	e := Entry{l.id, l.fields, fs, lv, time.Now(), text, l.name, EntryCaller{}}
	if l.hasCaller {
		e.Caller = NewEntryCaller(2 + l.callerSkip)
	}

	l.ch.Write(e)
}

func (l *Logger) clone() *Logger {
	return &Logger{
		level:  l.level,
		ch:     l.ch,
		fields: nil,
		id:     atomic.AddInt32(&nextID, 1),
		name:   l.name,
	}
}

func snapshotField(f *Field) {
	if f.Type&FieldTypeRawMask != 0 {
		switch f.Type {
		case FieldTypeRawBytes:
			snapshotRawBytes(f)
		case FieldTypeRawBytesToBools:
			snapshotRawBytesToBools(f)
		case FieldTypeRawBytesToInts64:
			snapshotRawBytesToInts64(f)
		case FieldTypeRawBytesToInts32:
			snapshotRawBytesToInts32(f)
		case FieldTypeRawBytesToInts16:
			snapshotRawBytesToInts16(f)
		case FieldTypeRawBytesToInts8:
			snapshotRawBytesToInts8(f)
		case FieldTypeRawBytesToUints64:
			snapshotRawBytesToUints64(f)
		case FieldTypeRawBytesToUints32:
			snapshotRawBytesToUints32(f)
		case FieldTypeRawBytesToUints16:
			snapshotRawBytesToUints16(f)
		case FieldTypeRawBytesToUints8:
			snapshotRawBytesToUints8(f)
		case FieldTypeRawBytesToFloats64:
			snapshotRawBytesToFloats64(f)
		case FieldTypeRawBytesToFloats32:
			snapshotRawBytesToFloats32(f)
		case FieldTypeRawBytesToDurations:
			snapshotRawBytesToDurations(f)
		}
	}
	if f.Type == FieldTypeAny {
		if f.Any == nil {
			return
		}
		switch rv := f.Any.(type) {
		case Snapshotter:
			f.Any = rv.TakeSnapshot()
		}
	}
}

func snapshotRawBytes(f *Field) {
	cc := make([]byte, len(f.Bytes))
	copy(cc, f.Bytes)
	f.Bytes = cc
	f.Type = FieldTypeBytes
}

func snapshotRawBytesToBools(f *Field) {
	s := *(*[]bool)(unsafe.Pointer(&f.Bytes))
	cc := make([]bool, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToBools
}

func snapshotRawBytesToInts64(f *Field) {
	s := *(*[]int64)(unsafe.Pointer(&f.Bytes))
	cc := make([]int64, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToInts64
}

func snapshotRawBytesToInts32(f *Field) {
	s := *(*[]int32)(unsafe.Pointer(&f.Bytes))
	cc := make([]int32, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToInts32
}

func snapshotRawBytesToInts16(f *Field) {
	s := *(*[]int16)(unsafe.Pointer(&f.Bytes))
	cc := make([]int16, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToInts16
}

func snapshotRawBytesToInts8(f *Field) {
	s := *(*[]int8)(unsafe.Pointer(&f.Bytes))
	cc := make([]int8, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToInts8
}

func snapshotRawBytesToUints64(f *Field) {
	s := *(*[]uint64)(unsafe.Pointer(&f.Bytes))
	cc := make([]uint64, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToUints64
}

func snapshotRawBytesToUints32(f *Field) {
	s := *(*[]uint32)(unsafe.Pointer(&f.Bytes))
	cc := make([]uint32, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToUints32
}

func snapshotRawBytesToUints16(f *Field) {
	s := *(*[]uint16)(unsafe.Pointer(&f.Bytes))
	cc := make([]uint16, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToUints16
}

func snapshotRawBytesToUints8(f *Field) {
	s := *(*[]uint8)(unsafe.Pointer(&f.Bytes))
	cc := make([]uint8, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToUints8
}

func snapshotRawBytesToFloats64(f *Field) {
	s := *(*[]float64)(unsafe.Pointer(&f.Bytes))
	cc := make([]float64, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToFloats64
}

func snapshotRawBytesToFloats32(f *Field) {
	s := *(*[]float32)(unsafe.Pointer(&f.Bytes))
	cc := make([]float32, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToFloats32
}

func snapshotRawBytesToDurations(f *Field) {
	s := *(*[]time.Duration)(unsafe.Pointer(&f.Bytes))
	cc := make([]time.Duration, len(s))
	copy(cc, s)
	f.Bytes = *(*[]byte)(unsafe.Pointer(&cc))
	f.Type = FieldTypeBytesToDurations
}
