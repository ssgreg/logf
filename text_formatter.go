package logf

// type TextFormatter struct {
// 	// Default fields names.
// 	FieldKeyMsg   string
// 	FieldKeyTime  string
// 	FieldKeyLevel string

// 	// TimestampFormat uses the default Go timestamp layout.
// 	// Default is time.RFC3339.
// 	TimestampFormat string

// 	// DisableTimestamp disables/enables timestamp in the output.
// 	// Default is false.
// 	DisableTimestamp bool

// 	configured bool
// }

// func (f *TextFormatter) Format(buf *Buffer, entry *Entry) error {
// 	f.setDefaults()

// 	if !f.DisableTimestamp {
// 		buf.AppendString(f.FieldKeyTime)
// 		buf.AppendString("=\"")
// 		buf.AppendString(entry.time.Format(f.TimestampFormat))
// 		buf.AppendString("\" ")
// 	}
// 	buf.AppendString(f.FieldKeyLevel)
// 	buf.AppendString("=\"")
// 	buf.AppendString(entry.level.String())
// 	buf.AppendString("\" ")
// 	buf.AppendString(f.FieldKeyMsg)
// 	buf.AppendString("=")
// 	if entry.format == "" {
// 		EscapeString(buf, fmt.Sprint(entry.args...))
// 	} else {
// 		EscapeString(buf, fmt.Sprintf(entry.format, entry.args...))
// 	}

// 	// var fields []Field
// 	// for l := entry.fields; l != nil; fields, l = l.Fields() {
// 	// 	for _, field := range fields {
// 	// 		buf.AppendString(" ")
// 	// 		buf.AppendString(field.Key)
// 	// 		buf.AppendString("=")

// 	// 		if !KnownTypeToBuf(buf, field.Value) {
// 	// 			buf.AppendString(fmt.Sprint(field.Value))
// 	// 		}
// 	// 	}
// 	// }

// 	buf.AppendString("\n")
// 	return nil
// }

// func (f *TextFormatter) setDefaults() {
// 	if f.configured {
// 		return
// 	}
// 	f.configured = true

// 	if f.FieldKeyMsg == "" {
// 		f.FieldKeyMsg = defaultFieldKeyMsg
// 	}
// 	if f.FieldKeyTime == "" {
// 		f.FieldKeyTime = defaultFieldKeyTime
// 	}
// 	if f.FieldKeyLevel == "" {
// 		f.FieldKeyLevel = defaultFieldKeyLevel
// 	}

// 	if f.TimestampFormat == "" {
// 		f.TimestampFormat = time.RFC3339
// 	}
// }
