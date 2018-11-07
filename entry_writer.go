package logf

// EntryWriter is the interface that should do real logging stuff.
type EntryWriter interface {
	WriteEntry(Entry)
}
