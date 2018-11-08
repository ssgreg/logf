package logf

import (
	"os"
)

// EscapeCode represents a single code in the ANSI escape sequences.
//
// These sequences define functions that change display graphics, control
// cursor movement, and reassign keys.
//
// ANSI escape sequence is a sequence of ASCII characters, the first two
// of which are the ASCII "Escape" character 27 (1Bh) and the left-bracket
// character " [ " (5Bh). The character or characters following the escape
// and left-bracket characters specify an alphanumeric code that controls
// a keyboard or display function.
type EscapeCode int8

// checkNoColor checks for NO_COLORS environment variable to disable color
// output.
//
// All command-line software which outputs text with ANSI color added should
// check for the presence of a NO_COLOR environment variable that, when
// present (regardless of its value), prevents the addition of ANSI color.
func checkNoColor() bool {
	_, ok := os.LookupEnv("NO_COLOR")

	return ok
}

// Text colors.
const (
	EscReset EscapeCode = iota + 0
	EscBold
	EscFaint
	EscItalic
	EscUnderline
	EscSlowBlink
	EscRapidBlink
	EscReverse
	EscConseal
	EscCrossedOut
)

// Text colors.
const (
	EscBlack EscapeCode = iota + 30
	EscRed
	EscGreen
	EscYellow
	EscBlue
	EscMagenta
	EscCyan
	EscWhite
)

// Background colors.
const (
	EscBgBlack EscapeCode = iota + 40
	EscBgRed
	EscBgGreen
	EscBgYellow
	EscBgBlue
	EscBgMagenta
	EscBgCyan
	EscBgWhite
)

// Bright text colors.
const (
	EscBrightBlack EscapeCode = iota + 90
	EscBrightRed
	EscBrightGreen
	EscBrightYellow
	EscBrightBlue
	EscBrightMagenta
	EscBrightCyan
	EscBrightWhite
)

// Bright background colors.
const (
	EscBrightBgBlack EscapeCode = iota + 100
	EscBrightBgRed
	EscBrightBgGreen
	EscBrightBgYellow
	EscBrightBgBlue
	EscBrightBgMagenta
	EscBrightBgCyan
	EscBrightBgWhite
)

// AtEscapeSequence calls the given fn, wrapped with the escape sequence,
// based on the given code.
func AtEscapeSequence(buf *Buffer, clr EscapeCode, fn func()) {
	if noColor {
		fn()
		return
	}

	buf.AppendString("\x1b[")
	AppendInt(buf, int64(clr))
	buf.AppendByte('m')
	fn()
	buf.AppendString("\x1b[0m")
}

// AtEscapeSequence2 calls the given fn, wrapped with the escape sequence,
// based on the given codes.
func AtEscapeSequence2(buf *Buffer, clr1, clr2 EscapeCode, fn func()) {
	if noColor {
		fn()
		return
	}

	buf.AppendString("\x1b[")
	AppendInt(buf, int64(clr1))
	buf.AppendByte(';')
	AppendInt(buf, int64(clr2))
	buf.AppendByte('m')
	fn()
	buf.AppendString("\x1b[0m")
}

// AtEscapeSequence3 calls the given fn, wrapped with the escape sequence,
// based on the given codes.
func AtEscapeSequence3(buf *Buffer, clr1, clr2, clr3 EscapeCode, fn func()) {
	if noColor {
		fn()
		return
	}

	buf.AppendString("\x1b[")
	AppendInt(buf, int64(clr1))
	buf.AppendByte(';')
	AppendInt(buf, int64(clr2))
	buf.AppendByte(';')
	AppendInt(buf, int64(clr2))
	buf.AppendByte('m')
	fn()
	buf.AppendString("\x1b[0m")
}

var (
	noColor = checkNoColor()
)
