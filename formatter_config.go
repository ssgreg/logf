package logf

const (
	DefaultFieldKeyLevel  = "level"
	DefaultFieldKeyMsg    = "msg"
	DefaultFieldKeyTime   = "ts"
	DefaultFieldKeyName   = "logger"
	DefaultFieldKeyCaller = "caller"
)

func SetFormatterConfigDefaults(c *FormatterConfig) *FormatterConfig {
	// Handle default for predefined field names.
	if c.FieldKeyLevel == "" {
		c.FieldKeyLevel = DefaultFieldKeyLevel
	}
	if c.FieldKeyMsg == "" {
		c.FieldKeyMsg = DefaultFieldKeyMsg
	}
	if c.FieldKeyTime == "" {
		c.FieldKeyTime = DefaultFieldKeyTime
	}
	if c.FieldKeyName == "" {
		c.FieldKeyName = DefaultFieldKeyName
	}
	if c.FieldKeyCaller == "" {
		c.FieldKeyCaller = DefaultFieldKeyCaller
	}

	// Handle defaults for type encoder.
	if c.FormatDuration == nil {
		c.FormatDuration = StringDurationFormatter
	}
	if c.FormatTime == nil {
		c.FormatTime = RFC3339TimeFormatter
	}
	if c.FormatError == nil {
		c.FormatError = DefaultErrorFormatter
	}
	if c.FormatCaller == nil {
		c.FormatCaller = ShortCallerFormatter
	}

	return c
}

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
