package logf

import (
	"time"
)

// TextEncoderConfig allows to configure text Encoder.
type TextEncoderConfig struct {
	DisableFieldName   bool
	DisableFieldCaller bool

	EncodeTime     TimeEncoder
	EncodeDuration DurationEncoder
	EncodeError    ErrorEncoder
	EncodeCaller   CallerEncoder
}

// WithDefaults returns the new config in which all uninitialized fields are
// filled with their default values.
func (c TextEncoderConfig) WithDefaults() TextEncoderConfig {
	// Handle defaults for type encoder.
	if c.EncodeDuration == nil {
		c.EncodeDuration = StringDurationEncoder
	}
	if c.EncodeTime == nil {
		c.EncodeTime = LayoutTimeEncoder(time.StampMilli)
	}
	if c.EncodeError == nil {
		c.EncodeError = DefaultErrorEncoder
	}
	if c.EncodeCaller == nil {
		c.EncodeCaller = ShortCallerEncoder
	}

	return c
}
