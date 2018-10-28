package logf

type FormatterConfig struct {
	FieldKeyMsg    string
	FieldKeyTime   string
	FieldKeyLevel  string
	FieldKeyName   string
	FieldKeyCaller string

	FormatTime     TimeFormatter
	FormatDuration DurationFormatter
	FormatError    ErrorFormatter
	// MarshalLevel
	// MarshalName
	FormatCaller CallerFormatter
}
