package logf

import "fmt"

type ErrorEncoder func(string, error, FieldEncoder)

func DefaultErrorEncoder(k string, e error, m FieldEncoder) {
	msg := e.Error()
	m.EncodeFieldString(k, msg)

	switch e.(type) {
	case fmt.Formatter:
		verbose := fmt.Sprintf("%+v", e)
		if verbose != msg {
			m.EncodeFieldString(k+".verbose", verbose)
		}
	}
}
