package logf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntryCallerFileWithPackage(t *testing.T) {
	cases := []struct {
		caller EntryCaller
		golden string
	}{
		{
			caller: EntryCaller{0, "/a/b/c/d.go", 66, true},
			golden: "c/d.go",
		},
		{
			caller: EntryCaller{0, "c/d.go", 66, true},
			golden: "c/d.go",
		},
		{
			caller: EntryCaller{0, "d.go", 66, true},
			golden: "d.go",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.golden, c.caller.FileWithPackage())
	}
}

func TestEntryCaller(t *testing.T) {
	caller := NewEntryCaller(0)

	assert.True(t, caller.Specified)
	assert.True(t, caller.Line > 0 && caller.Line < 1000)
	assert.Equal(t, "logf/caller_test.go", caller.FileWithPackage())
	assert.Contains(t, caller.File, "/logf/caller_test.go")
}

func TestShortCallerEncoder(t *testing.T) {
	enc := testTypeEncoder{}
	caller := EntryCaller{0, "/a/b/c/d.go", 66, true}
	ShortCallerEncoder(caller, &enc)

	assert.EqualValues(t, "c/d.go:66", enc.result)
}

func TestFullCallerEncoder(t *testing.T) {
	enc := testTypeEncoder{}
	caller := EntryCaller{0, "/a/b/c/d.go", 66, true}
	FullCallerEncoder(caller, &enc)

	assert.EqualValues(t, "/a/b/c/d.go:66", enc.result)
}
