package logf

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

// Enabled returns true if the given level is allowed within the current level.
func (l Level) Enabled(o Level) bool {
	return l >= o
}

// String implements fmt.Stringer
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warning"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// LevelChecker abstracts level checking process.
type LevelChecker func(Level) bool

type LevelEncoder func(Level, TypeEncoder)

func DefaultLevelEncoder(lvl Level, m TypeEncoder) {
	m.EncodeTypeString(lvl.String())
}
