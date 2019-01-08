package benchmarks

import (
	"io/ioutil"

	"github.com/ssgreg/logrus"
)

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
