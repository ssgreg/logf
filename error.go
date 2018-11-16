package logf

import "fmt"

// ErrorEncoder is the function type to encode the given error.
type ErrorEncoder func(string, error, FieldEncoder)

// DefaultErrorEncoder encodes the given error as a set of fields.
//
// A mandatory field with the given key and an optional field with the
// full verbose error message.
func DefaultErrorEncoder(k string, e error, m FieldEncoder) {
	var msg string
	if e == nil {
		msg = "<nil>"
	} else {
		msg = e.Error()
	}
	m.EncodeFieldString(k, msg)

	switch e.(type) {
	case fmt.Formatter:
		verbose := fmt.Sprintf("%+v", e)
		if verbose != msg {
			m.EncodeFieldString(k+".verbose", verbose)
		}
	}
}
