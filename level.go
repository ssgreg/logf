package logf

import (
	"strings"
)

// Level defines severity level of a log message.
type Level uint32

// Severity levels.
const (
	// LevelError logs errors only.
	LevelError Level = iota
	// LevelWarning logs errors and warnings.
	LevelWarn
	// LevelInfo is the default logging level. Logs errors, warnings and infos.
	LevelInfo
	// LevelDebug logs everything.
	LevelDebug
)

// Checker is common way to get LevelChecker. Use it with every custom
// implementation of Level.
func (l Level) Checker() LevelChecker {
	return func(o Level) bool {
		return l.Enabled(o)
	}
}

// LevelChecker implements LevelCheckerGetter.
func (l Level) LevelChecker() LevelChecker {
	return l.Checker()
}

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

// LevelChecker abstracts level checking process.
type LevelChecker func(Level) bool

// LevelCheckerGetter allows the implementor to act like a common Level
// checker for the Logger.
type LevelCheckerGetter interface {
	LevelChecker() LevelChecker
}

// LevelCheckerGetterFunc defines a function that returns LevelChecker.
type LevelCheckerGetterFunc func() LevelChecker

// LevelChecker implements LevelCheckerGetter interface.
func (fn LevelCheckerGetterFunc) LevelChecker() LevelChecker {
	return fn()
}

// LevelEncoder is the function to encode Level.
type LevelEncoder func(Level, TypeEncoder)

// DefaultLevelEncoder implements LevelEncoder by calling Level itself.
func DefaultLevelEncoder(lvl Level, m TypeEncoder) {
	m.EncodeTypeString(lvl.String())
}

// UpperCaseLevelEncoder implements LevelEncoder by calling Level itself.
func UpperCaseLevelEncoder(lvl Level, m TypeEncoder) {
	m.EncodeTypeString(lvl.UpperCaseString())
}
