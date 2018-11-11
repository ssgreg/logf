package bench

import (
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/ssgreg/logf"
	"github.com/ssgreg/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var disableOthers = false

func BenchmarkDisabledWithoutFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, _ := newLogger(logf.LevelError)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("logf.check", func(b *testing.B) {
		logger, _ := newLogger(logf.LevelError)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.AtLevel(logf.LevelInfo, func(log logf.LogFunc) {
				log(getMessage(0))
			})
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("uber/zap.check", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if m := logger.Check(zap.InfoLevel, getMessage(0)); m != nil {
				logger.Info(getMessage(0))
			}
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().Msg(getMessage(0))
		}
	})
	b.Run("rs/zerolog.check", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if e := logger.Info(); e.Enabled() {
				logger.Info().Msg(getMessage(0))
			}
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newDisabledLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
}

func BenchmarkDisabledAccumulatedContext(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger := logf.NewDisabledLogger().With(fakeFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("logf.check", func(b *testing.B) {
		logger := logf.NewDisabledLogger().With(fakeFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.AtLevel(logf.LevelInfo, func(log logf.LogFunc) {
				log(getMessage(0))
			})
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).With(fakeZapFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("uber/zap.check", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).With(fakeZapFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if m := logger.Check(zap.InfoLevel, getMessage(0)); m != nil {
				logger.Info(getMessage(0))
			}
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := fakeZerologContext(newDisabledZerolog().With()).Logger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().Msg(getMessage(0))
		}
	})
	b.Run("rs/zerolog.check", func(b *testing.B) {
		logger := fakeZerologContext(newDisabledZerolog().With()).Logger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if e := logger.Info(); e.Enabled() {
				logger.Info().Msg(getMessage(0))
			}
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newDisabledLogrus().WithFields(fakeLogrusFields())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
}

func BenchmarkDisabledAddingFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger := logf.NewDisabledLogger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0), fakeFields()...)
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0), fakeZapFields()...)
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fakeZerologFields(logger.Info()).Msg(getMessage(0))
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newDisabledLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithFields(fakeLogrusFields()).Info(getMessage(0))
		}
	})
}

func BenchmarkCreateContextLogger(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.With(fakeFields()...)
			_ = l
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.With(fakeZapFields()...)
			_ = l
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := fakeZerologContext(logger.With()).Logger()
			_ = l
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.WithFields(fakeLogrusFields())
			_ = l
		}
	})
}

func BenchmarkWithName(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.WithName("test")
			_ = l
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.Named("test")
			_ = l
		}
	})
}

func BenchmarkWithCaller(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()
		logger = logger.WithCaller()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).WithOptions(zap.AddCaller())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
}

func BenchmarkCreateContextWithAccumulatedContextLogger(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()
		logger = logger.With(fakeFields()...)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.With(fakeFields()...)
			_ = l
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel).With(fakeZapFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.With(fakeZapFields()...)
			_ = l
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := fakeZerologContext(newZerolog().With()).Logger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := fakeZerologContext(logger.With()).Logger()
			_ = l
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus().WithFields(fakeLogrusFields())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := logger.WithFields(fakeLogrusFields())
			_ = l
		}
	})
}

func BenchmarkWithoutFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().Msg(getMessage(0))
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
}

func BenchmarkAccumulatedContext(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()
		logger = logger.With(fakeFields()...)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel).With(fakeZapFields()...)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := fakeZerologContext(newZerolog().With()).Logger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().Msg(getMessage(0))
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus().WithFields(fakeLogrusFields())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0))
		}
	})
}

func BenchmarkAddingFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger, close := newLogger(logf.LevelDebug)
		defer close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0), fakeFields()...)
		}
	})
	if disableOthers == true {
		return
	}

	b.Run("uber/zap", func(b *testing.B) {
		logger := newZapLogger(zap.DebugLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0), fakeZapFields()...)
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fakeZerologFields(logger.Info()).Msg(getMessage(0))
		}
	})
	b.Run("sirupsen/logrus", func(b *testing.B) {
		logger := newLogrus()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithFields(fakeLogrusFields()).Info(getMessage(0))
		}
	})
}

type user struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
type users []*user

func (u *user) EncodeLogfObject(enc logf.FieldEncoder) error {
	enc.EncodeFieldString("name", u.Name)
	enc.EncodeFieldString("email", u.Email)
	enc.EncodeFieldInt64("createdAt", u.CreatedAt.UnixNano())
	// enc.EncodeFieldTime("createdAt", u.CreatedAt)

	return nil
}

