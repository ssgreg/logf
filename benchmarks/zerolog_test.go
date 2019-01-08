package benchmarks

import (
	"io/ioutil"

	"github.com/rs/zerolog"
)

func newZerolog() zerolog.Logger {
	return zerolog.New(ioutil.Discard).With().Timestamp().Logger()
}

func newDisabledZerolog() zerolog.Logger {
	return newZerolog().Level(zerolog.Disabled)
}

func fakeZerologFields(e *zerolog.Event) *zerolog.Event {
	return e.
		Int("int", tenInts[0]).
		Interface("ints", tenInts).
		Str("string", tenStrings[0]).
		Interface("strings", tenStrings).
		Time("fm", tenTimes[0]).
		// Interface("times", tenTimes).
		Interface("user1", oneUser).
		// Interface("user2", oneUser).
		// Interface("users", tenUsers).
		Err(errExample)
}

func fakeZerologContext(c zerolog.Context) zerolog.Context {
	return c.
		Int("int", tenInts[0]).
		Interface("ints", tenInts).
		Str("string", tenStrings[0]).
		Interface("strings", tenStrings).
		Time("tm", tenTimes[0]).
		// Interface("times", tenTimes).
		Interface("user1", oneUser).
		// Interface("user2", oneUser).
		// Interface("users", tenUsers).
		Err(errExample)
}
