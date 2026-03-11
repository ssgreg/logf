package benchmarks

import (
	"context"
	"testing"
	"time"

	logf "github.com/ssgreg/logf/v2"
)

func benchField(b *testing.B, f logf.Field) {
	logger := newLogfSync()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(ctx, "msg", f)
	}
}

// --- Scalar types ---

func BenchmarkLogf_Field_Bool(b *testing.B)     { benchField(b, logf.Bool("k", true)) }
func BenchmarkLogf_Field_Int64(b *testing.B)    { benchField(b, logf.Int64("k", 42)) }
func BenchmarkLogf_Field_Float64(b *testing.B)  { benchField(b, logf.Float64("k", 3.14)) }
func BenchmarkLogf_Field_String(b *testing.B)   { benchField(b, logf.String("k", "value")) }
func BenchmarkLogf_Field_Duration(b *testing.B) { benchField(b, logf.Duration("k", 5*time.Second)) }
func BenchmarkLogf_Field_Time(b *testing.B)     { benchField(b, logf.Time("k", heavyTime)) }
func BenchmarkLogf_Field_Error(b *testing.B)    { benchField(b, logf.NamedError("k", errExample)) }

// --- Slice types (copying) ---

func BenchmarkLogf_Field_Bytes(b *testing.B)     { benchField(b, logf.Bytes("k", heavyBytes)) }
func BenchmarkLogf_Field_Ints64(b *testing.B)    { benchField(b, logf.Ints64("k", heavyInts64)) }
func BenchmarkLogf_Field_Strings(b *testing.B)   { benchField(b, logf.Strings("k", heavyStrings)) }
func BenchmarkLogf_Field_Bools(b *testing.B)     { benchField(b, logf.Bools("k", []bool{true, false, true, false})) }
func BenchmarkLogf_Field_Floats64(b *testing.B)  { benchField(b, logf.Floats64("k", []float64{1.1, 2.2, 3.3, 4.4})) }
func BenchmarkLogf_Field_Durations(b *testing.B) { benchField(b, logf.Durations("k", []time.Duration{time.Second, time.Minute, time.Hour, time.Millisecond})) }

// --- Slice types (const — no copy, async-unsafe) ---

func BenchmarkLogf_Field_ConstBytes(b *testing.B)     { benchField(b, logf.ConstBytes("k", heavyBytes)) }
func BenchmarkLogf_Field_ConstInts64(b *testing.B)    { benchField(b, logf.ConstInts64("k", heavyInts64)) }
func BenchmarkLogf_Field_ConstStrings(b *testing.B)   { benchField(b, logf.ConstStrings("k", heavyStrings)) }
func BenchmarkLogf_Field_ConstBools(b *testing.B)     { benchField(b, logf.ConstBools("k", []bool{true, false, true, false})) }
func BenchmarkLogf_Field_ConstFloats64(b *testing.B)  { benchField(b, logf.ConstFloats64("k", []float64{1.1, 2.2, 3.3, 4.4})) }
func BenchmarkLogf_Field_ConstDurations(b *testing.B) { benchField(b, logf.ConstDurations("k", []time.Duration{time.Second, time.Minute, time.Hour, time.Millisecond})) }

// --- Composite types ---

func BenchmarkLogf_Field_Object(b *testing.B)  { benchField(b, logf.Object("k", heavyUser)) }
func BenchmarkLogf_Field_Array(b *testing.B)   { benchField(b, logf.Array("k", benchArray{1, 2, 3, 4})) }
func BenchmarkLogf_Field_Group(b *testing.B)   { benchField(b, logf.Group("k", logf.String("a", "1"), logf.Int("b", 2))) }
func BenchmarkLogf_Field_Stringer(b *testing.B) { benchField(b, logf.Stringer("k", &benchStringer{Value: "hello"})) }
func BenchmarkLogf_Field_Formatter(b *testing.B) { benchField(b, logf.Formatter("k", "%d", 42)) }

// --- Any variants ---

func BenchmarkLogf_Field_AnyValue(b *testing.B)       { benchField(b, logf.Any("k", 42)) }
func BenchmarkLogf_Field_AnySnapshotter(b *testing.B) { benchField(b, logf.Any("k", &benchSnapshotter{Value: "snap"})) }
func BenchmarkLogf_Field_AnyStringer(b *testing.B)    { benchField(b, logf.Any("k", &benchStringer{Value: "hello"})) }
func BenchmarkLogf_Field_AnyMap(b *testing.B)         { benchField(b, logf.Any("k", map[string]int{"a": 1, "b": 2})) }
