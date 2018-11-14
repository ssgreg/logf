package logf

import "time"

// Entry holds a single log message and fields.
type Entry struct {
	// LoggerID specifies a unique logger identifies.
	LoggerID int32

	// LoggerName specifies a non-unique name of a logger.
	// Can be empty.
	LoggerName string

	// DeriviedFields specifies logger data fields including fields of
	// logger parents. The earliest fields (parent's fields) go first.
	DerivedFields []Field

	// Fields specifies data fields of a log message.
	Fields []Field

	// Level specifies a severity level of a log message.
	Level Level

	// Time specifies a timestamp of a log message.
	Time time.Time

	// Text specifies a text message of a log message.
	Text string

	// Caller specifies file:line info about an Entry's caller.
	Caller EntryCaller
}
