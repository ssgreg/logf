package logf

import "sync/atomic"

// NewMutableLevel creates an instance of MutableLevel with the given
// starting level.
func NewMutableLevel(l Level) *MutableLevel {
	return &MutableLevel{level: uint32(l)}
}

// MutableLevel allows to switch the logging level atomically.
//
// The logger does not allow to change logging level in runtime by itself.
type MutableLevel struct {
	level uint32
}

// Checker is common way to get LevelChecker. Use it with every custom
// implementation of Level.
func (l *MutableLevel) Checker() LevelChecker {
	return func(o Level) bool {
		return l.Level().Enabled(o)
	}
}

// Level returns the current logging level.
func (l *MutableLevel) Level() Level {
	return (Level)(atomic.LoadUint32(&l.level))
}

// Set switches the current logging level to the given one.
func (l *MutableLevel) Set(o Level) {
	atomic.StoreUint32(&l.level, uint32(o))
}
