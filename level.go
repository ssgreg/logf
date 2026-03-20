package logf

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
)

// Level represents the severity of a log message. Higher numeric values
// mean more verbose output — LevelDebug (3) lets everything through,
// while LevelError (0) only lets errors pass.
type Level int8

// Severity levels.
const (
	// LevelError logs errors only — the quietest setting.
	LevelError Level = iota
	// LevelWarn logs errors and warnings.
	LevelWarn
	// LevelInfo logs errors, warnings, and informational messages. This is
	// the typical production setting.
	LevelInfo
	// LevelDebug logs everything — all severity levels pass through.
	LevelDebug
)

// Enabled reports whether a message at level o would be logged under this
// level threshold. For example, LevelInfo.Enabled(LevelDebug) is false.
func (l Level) Enabled(o Level) bool {
	return l >= o
}

// String returns a lower-case string representation of the Level
// ("debug", "info", "warn", "error").
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

// UpperCaseString returns an upper-case string representation of the Level
// ("DEBUG", "INFO", "WARN", "ERROR").
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

// MarshalText marshals the Level to its lower-case text representation.
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText parses a level string (case-insensitive) and sets the Level.
// Returns an error for unrecognized values.
func (l *Level) UnmarshalText(text []byte) error {
	s := string(text)
	lvl, ok := LevelFromString(s)
	if !ok {
		return fmt.Errorf("invalid logging level %q", s)
	}

	*l = lvl

	return nil
}

// LevelFromString parses a level name (case-insensitive) and returns the
// corresponding Level. Returns false if the name is not recognized.
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

// LevelEncoder is a function that formats a Level into the log output via
// TypeEncoder. Swap it out to control how levels appear in your logs.
type LevelEncoder func(Level, TypeEncoder)

// DefaultLevelEncoder formats levels as lower-case strings ("debug",
// "info", "warn", "error"). This is the default for JSON output.
func DefaultLevelEncoder(lvl Level, m TypeEncoder) {
	m.EncodeTypeString(lvl.String())
}

// UpperCaseLevelEncoder formats levels as upper-case strings ("DEBUG",
// "INFO", "WARN", "ERROR").
func UpperCaseLevelEncoder(lvl Level, m TypeEncoder) {
	m.EncodeTypeString(lvl.UpperCaseString())
}

// ShortTextLevelEncoder formats levels as compact 3-character uppercase
// strings (DBG, INF, WRN, ERR). This is the default for text/console
// output where horizontal space is precious.
func ShortTextLevelEncoder(lvl Level, m TypeEncoder) {
	switch lvl {
	case LevelDebug:
		m.EncodeTypeString("DBG")
	case LevelInfo:
		m.EncodeTypeString("INF")
	case LevelWarn:
		m.EncodeTypeString("WRN")
	case LevelError:
		m.EncodeTypeString("ERR")
	default:
		m.EncodeTypeString("UNK")
	}
}

// NewMutableLevel creates a MutableLevel starting at the given level.
// Pass it where a Level is expected and call Set later to change the
// threshold at runtime without restarting.
func NewMutableLevel(l Level) *MutableLevel {
	return &MutableLevel{level: uint32(l)}
}

// MutableLevel is a concurrency-safe level that can be changed at runtime
// without rebuilding the Logger. Perfect for admin endpoints that toggle
// debug logging on a live system — just call Set and every subsequent
// log call picks up the new level atomically.
type MutableLevel struct {
	level uint32
}

// Enabled reports whether the given level is enabled at the current mutable
// level. Safe for concurrent use.
func (l *MutableLevel) Enabled(_ context.Context, lvl Level) bool {
	return l.Level().Enabled(lvl)
}

// Level returns the current logging level atomically.
func (l *MutableLevel) Level() Level {
	return (Level)(atomic.LoadUint32(&l.level))
}

// Set atomically switches the logging level. All subsequent log calls
// will use the new threshold.
func (l *MutableLevel) Set(o Level) {
	atomic.StoreUint32(&l.level, uint32(o))
}
