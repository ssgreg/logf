package benchmarks

import (
	"io"

	"github.com/ssgreg/logf/v2"
)

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
	w, close := logf.NewChannelWriter(l, logf.ChannelWriterConfig{
		Appender: logf.NewWriteAppender(io.Discard, encoder),
	})

	return logf.NewLogger(w), close
}
