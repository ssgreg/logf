package logf

import (
	"fmt"
	"time"
)

const (
	defaultFieldKeyMsg   = "msg"
	defaultFieldKeyTime  = "time"
	defaultFieldKeyLevel = "level"
)

// JSONFormatter TODO
type JSONFormatter struct {
	// Default fields names.
	FieldKeyMsg   string
	FieldKeyTime  string
	FieldKeyLevel string

	// TimestampFormat uses the default Go timestamp layout.
	// Default is time.RFC3339.
	TimestampFormat string

	// DisableTimestamp disables/enables timestamp in the output.
	// Default is false.
	DisableTimestamp bool

	configured bool
}

func (f *JSONFormatter) Format(buf *Buffer, entry *Entry) error {
	f.setDefaults()

	buf.AppendString("{")
	if !f.DisableTimestamp {
		EscapeString(buf, f.FieldKeyTime)
		buf.AppendString(":\"")
		buf.AppendString(entry.time.Format(f.TimestampFormat))
		buf.AppendString("\",")
	}
	EscapeString(buf, f.FieldKeyLevel)
	buf.AppendString(":\"")
	buf.AppendString(entry.level.String())
	buf.AppendString("\",")
	EscapeString(buf, f.FieldKeyMsg)
	buf.AppendString(":")
	if len(entry.args) == 0 {
		EscapeString(buf, entry.format)
	} else {
		EscapeString(buf, fmt.Sprint(entry.args...))
	}

	var fields []Field
	for l := entry.fields; l != nil; fields, l = l.Fields() {
		for _, field := range fields {
			buf.AppendString(",")
			EscapeString(buf, field.Key)
			buf.AppendString(":")

			// if !KnownTypeToBuf(buf, field.Value) {
			// 	b, err := json.Marshal(field.Value)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	buf.AppendBytes(b)
			// }
		}
	}

	buf.AppendString("}\n")
	return nil
}

func (f *JSONFormatter) setDefaults() {
	if f.configured {
		return
	}
	f.configured = true

	if f.FieldKeyMsg == "" {
		f.FieldKeyMsg = defaultFieldKeyMsg
	}
	if f.FieldKeyTime == "" {
		f.FieldKeyTime = defaultFieldKeyTime
	}
	if f.FieldKeyLevel == "" {
		f.FieldKeyLevel = defaultFieldKeyLevel
	}

	if f.TimestampFormat == "" {
		f.TimestampFormat = time.RFC3339
	}
}
