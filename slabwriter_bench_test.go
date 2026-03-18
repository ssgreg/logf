package logf

import (
	"os"
	"runtime"
	"testing"
	"time"
)

// discardWriter implements Writer, discards all data.
type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
func (discardWriter) Flush() error                { return nil }
func (discardWriter) Sync() error                 { return nil }

var msg = make([]byte, 200) // typical log message size

// BenchmarkSlabBuffer measures throughput: producer → slab buffer → I/O goroutine → discard.
func BenchmarkSlabBuffer(b *testing.B) {
	slab := NewSlabWriter(discardWriter{}, 64*1024, 8)

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = slab.Write(msg)
	}
	b.StopTimer()
	_ = slab.Close()
}

// BenchmarkChannel measures throughput: producer → chan []byte → consumer goroutine → discard.
func BenchmarkChannel(b *testing.B) {
	ch := make(chan []byte, runtime.NumCPU()*2)
	if cap(ch) < 4 {
		ch = make(chan []byte, 4)
	}
	done := make(chan struct{})
	w := discardWriter{}
	go func() {
		defer close(done)
		for data := range ch {
			_, _ = w.Write(data)
		}
	}()

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- msg
	}
	b.StopTimer()
	close(ch)
	<-done
}

// --- Parallel benchmarks: multiple producers → channel → writer ---

// BenchmarkParallelSlabBufferSlowIO: N producers → channel → slab buffer → slow writer.
func BenchmarkParallelSlabBufferSlowIO(b *testing.B) {
	sw := &slowWriter{delay: 100 * time.Microsecond}
	slab := NewSlabWriter(sw, 64*1024, 8)
	ch := make(chan []byte, runtime.NumCPU()*2)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for data := range ch {
			_, _ = slab.Write(data)
		}
		_ = slab.Close()
	}()

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch <- msg
		}
	})
	b.StopTimer()
	close(ch)
	<-done
	b.ReportMetric(float64(b.N)/float64(sw.writes), "msgs/write")
}

// BenchmarkParallelChannelSlowIO: N producers → channel → direct slow writer (no buffering).
func BenchmarkParallelChannelSlowIO(b *testing.B) {
	sw := &slowWriter{delay: 100 * time.Microsecond}
	ch := make(chan []byte, runtime.NumCPU()*2)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for data := range ch {
			_, _ = sw.Write(data)
		}
	}()

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch <- msg
		}
	})
	b.StopTimer()
	close(ch)
	<-done
	b.ReportMetric(float64(b.N)/float64(sw.writes), "msgs/write")
}

// --- Burst benchmarks: proportional I/O (10μs + 1ns/byte) ---

func BenchmarkBurstSlabBuffer(b *testing.B) {
	sw := &slowWriter{delay: 10 * time.Microsecond, perByte: time.Nanosecond}
	slab := NewSlabWriter(sw, 64*1024, 8)
	ch := make(chan []byte, 20)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for data := range ch {
			_, _ = slab.Write(data)
		}
		_ = slab.Close()
	}()

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- msg
	}
	b.StopTimer()
	close(ch)
	<-done
	b.ReportMetric(float64(b.N)/float64(sw.writes), "msgs/write")
}

// --- Real file I/O benchmarks ---

func BenchmarkFileSlabBuffer(b *testing.B) {
	f, err := os.CreateTemp("", "bench-slab-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	slab := NewSlabWriter(f, 64*1024, 8)
	ch := make(chan []byte, 20)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for data := range ch {
			_, _ = slab.Write(data)
		}
		_ = slab.Close()
	}()

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- msg
	}
	b.StopTimer()
	close(ch)
	<-done
}

// --- Variant 7: concurrent slab (built-in mutex), no channel ---

// BenchmarkConcurrentSlabBuffer: single producer → concurrent slab → I/O goroutine → discard.
func BenchmarkConcurrentSlabBuffer(b *testing.B) {
	slab := NewSlabWriter(discardWriter{}, 64*1024, 8)

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = slab.Write(msg)
	}
	b.StopTimer()
	_ = slab.Close()
}

// BenchmarkParallelConcurrentSlabDiscard: N producers → concurrent slab → discard (no I/O cost).
func BenchmarkParallelConcurrentSlabDiscard(b *testing.B) {
	slab := NewSlabWriter(discardWriter{}, 64*1024, 8)

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = slab.Write(msg)
		}
	})
	b.StopTimer()
	_ = slab.Close()
}

// BenchmarkParallelChannelSlabDiscard: N producers → channel → slab → discard (no I/O cost).
func BenchmarkParallelChannelSlabDiscard(b *testing.B) {
	slab := NewSlabWriter(discardWriter{}, 64*1024, 8)
	ch := make(chan []byte, runtime.NumCPU()*2)
	if cap(ch) < 4 {
		ch = make(chan []byte, 4)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for data := range ch {
			_, _ = slab.Write(data)
		}
		_ = slab.Close()
	}()

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cp := make([]byte, len(msg))
			copy(cp, msg)
			ch <- cp
		}
	})
	b.StopTimer()
	close(ch)
	<-done
}

