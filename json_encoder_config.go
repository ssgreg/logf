package logf

// Default field keys.
const (
	DefaultFieldKeyLevel  = "level"
	DefaultFieldKeyMsg    = "msg"
	DefaultFieldKeyTime   = "ts"
	DefaultFieldKeyName   = "logger"
	DefaultFieldKeyCaller = "caller"
)

// JSONEncoderConfig allows to configure journal JSON Encoder.
type JSONEncoderConfig struct {
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

	EncodeTime     TimeEncoder
	EncodeDuration DurationEncoder
	EncodeError    ErrorEncoder
	EncodeLevel    LevelEncoder
	EncodeCaller   CallerEncoder
}

// WithDefaults returns the new config in which all uninitialized fields are
// filled with their default values.
func (c JSONEncoderConfig) WithDefaults() JSONEncoderConfig {
	// Handle default for predefined field names.
	if c.FieldKeyMsg == "" {
		c.FieldKeyMsg = DefaultFieldKeyMsg
	}
	if c.FieldKeyTime == "" {
		c.FieldKeyTime = DefaultFieldKeyTime
	}
	if c.FieldKeyLevel == "" {
		c.FieldKeyLevel = DefaultFieldKeyLevel
	}
	if c.FieldKeyName == "" {
		c.FieldKeyName = DefaultFieldKeyName
	}
	if c.FieldKeyCaller == "" {
		c.FieldKeyCaller = DefaultFieldKeyCaller
	}

	// Handle defaults for type encoder.
	if c.EncodeDuration == nil {
		c.EncodeDuration = StringDurationEncoder
	}
	if c.EncodeTime == nil {
		c.EncodeTime = RFC3339TimeEncoder
	}
	if c.EncodeError == nil {
		c.EncodeError = DefaultErrorEncoder
	}
	if c.EncodeLevel == nil {
		c.EncodeLevel = DefaultLevelEncoder
	}
	if c.EncodeCaller == nil {
		c.EncodeCaller = ShortCallerEncoder
	}

	return c
}
