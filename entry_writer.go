package logf

// EntryWriter is the interface that should do real logging stuff.
type EntryWriter interface {
	WriteEntry(Entry)
}

// NewUnbufferedEntryWriter returns an implementation of EntryWriter which
// puts entries directly to the appender immediately and synchronously.
func NewUnbufferedEntryWriter(appender Appender) EntryWriter {
	return unbufferedEntryWriter{appender}
}

type unbufferedEntryWriter struct {
	appender Appender
}

func (w unbufferedEntryWriter) WriteEntry(entry Entry) {
	w.appender.Append(entry)
}
