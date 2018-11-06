package logf

import "fmt"

type ErrorEncoder func(string, error, FieldEncoder)

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
