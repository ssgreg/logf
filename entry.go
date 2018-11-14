package logf

import "time"

// Entry holds a single log message and fields.
type Entry struct {
	LoggerID      int32
	DerivedFields []Field
	Fields        []Field
	Level         Level
	Time          time.Time
	Text          string
	LoggerName    string
	Caller        EntryCaller
}
