package logf

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type formatArguments struct {
	format string
	args   []interface{}
}

// extLogger TODO
type extLogger struct {
	parent Logger
	base   baseLogger
	fields []Field
}

var CounterPool = 0

var loggerPool = &sync.Pool{
	New: func() interface{} {
		// fmt.Println("Pool")
		CounterPool++
		return &extLogger{nil, nil, make([]Field, 0, 10)}
	},
}

// TODO make a different FieldLogger interface implementation, if l == extlogger can be removed
var emptyLogger = &disabledLogger{}

// New TODO
func New(params LoggerParams) FieldLogger {
	return &extLogger{base: newBaseLogger(params)}
}

type disabledLogger struct {
}

func NewDisabled() FieldLogger {
	return &disabledLogger{}
}

func (l *disabledLogger) WithInt(key string, v int) FieldLogger {
	return l
}

func (l *disabledLogger) WithFloat64(key string, v float64) FieldLogger {
	return l
}

func (l *disabledLogger) WithAny(k string, v interface{}) FieldLogger {
	return l
}

func (l *disabledLogger) WithStr(k string, v string) FieldLogger {
	return l
}

func (l *disabledLogger) WithTime(k string, v time.Time) FieldLogger {
	return l
}

func (l *disabledLogger) WithErr(v error) FieldLogger {
	return l
}

func (l *disabledLogger) Logger() FieldLogger {
	return l
}

func (l *disabledLogger) Info() FieldLogger {
	return l
}

func (l *disabledLogger) Msg(string) {
}

func (l *disabledLogger) Msgf(string, ...interface{}) {
}

func (l *disabledLogger) Fields() ([]Field, Logger) {
	return nil, l
}

func (l *disabledLogger) Level() Level {
	return PanicLevel
}

func (l *disabledLogger) Close() {
}

// FieldLogger interface.

// WithField TODO
func (l *extLogger) WithField(key string, v interface{}) FieldLogger {
	// return &extLogger{base: l.base, parent: l, fields: []Field{{key, TakeSnapshot(v), nil}}}
	return nil
}

func (l *extLogger) WithField2(k1 string, v1 interface{}, k2 string, v2 interface{}) FieldLogger {
	// return &extLogger{base: l.base, parent: l, fields: []Field{
	// 	{k1, TakeSnapshot(v1)}, {k2, TakeSnapshot(v2)}},
	// }
	return nil
}

func (l *extLogger) WithField10(k1 string, v1 interface{}, k2 string, v2 interface{}, k3 string, v3 interface{}, k4 string, v4 interface{}, k5 string, v5 interface{}, k6 string, v6 interface{}, k7 string, v7 interface{}, k8 string, v8 interface{}, k9 string, v9 interface{}, k10 string, v10 interface{}) FieldLogger {
	// return &extLogger{base: l.base, parent: l, fields: []Field{
	// 	{k1, TakeSnapshot(v1)}, {k2, TakeSnapshot(v2)}, {k3, TakeSnapshot(v3)}, {k4, TakeSnapshot(v4)},
	// 	{k5, TakeSnapshot(v5)}, {k6, TakeSnapshot(v6)}, {k7, TakeSnapshot(v7)}, {k8, TakeSnapshot(v8)},
	// 	{k9, TakeSnapshot(v9)}, {k10, TakeSnapshot(v10)},
	// }}
	return nil
}

func (l *extLogger) WithFields(fields ...Field) FieldLogger {
	if l.base.Level() < InfoLevel {
		return emptyLogger
	}
	// return &extLogger{base: l.base, parent: l, fields: fields}
	return l
}

type MyFunc func() []Field

func (l *extLogger) WithFields1(a func() []Field) FieldLogger {
	// return &extLogger{base: l.base, parent: l, fields: nil}
	return nil
}

func (l *extLogger) WithInt(key string, v int) FieldLogger {
	// if l == emptyLogger {
	// 	return l
	// }
	l.fields = append(l.fields, Field{Key: key, Value: v})
	return l
}

func (l *extLogger) WithFloat64(key string, v float64) FieldLogger {
	// if l == emptyLogger {
	// 	return l
	// }
	l.fields = append(l.fields, Field{key, v})
	return l
}

func (l *extLogger) WithAny(k string, v interface{}) FieldLogger {
	l.fields = append(l.fields, Field{k, v})
	return l
}

func (l *extLogger) WithStr(k string, v string) FieldLogger {
	l.fields = append(l.fields, Field{k, v})
	return l
}

func (l *extLogger) WithTime(k string, v time.Time) FieldLogger {
	l.fields = append(l.fields, Field{k, v})
	return l
}

func (l *extLogger) WithErr(v error) FieldLogger {
	l.fields = append(l.fields, Field{"error", v})
	return l
}

func (l *extLogger) Logger() FieldLogger {
	return &extLogger{base: l.base, parent: l, fields: make([]Field, 0)}
}

