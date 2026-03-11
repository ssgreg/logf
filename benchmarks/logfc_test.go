package benchmarks

import (
	"context"
	"testing"

	logf "github.com/ssgreg/logf/v2"
)

// GetFromContext uses logf.FromContext directly (same as logfc.Get).
// PutToContext uses logf.NewContext directly — pure context.WithValue cost.

func BenchmarkLogfc_GetFromContext(b *testing.B) {
	logger := newLogfSync()
	ctx := logf.NewContext(context.Background(), logger)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logf.FromContext(ctx)
	}
}

func BenchmarkLogfc_PutToContext(b *testing.B) {
	logger := newLogfSync()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logf.NewContext(ctx, logger)
	}
}
