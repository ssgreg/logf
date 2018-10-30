package logf

const (
	DefaultFieldKeyLevel  = "level"
	DefaultFieldKeyMsg    = "msg"
	DefaultFieldKeyTime   = "ts"
	DefaultFieldKeyName   = "logger"
	DefaultFieldKeyCaller = "caller"
)

type FormatterConfig struct {
	FieldKeyMsg    string
	FieldKeyTime   string
	FieldKeyLevel  string
	FieldKeyName   string
	FieldKeyCaller string

	DisableFieldMsg    bool
	DisableFieldTime   bool
	DisableFieldLevel  bool
	DisableFieldName   bool
	DisableFieldCaller bool

	FormatTime     TimeFormatter
	FormatDuration DurationFormatter
	FormatError    ErrorFormatter
	// MarshalLevel
	// MarshalName
	FormatCaller CallerFormatter
}
