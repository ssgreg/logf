package benchmarks

import (
	"io/ioutil"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
