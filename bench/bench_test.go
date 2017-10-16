package bench

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ssgreg/logf"
)

func BenchmarkDisabledWithoutFields(b *testing.B) {
	b.Run("Logf-disabled", func(b *testing.B) {
		logger := logf.NewDisabled()
		defer logger.Close()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(getMessage(0))
			}
		})
	})
	b.Run("Logf-below-level", func(b *testing.B) {
		logger := newDiscardLogger(logf.ErrorLevel, 0)
		defer logger.Close()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(getMessage(0))
			}
		})
	})
}

func BenchmarkDisabledAccumulatedContext(b *testing.B) {
	b.Run("Logf-disabled", func(b *testing.B) {
		logger := fakeFields(logf.NewDisabled()).Logger()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(getMessage(0))
			}
		})
	})
	b.Run("Logf-below-level", func(b *testing.B) {
		logger := fakeFields(newDiscardLogger(logf.ErrorLevel, 0)).Logger()
		defer logger.Close()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(getMessage(0))
			}
		})
	})
}

func BenchmarkDisabledAddingFields(b *testing.B) {
	b.Run("Logf-disabled", func(b *testing.B) {
		logger := logf.NewDisabled()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				fakeFields(logger.Info()).Msg(getMessage(0))
			}
		})
	})
	b.Run("Logf-below-level", func(b *testing.B) {
		logger := newDiscardLogger(logf.ErrorLevel, 0)
		defer logger.Close()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				fakeFields(logger.Info()).Msg(getMessage(0))
			}
		})
	})
}

func BenchmarkWithoutFields(b *testing.B) {
	b.Run("Logf", func(b *testing.B) {
		logger := newDiscardLogger(logf.DebugLevel, 1000)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(getMessage(0))
			}
		})
		logger.Close()
	})
	b.Run("Logf-formatting", func(b *testing.B) {
		logger := newDiscardLogger(logf.DebugLevel, 1000)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msgf("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
		logger.Close()
	})
}

func BenchmarkAccumulatedContext(b *testing.B) {
	b.Run("Logf", func(b *testing.B) {
		logger := fakeFields(newDiscardLogger(logf.DebugLevel, 1000)).Logger()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msg(getMessage(0))
			}
		})
		logger.Close()
	})
	b.Run("Logf-formatting", func(b *testing.B) {
		logger := fakeFields(newDiscardLogger(logf.DebugLevel, 1000)).Logger()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info().Msgf("%v %v %v %s %v %v %v %v %v %s\n", fakeFmtArgs()...)
			}
		})
		logger.Close()
	})
}

func BenchmarkAddingFields(b *testing.B) {
	b.Run("Logf", func(b *testing.B) {
		logger := newDiscardLogger(logf.DebugLevel, 1000)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				fakeFields(logger.Info()).Msg(getMessage(0))
			}
		})
	})
}

type user struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
type users []*user

var (
	errExample = errors.New("example")

	messages   = makePseudoMessages(1000)
	tenInts    = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	tenStrings = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	tenTimes   = []time.Time{
		time.Unix(0, 0), time.Unix(1, 0), time.Unix(2, 0), time.Unix(3, 0), time.Unix(4, 0),
		time.Unix(5, 0), time.Unix(6, 0), time.Unix(7, 0), time.Unix(8, 0), time.Unix(9, 0),
	}
	oneUser = &user{
		Name:      "Grigory Zubankov",
		Email:     "greg@acronis.com",
		CreatedAt: time.Date(1980, 3, 2, 12, 0, 0, 0, time.UTC),
	}
	tenUsers = users{oneUser, oneUser, oneUser, oneUser, oneUser, oneUser, oneUser, oneUser, oneUser, oneUser}
)

func makePseudoMessages(n int) []string {
	messages := make([]string, n)
	for i := range messages {
		messages[i] = fmt.Sprintf("A text that pretend to be a real message in case of length %d", i)
	}
	return messages
}

func getMessage(n int) string {
	return messages[n%len(messages)]
}

func fakeFields(l logf.FieldLogger) logf.FieldLogger {
	return l.
		WithInt("int", tenInts[0]).
		WithAny("ints", tenInts).
		WithStr("string", tenStrings[0]).
		WithAny("strings", tenStrings).
		WithTime("time", tenTimes[0]).
		WithAny("times", tenTimes).
		WithAny("user1", oneUser).
		WithAny("user2", oneUser).
		WithAny("users", tenUsers).
		WithErr(errExample)
}

func fakeFmtArgs() []interface{} {
	// Need to keep this a function instead of a package-global var so that we
	// pay the cast-to-interface{} penalty on each call.
	return []interface{}{
		tenInts[0],
		tenInts,
		tenStrings[0],
		tenStrings,
		tenTimes[0],
		tenTimes,
		oneUser,
		oneUser,
		tenUsers,
		errExample,
	}
}

func (u user) TakeSnapshot() interface{} {
	return u
}

func newDiscardLogger(level logf.Level, capacity int) logf.FieldLogger {
	return logf.New(logf.LoggerParams{
		Level:    level,
		Capacity: capacity,
		Appender: &logf.DiscardAppender{},
	})
}

func newLogger(level logf.Level, capacity int) logf.FieldLogger {
	return logf.New(logf.LoggerParams{
		Level:    level,
		Capacity: capacity,
		Appender: logf.NewFileAppender("dat",
			&logf.JSONFormatter{
				TimestampFormat: time.RFC3339Nano,
			}),
	})
}
