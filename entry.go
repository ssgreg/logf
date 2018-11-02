package logf

import "time"

type Entry struct {
	LoggerID      int32
	DerivedFields []Field
	Fields        []Field
	Level         Level
	Time          time.Time
	Text          string
	LoggerName    string
	Caller        EntryCaller
}

func newErrorEntry(text string, fs ...Field) Entry {
	return Entry{
		LoggerID: -1,
		Level:    LevelError,
		Time:     time.Now(),
		Text:     text,
		Fields:   fs,
	}
}