func (l *extLogger) Info() FieldLogger {
	if l.base.Level() < InfoLevel {
		return emptyLogger
	}
	// lp := loggerPool.Get().(*extLogger)
	// lp.parent = l
	// lp.base = l.base
	// return lp

	return &extLogger{base: l.base, parent: l, fields: make([]Field, 0)}
}

func (l *extLogger) Msg(msg string) {
	// loggerPool.Put(l)
	// return
	l.base.Log(InfoLevel, msg, nil, l)
}

func (l *extLogger) Msgf(format string, args ...interface{}) {
	l.base.Log(InfoLevel, format, args, l)
}

// Добавление интерфейса с функции логирования ухудшает производительность.
// Если сделать в Field поле с конкретным типом вместо интерфейс - станет быстрее, в производительности теряем на сохранение типа сохраняемого значения и незначительно в памяти.
//
// Существенно теряем на создание нового объекта extLogger для полей
// ->
// Идея - проверить сохранение полей под мьютесом.  - Фейл. Когда нет возможности удалять филды. В разных потоках они будут плодиться.
// Идея - использовать пул Entry. - // Фейл и быстром добавлении логов и медленной записи  - при буфере 10 000 000

// func FieldAny(key string, value interface{}) Field {
// 	return Field{key, TakeSnapshot(value)}
// }

// func FieldInt(key string, value int) Field {
// 	return Field{key, value}
// }

// func FieldFloat(key string, value float64) Field {
// 	return Field{key, value}
// }

// WithFields TODO
// func (l *extLogger) WithFields(fields Fields) FieldLogger {
// 	new := make([]Field, len(fields))
// 	i := 0
// 	for k, v := range fields {
// 		new[i] = Field{k, TakeSnapshot(v)}
// 		i++
// 	}
// 	return &extLogger{base: l.base, parent: l, Fields: new}
// }

// Logger interface.

// Fields TODO
func (l *extLogger) Fields() ([]Field, Logger) {
	return l.fields, l.parent
}

// Level TODO
func (l *extLogger) Level() Level {
	return l.base.Level()
}

// Close TODO
func (l *extLogger) Close() {
	l.base.Close()
}

// Debug TODO
func (l *extLogger) Debug(args ...interface{}) {
	l.base.Log(DebugLevel, "", args, l)
}

// Info TODO
// func (l *extLogger) Info(a string) {
// 	if l.base.Level() < InfoLevel {
// 		return
// 	}
// 	l.base.Log(InfoLevel, a, nil, l)
// }

// Warn TODO
func (l *extLogger) Warn(args ...interface{}) {
	l.base.Log(WarnLevel, "", args, l)
}

// Error TODO
func (l *extLogger) Error(args ...interface{}) {
	l.base.Log(ErrorLevel, "", args, l)
}

// Debugf TODO
func (l *extLogger) Debugf(format string, args ...interface{}) {
	l.base.Log(DebugLevel, format, args, l)
}

// Infof TODO
func (l *extLogger) Infof(format string, args ...interface{}) {
	if l.base.Level() < InfoLevel {
		return
	}
	l.base.Log(InfoLevel, format, args, l)
}

// Warnf TODO
func (l *extLogger) Warnf(format string, args ...interface{}) {
	l.base.Log(WarnLevel, format, args, l)
}

// Errorf TODO
func (l *extLogger) Errorf(format string, args ...interface{}) {
	l.base.Log(ErrorLevel, format, args, l)
}

// StdLogger interface.

// Printf TODO
func (l *extLogger) Printf(format string, args ...interface{}) {
	l.base.Log(InfoLevel, format, args, l)
}

// Fatalf TODO
func (l *extLogger) Fatalf(format string, args ...interface{}) {
	l.base.Log(FatalLevel, format, args, l)
	l.Close()
	os.Exit(1)
}

// Panicf TODO
func (l *extLogger) Panicf(format string, args ...interface{}) {
	l.base.Log(PanicLevel, format, args, l)
	l.Close()
	panic(fmt.Sprintf(format, args...))
}

// Print TODO
func (l *extLogger) Print(args ...interface{}) {
	l.base.Log(InfoLevel, "", args, l)
}

// Fatal TODO
func (l *extLogger) Fatal(args ...interface{}) {
	l.base.Log(FatalLevel, "", args, l)
	l.Close()
	os.Exit(1)
}

// Panic TODO
func (l *extLogger) Panic(args ...interface{}) {
	l.base.Log(PanicLevel, "", args, l)
	l.Close()
	panic(fmt.Sprint(args...))
}

// Println TODO
func (l *extLogger) Println(args ...interface{}) {
	l.base.Log(InfoLevel, "", args, l)
}

// Fatalln TODO
func (l *extLogger) Fatalln(args ...interface{}) {
	l.base.Log(FatalLevel, "", args, l)
	l.Close()
	os.Exit(1)
}

// Panicln TODO
func (l *extLogger) Panicln(args ...interface{}) {
	l.base.Log(PanicLevel, "", args, l)
	l.Close()
	panic(fmt.Sprint(args...))
}
