package logf

import "fmt"

// ErrorEncoder is a function that writes an error into the log output.
// It receives the field key, the error, and a FieldEncoder so it can
// emit one or more fields (e.g., a short message plus a verbose stack).
type ErrorEncoder func(string, error, FieldEncoder)

// DefaultErrorEncoder encodes an error as one or two fields: the error
// message under the given key, and (if the error implements fmt.Formatter)
// a verbose field with the full "%+v" output (stack traces, etc.).
func DefaultErrorEncoder(key string, err error, enc FieldEncoder) {
	NewErrorEncoder.Default()(key, err, enc)
}

// ErrorEncoderConfig controls how errors are encoded — specifically the
// verbose field suffix and whether verbose output is included at all.
type ErrorEncoderConfig struct {
	VerboseFieldSuffix string
	NoVerboseField     bool
}

// WithDefaults returns a copy of the config with zero-value fields replaced
// by defaults (verbose suffix ".verbose").
func (c ErrorEncoderConfig) WithDefaults() ErrorEncoderConfig {
	if c.VerboseFieldSuffix == "" {
		c.VerboseFieldSuffix = ".verbose"
	}

	return c
}

// NewErrorEncoder creates an ErrorEncoder with the given config. Call it
// as a function: NewErrorEncoder(cfg) returns an ErrorEncoder.
var NewErrorEncoder = errorEncoderGetter(
	func(c ErrorEncoderConfig) ErrorEncoder {
		return func(key string, err error, enc FieldEncoder) {
			encodeError(key, err, enc, c.WithDefaults())
		}
	},
)

type errorEncoderGetter func(c ErrorEncoderConfig) ErrorEncoder

var defaultErrorEncoder ErrorEncoder

func (c errorEncoderGetter) Default() ErrorEncoder {
	if defaultErrorEncoder == nil {
		defaultErrorEncoder = c(ErrorEncoderConfig{})
	}
	return defaultErrorEncoder
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
