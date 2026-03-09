package logf

import (
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

// CallerPC captures the program counter of the caller, skipping the
// given number of frames. Returns 0 if the caller cannot be determined.
func CallerPC(skip int) uintptr {
	var pcs [1]uintptr
	if runtime.Callers(skip+2, pcs[:]) < 1 {
		return 0
	}

	return pcs[0]
}

// callerFrame resolves a program counter to file and line.
func callerFrame(pc uintptr) (file string, line int) {
	frames := runtime.CallersFrames([]uintptr{pc})
	f, _ := frames.Next()

	return f.File, f.Line
}

// fileWithPackage cuts a package name and a file name from a full file path.
//
// As for os-specific path separator battle here, my opinion coincides
// with the last comment here https://github.com/golang/go/issues/3335.
//
// Go team should simply document the current behavior of always using
// '/' in stack frame data. That's the way it's been implemented for
// years, and packages like github.com/go-stack/stack that have been
// stable for years expect it. Changing the behavior in a future version
// of Go will break working code without a clearly documented benefit.
// Documenting the behavior will help avoid new code from making the
// wrong assumptions.
func fileWithPackage(file string) string {
	found := strings.LastIndexByte(file, '/')
	if found == -1 {
		return file
	}
	found = strings.LastIndexByte(file[:found], '/')
	if found == -1 {
		return file
	}

	return file[found+1:]
}

// CallerEncoder is the function type to encode a caller program counter.
type CallerEncoder func(pc uintptr, m TypeEncoder)

// ShortCallerEncoder resolves the given PC and encodes it as package/file:line.
func ShortCallerEncoder(pc uintptr, m TypeEncoder) {
	file, line := callerFrame(pc)
	var callerBuf [64]byte
	var b []byte
	b = callerBuf[:0]
	b = append(b, fileWithPackage(file)...)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(line), 10)

	m.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}

// FullCallerEncoder resolves the given PC and encodes it as full/path/file:line.
func FullCallerEncoder(pc uintptr, m TypeEncoder) {
	file, line := callerFrame(pc)
	var callerBuf [256]byte
	var b []byte
	b = callerBuf[:0]
	b = append(b, file...)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(line), 10)

	m.EncodeTypeUnsafeBytes(noescape(unsafe.Pointer(&b)))
	runtime.KeepAlive(&b)
}
