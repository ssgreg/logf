package benchmarks

import (
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ssgreg/logf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var disableOthers = true

func BenchmarkDisabledPlainText(b *testing.B) {
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
				m.Write()
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

func BenchmarkDisabledPlainTextWithAccumulatedFields(b *testing.B) {
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
				m.Write()
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

func BenchmarkDisabledTextWithFields(b *testing.B) {
	b.Run("logf", func(b *testing.B) {
		logger := logf.NewDisabledLogger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info(getMessage(0), fakeFields()...)
		}
	})
	b.Run("logf.check", func(b *testing.B) {
		logger := logf.NewDisabledLogger()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.AtLevel(logf.LevelInfo, func(log logf.LogFunc) {
				logger.Info(getMessage(0), fakeFields()...)
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
			logger.Info(getMessage(0), fakeZapFields()...)
		}
	})
	b.Run("uber/zap.check", func(b *testing.B) {
		logger := newZapLogger(zap.ErrorLevel)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if m := logger.Check(zap.InfoLevel, getMessage(0)); m != nil {
				m.Write(fakeZapFields()...)
			}
		}
	})
	b.Run("rs/zerolog", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fakeZerologFields(logger.Info()).Msg(getMessage(0))
		}
	})
	b.Run("rs/zerolog.check", func(b *testing.B) {
		logger := newDisabledZerolog()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if e := logger.Info(); e.Enabled() {
				fakeZerologFields(logger.Info()).Msg(getMessage(0))
			}
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

func BenchmarkAccumulateFields(b *testing.B) {
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

func BenchmarkAddName(b *testing.B) {
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

func BenchmarkPlainTextWithCaller(b *testing.B) {
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

func BenchmarkAccumulateFieldsWithAccumulatedFields(b *testing.B) {
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

func BenchmarkPlainText(b *testing.B) {
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

func BenchmarkPlainTextWithAccumulatedFields(b *testing.B) {
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

func BenchmarkTextWithFields(b *testing.B) {
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

func newLogger(l logf.Level) (*logf.Logger, logf.ChannelWriterCloseFunc) {
	encoder := logf.NewJSONEncoder.Default()
	w, close := logf.NewChannelWriter(logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(ioutil.Discard, encoder),
	})

	return logf.NewLogger(logf.NewMutableLevel(l), w), close
}
