package logf

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
)

// Level defines severity level of a log message.
type Level int8

// Severity levels.
const (
	// LevelError allows to log errors only.
	LevelError Level = iota
	// LevelWarn allows to log errors and warnings.
	LevelWarn
	// LevelInfo is the default logging level. Allows to log errors, warnings and infos.
	LevelInfo
	// LevelDebug allows to log messages with all severity levels.
	LevelDebug
)

// Enabled returns true if the given level is allowed within the current level.
func (l Level) Enabled(o Level) bool {
	return l >= o
}

// String implements fmt.Stringer.
// String returns a lower-case string representation of the Level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// UpperCaseString returns an upper-case string representation of the Level.
func (l Level) UpperCaseString() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// MarshalText marshals the Level to text.
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText unmarshals the Level from text.
func (l *Level) UnmarshalText(text []byte) error {
	s := string(text)
	lvl, ok := LevelFromString(s)
	if !ok {
		return fmt.Errorf("invalid logging level %q", s)
	}

	*l = lvl

	return nil
}

// LevelFromString creates the new Level with the given string.
func LevelFromString(lvl string) (Level, bool) {
	switch strings.ToLower(lvl) {
	case "debug":
		return LevelDebug, true
	case "info", "information":
		return LevelInfo, true
	case "warn", "warning":
		return LevelWarn, true
	case "error":
		return LevelError, true
	}

	return LevelError, false
}

// LevelEncoder is the function type to encode Level.
type LevelEncoder func(Level, TypeEncoder)

// DefaultLevelEncoder implements LevelEncoder by calling Level itself.
func DefaultLevelEncoder(lvl Level, m TypeEncoder) {
	m.EncodeTypeString(lvl.String())
}

// UpperCaseLevelEncoder implements LevelEncoder by calling Level itself.
func UpperCaseLevelEncoder(lvl Level, m TypeEncoder) {
	m.EncodeTypeString(lvl.UpperCaseString())
}

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

// Enabled reports whether the given level is enabled at the current
// mutable level.
func (l *MutableLevel) Enabled(_ context.Context, lvl Level) bool {
	return l.Level().Enabled(lvl)
}

// Level returns the current logging level.
func (l *MutableLevel) Level() Level {
	return (Level)(atomic.LoadUint32(&l.level))
}

// Set switches the current logging level to the given one.
func (l *MutableLevel) Set(o Level) {
	atomic.StoreUint32(&l.level, uint32(o))
}
