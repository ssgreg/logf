package logf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileWithPackage(t *testing.T) {
	cases := []struct {
		file   string
		golden string
	}{
		{"/a/b/c/d.go", "c/d.go"},
		{"c/d.go", "c/d.go"},
		{"d.go", "d.go"},
	}

	for _, c := range cases {
		assert.Equal(t, c.golden, fileWithPackage(c.file))
	}
}

func TestCallerPC(t *testing.T) {
	pc := CallerPC(0)
	assert.NotZero(t, pc)

	file, line := callerFrame(pc)
	assert.True(t, line > 0 && line < 1000)
	assert.Equal(t, "logf/caller_test.go", fileWithPackage(file))
	assert.Contains(t, file, "/logf/caller_test.go")
}

func TestShortCallerEncoder(t *testing.T) {
	enc := testTypeEncoder{}
	pc := CallerPC(0)
	ShortCallerEncoder(pc, &enc)

	assert.Contains(t, enc.result, "logf/caller_test.go:")
}

func TestFullCallerEncoder(t *testing.T) {
	enc := testTypeEncoder{}
	pc := CallerPC(0)
	FullCallerEncoder(pc, &enc)

	assert.Contains(t, enc.result, "/logf/caller_test.go:")
}