// func (uu users) MarshalLogArray(arr zapcore.ArrayEncoder) error {
// 	var err error
// 	for i := range uu {
// 		err = multierr.Append(err, arr.AppendObject(uu[i]))
// 	}
// 	return err
// }

func (u *user) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", u.Name)
	enc.AddString("email", u.Email)
	enc.AddInt64("createdAt", u.CreatedAt.UnixNano())
	return nil
}

var (
	errExample = errors.New("example")

	messages = makePseudoMessages(1000)

	tenInts    = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	tenStrings = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	tenTimes = []time.Time{
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

func fakeFields() []logf.Field {
	return []logf.Field{
		logf.Int("int", tenInts[0]),
		logf.ConstInts("ints", tenInts),
		logf.String("string", tenStrings[0]),
		logf.Strings("strings", tenStrings),
		logf.Time("tm", tenTimes[0]),
		// logf.Duration("dur", time.Second),
		// logf.Durations("durs", []time.Duration{time.Second, time.Millisecond}),
		// // logf.Any("times", tenTimes),
		logf.Object("user1", oneUser),
		// // logf.Any("user2", oneUser),
		// // logf.Any("users", tenUsers),
		logf.NamedError("error", errExample),
	}
}

func fakeZapFields() []zapcore.Field {
	return []zap.Field{
		zap.Int("int", tenInts[0]),
		zap.Ints("ints", tenInts),
		zap.String("string", tenStrings[0]),
		zap.Strings("strings", tenStrings),
		zap.Time("fm", tenTimes[0]),
		// zap.Times("times", tenTimes),
		zap.Object("user1", oneUser),
		// zap.Any("user2", oneUser),
		// zap.Any("users", tenUsers),
		zap.Error(errExample),
	}
}

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
		// TODO: use zero log object marshaller
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

// A Syncer is a spy for the Sync portion of zapcore.WriteSyncer.
type Syncer struct {
	err    error
	called bool
}

// SetError sets the error that the Sync method will return.
func (s *Syncer) SetError(err error) {
	s.err = err
}

// Sync records that it was called, then returns the user-supplied error (if
// any).
func (s *Syncer) Sync() error {
	s.called = true
	return s.err
}

// Called reports whether the Sync method was called.
func (s *Syncer) Called() bool {
	return s.called
}

// A Discarder sends all writes to ioutil.Discard.
type Discarder struct{ Syncer }

// Write implements io.Writer.
func (d *Discarder) Write(b []byte) (int, error) {
	return ioutil.Discard.Write(b)
	// fmt.Println(string(b))
	// panic(33)
	// return 0, nil
}

func newZapLogger(lvl zapcore.Level) *zap.Logger {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeDuration = zapcore.NanosDurationEncoder
	// ec.EncodeTime = zapcore.EpochNanosTimeEncoder
	ec.EncodeTime = zapcore.ISO8601TimeEncoder
	enc := zapcore.NewJSONEncoder(ec)
	return zap.New(zapcore.NewCore(
		enc,
		&Discarder{},
		lvl,
	))
}

func newLogger(l logf.Level) (*logf.Logger, logf.ChannelWriterCloseFunc) {
	encoder := logf.NewJSONEncoder.Default()
	// encoder := logf.NewJSONEncoder(logf.JSONEncoderConfig{
	// 	EncodeTime: logf.LayoutTimeEncoder(time.RFC3339),
	// })
	// encoder := logf.NewTextEncoder.Default()
	// encoder := logfjournald.NewEncoder.Default()

	w, close := logf.NewChannelWriter(logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(ioutil.Discard, encoder),
		// Appender: logf.NewDiscardAppender(),
	})

	return logf.NewLogger(logf.NewMutableLevel(l), w), close
}

func newDisabledLogrus() *logrus.Logger {
	logger := newLogrus()
	logger.Level = logrus.ErrorLevel
	return logger
}

func newLogrus() *logrus.Logger {
	return &logrus.Logger{
		Out:       ioutil.Discard,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
}

func fakeLogrusFields() logrus.Fields {
	return logrus.Fields{
		"int":     tenInts[0],
		"ints":    tenInts,
		"string":  tenStrings[0],
		"strings": tenStrings,
		"tm":      tenTimes[0],
		// "times":   tenTimes,
		"user1": oneUser,
		// "user2":   oneUser,
		// "users":   tenUsers,
		"error": errExample,
	}
}
