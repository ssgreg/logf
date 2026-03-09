package logf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	// Check if no logger is associated with the Context — returns DisabledLogger.
	assert.Equal(t, DisabledLogger(), FromContext(context.Background()))

	logger := NewDisabledLogger()
	ctx := NewContext(context.Background(), logger)
	// First try.
	assert.Equal(t, logger, FromContext(ctx))
	// Second try.
	assert.Equal(t, logger, FromContext(ctx))
}
