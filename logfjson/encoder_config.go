package logfjson

import "github.com/ssgreg/logf"

const (
	DefaultFieldKeyLevel  = "level"
	DefaultFieldKeyMsg    = "msg"
	DefaultFieldKeyTime   = "ts"
	DefaultFieldKeyName   = "logger"
	DefaultFieldKeyCaller = "caller"
)

type EncoderConfig struct {
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

	EncodeTime     logf.TimeEncoder
	EncodeDuration logf.DurationEncoder
	EncodeError    logf.ErrorEncoder
	EncodeLevel    logf.LevelEncoder
	EncodeCaller   logf.CallerEncoder
}

func (c EncoderConfig) WithDefaults() EncoderConfig {
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
		c.EncodeDuration = logf.StringDurationEncoder
	}
	if c.EncodeTime == nil {
		c.EncodeTime = logf.RFC3339TimeEncoder
	}
	if c.EncodeError == nil {
		c.EncodeError = logf.DefaultErrorEncoder
	}
	if c.EncodeLevel == nil {
		c.EncodeLevel = logf.DefaultLevelEncoder
	}
	if c.EncodeCaller == nil {
		c.EncodeCaller = logf.ShortCallerEncoder
	}

	return c
}
