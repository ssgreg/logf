package logf

import (
	"runtime"
	"strings"
)

// EntryCaller holds values returned by runtime.Caller.
type EntryCaller struct {
	PC        uintptr
	File      string
	Line      int
	Specified bool
}

// NewEntryCaller creates an instance of EntryCaller with the given number
// of frames to skip.
func NewEntryCaller(skip int) EntryCaller {
	var c EntryCaller
	c.PC, c.File, c.Line, c.Specified = runtime.Caller(skip + 1)

	return c
}

// FileWithPackage cuts a package name and a file name from EntryCaller.File.
func (c EntryCaller) FileWithPackage() string {

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

	found := strings.LastIndexByte(c.File, '/')
	if found == -1 {
		return c.File
	}
	found = strings.LastIndexByte(c.File[:found], '/')
	if found == -1 {
		return c.File
	}

	return c.File[found+1:]
}