// BenchmarkParallelConcurrentSlabSlowIO: N producers → concurrent slab → slow writer.
func BenchmarkParallelConcurrentSlabSlowIO(b *testing.B) {
	sw := &slowWriter{delay: 100 * time.Microsecond}
	slab := NewSlabWriter(sw, 64*1024, 8)

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = slab.Write(msg)
		}
	})
	b.StopTimer()
	_ = slab.Close()
	b.ReportMetric(float64(b.N)/float64(sw.writes), "msgs/write")
}

// BenchmarkBurstConcurrentSlab: single producer → concurrent slab → proportional I/O.
func BenchmarkBurstConcurrentSlab(b *testing.B) {
	sw := &slowWriter{delay: 10 * time.Microsecond, perByte: time.Nanosecond}
	slab := NewSlabWriter(sw, 64*1024, 8)

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = slab.Write(msg)
	}
	b.StopTimer()
	_ = slab.Close()
	b.ReportMetric(float64(b.N)/float64(sw.writes), "msgs/write")
}

// BenchmarkFileConcurrentSlab: single producer → concurrent slab → real file.
func BenchmarkFileConcurrentSlab(b *testing.B) {
	f, err := os.CreateTemp("", "bench-cslab-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	slab := NewSlabWriter(f, 64*1024, 8)

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = slab.Write(msg)
	}
	b.StopTimer()
	_ = slab.Close()
}

// --- Parallel proportional I/O benchmarks (10µs + 1ns/byte) ---

// BenchmarkParallelBurstConcurrentSlab: N producers → concurrent slab → proportional slow writer.
func BenchmarkParallelBurstConcurrentSlab(b *testing.B) {
	sw := &slowWriter{delay: 10 * time.Microsecond, perByte: time.Nanosecond}
	slab := NewSlabWriter(sw, 64*1024, 8)

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = slab.Write(msg)
		}
	})
	b.StopTimer()
	_ = slab.Close()
	b.ReportMetric(float64(b.N)/float64(sw.writes), "msgs/write")
}

// --- Parallel real file I/O benchmarks ---

// BenchmarkParallelFileConcurrentSlab: N producers → concurrent slab → real file.
func BenchmarkParallelFileConcurrentSlab(b *testing.B) {
	f, err := os.CreateTemp("", "bench-pcslab2-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	slab := NewSlabWriter(f, 64*1024, 8)

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = slab.Write(msg)
		}
	})
	b.StopTimer()
	_ = slab.Close()
}

// BenchmarkParallelFileConcurrentSlabFlush: N producers → concurrent slab (idle flush 100ms) → real file.
func BenchmarkParallelFileConcurrentSlabFlush(b *testing.B) {
	f, err := os.CreateTemp("", "bench-pcslabf-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	slab := NewSlabWriter(f, 64*1024, 8, WithFlushInterval(100*time.Millisecond))

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = slab.Write(msg)
		}
	})
	b.StopTimer()
	_ = slab.Close()
}

// BenchmarkParallelFileConcurrentSlabSmall: N producers → concurrent slab (2×32KB=64KB) → real file.
func BenchmarkParallelFileConcurrentSlabSmall(b *testing.B) {
	f, err := os.CreateTemp("", "bench-pcslabs-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	slab := NewSlabWriter(f, 32*1024, 2)

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = slab.Write(msg)
		}
	})
	b.StopTimer()
	_ = slab.Close()
}

// BenchmarkParallelFileChannelSlab: N producers → channel → slab → real file.
func BenchmarkParallelFileChannelSlab(b *testing.B) {
	f, err := os.CreateTemp("", "bench-pcslab-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	slab := NewSlabWriter(f, 64*1024, 8)
	ch := make(chan []byte, runtime.NumCPU()*2)
	if cap(ch) < 4 {
		ch = make(chan []byte, 4)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for data := range ch {
			_, _ = slab.Write(data)
		}
		_ = slab.Close()
	}()

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cp := make([]byte, len(msg))
			copy(cp, msg)
			ch <- cp
		}
	})
	b.StopTimer()
	close(ch)
	<-done
}

// slowWriter simulates I/O latency. If perByte > 0, delay scales with
// data size (realistic). Otherwise fixed delay per Write call.
type slowWriter struct {
	writes  int64
	delay   time.Duration
	perByte time.Duration
}

func (w *slowWriter) Write(p []byte) (int, error) {
	w.writes++
	d := w.delay
	if w.perByte > 0 {
		d += w.perByte * time.Duration(len(p))
	}
	if d > 0 {
		time.Sleep(d)
	}
	return len(p), nil
}
func (w *slowWriter) Flush() error { return nil }
func (w *slowWriter) Sync() error  { return nil }
