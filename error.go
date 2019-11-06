package logf

import "fmt"

// ErrorEncoder is the function type to encode the given error.
type ErrorEncoder func(string, error, FieldEncoder)

// DefaultErrorEncoder encodes the given error as a set of fields.
//
// A mandatory field with the given key and an optional field with the
// full verbose error message.
func DefaultErrorEncoder(key string, err error, enc FieldEncoder) {
	NewErrorEncoder.Default()(key, err, enc)
}

// ErrorEncoderConfig allows to configure ErrorEncoder.
type ErrorEncoderConfig struct {
	VerboseFieldSuffix string
	NoVerboseField     bool
}

// WithDefaults returns the new config in which all uninitialized fields are
// filled with their default values.
func (c ErrorEncoderConfig) WithDefaults() ErrorEncoderConfig {
	if c.VerboseFieldSuffix == "" {
		c.VerboseFieldSuffix = ".verbose"
	}

	return c
}

// NewErrorEncoder creates the new instance of the ErrorEncoder with the
// given ErrorEncoderConfig.
var NewErrorEncoder = errorEncoderGetter(
	func(c ErrorEncoderConfig) ErrorEncoder {
		return func(key string, err error, enc FieldEncoder) {
			encodeError(key, err, enc, c.WithDefaults())
		}
	},
)

type errorEncoderGetter func(c ErrorEncoderConfig) ErrorEncoder

func (c errorEncoderGetter) Default() ErrorEncoder {
	return c(ErrorEncoderConfig{})
}

// encodeError encodes the given error as a set of fields.
//
// A mandatory field with the given key and an optional field with the
// full verbose error message according to the given config.
func encodeError(key string, err error, enc FieldEncoder, c ErrorEncoderConfig) {
	var msg string
	if err == nil {
		msg = "<nil>"
	} else {
		msg = err.Error()
	}
	enc.EncodeFieldString(key, msg)

	if !c.NoVerboseField {
		switch err.(type) {
		case fmt.Formatter:
			verbose := fmt.Sprintf("%+v", err)
			if verbose != msg {
				enc.EncodeFieldString(key+c.VerboseFieldSuffix, verbose)
			}
		}
	}
}
