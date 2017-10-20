package logf

import "time"

var emptyLogger = &disabledLogger{}

type disabledLogger struct {
}

func NewDisabled() FieldLogger {
	return &disabledLogger{}
}

func (l *disabledLogger) WithInt(key string, v int) FieldLogger {
	return l
}

func (l *disabledLogger) WithInts(key string, v []int) FieldLogger {
	return l
}

func (l *disabledLogger) WithFloat64(key string, v float64) FieldLogger {
	return l
}

func (l *disabledLogger) WithAny(string, interface{}) FieldLogger {
	return l
}

func (l *disabledLogger) WithStr(k string, v string) FieldLogger {
	return l
}

func (l *disabledLogger) WithStrs(k string, v []string) FieldLogger {
	return l
}

func (l *disabledLogger) WithTime(k string, v time.Time) FieldLogger {
	return l
}

func (l *disabledLogger) WithTimes(k string, v []time.Time) FieldLogger {
	return l
}

func (l *disabledLogger) WithErr(v error) FieldLogger {
	return l
}

func (l *disabledLogger) WithSnapshot(string, Greger) FieldLogger {
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

func (l *disabledLogger) Fields() ([]Field, FieldLogger) {
	return nil, l
}

func (l *disabledLogger) Level() Level {
	return PanicLevel
}

func (l *disabledLogger) Close() {
}
