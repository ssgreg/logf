package benchmarks

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	logf "github.com/ssgreg/logf/v2"
	"github.com/ssgreg/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// =============================================================================
// Test 1: Latency Distribution (p50/p99/p999)
// Shows that logf async has much tighter tail latency than sync loggers
// because the caller never waits for write(fd).
// =============================================================================

const latencySamples = 50000

func percentile(sorted []time.Duration, p float64) time.Duration {
	idx := int(float64(len(sorted)) * p)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func reportLatency(t *testing.T, name string, samples []time.Duration) {
	sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })
	t.Logf("%-35s  p50=%7s  p90=%7s  p99=%7s  p999=%7s  max=%7s",
		name,
		percentile(samples, 0.50),
		percentile(samples, 0.90),
		percentile(samples, 0.99),
		percentile(samples, 0.999),
		samples[len(samples)-1],
	)
}

func TestLatencyDistribution(t *testing.T) {
	f, err := os.CreateTemp("", "latency-bench-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	t.Logf("Writing %d log entries per logger to %s", latencySamples, f.Name())
	t.Logf("")

	// --- logf async (channel writer + buffered appender) ---
	{
		ff, _ := os.CreateTemp("", "latency-logf-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		w, cl := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
			Appender: logf.NewWriteAppender(ff, logf.NewJSONEncoder.Default()),
		})
		defer cl()
		logger := logf.NewLogger(w).WithCaller(false)
		ctx := context.Background()

		samples := make([]time.Duration, latencySamples)
		runtime.GC()
		for i := 0; i < latencySamples; i++ {
			start := time.Now()
			logger.Info(ctx, "request handled")
			samples[i] = time.Since(start)
		}
		reportLatency(t, "logf async", samples)
	}

	// --- logf sync (unbuffered entry writer, buffered appender) ---
	{
		ff, _ := os.CreateTemp("", "latency-logfsync-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		enc := logf.NewJSONEncoder.Default()
		w := logf.NewSyncWriter(logf.LevelDebug, logf.NewWriteAppender(ff, enc))
		logger := logf.NewLogger(w).WithCaller(false)
		ctx := context.Background()

		samples := make([]time.Duration, latencySamples)
		runtime.GC()
		for i := 0; i < latencySamples; i++ {
			start := time.Now()
			logger.Info(ctx, "request handled")
			samples[i] = time.Since(start)
		}
		reportLatency(t, "logf sync (buffered appender)", samples)
	}

	// --- zap (sync, unbuffered) ---
	{
		ff, _ := os.CreateTemp("", "latency-zap-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		ec := zap.NewProductionEncoderConfig()
		ec.EncodeDuration = zapcore.NanosDurationEncoder
		ec.EncodeTime = zapcore.ISO8601TimeEncoder
		enc := zapcore.NewJSONEncoder(ec)
		logger := zap.New(zapcore.NewCore(enc, zapcore.AddSync(ff), zap.DebugLevel))

		samples := make([]time.Duration, latencySamples)
		runtime.GC()
		for i := 0; i < latencySamples; i++ {
			start := time.Now()
			logger.Info("request handled")
			samples[i] = time.Since(start)
		}
		reportLatency(t, "zap (sync, unbuffered)", samples)
	}

	// --- zap (sync, BufferedWriteSyncer) ---
	{
		ff, _ := os.CreateTemp("", "latency-zapbuf-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		ec := zap.NewProductionEncoderConfig()
		ec.EncodeDuration = zapcore.NanosDurationEncoder
		ec.EncodeTime = zapcore.ISO8601TimeEncoder
		enc := zapcore.NewJSONEncoder(ec)
		bws := &zapcore.BufferedWriteSyncer{WS: zapcore.AddSync(ff)}
		defer bws.Stop()
		logger := zap.New(zapcore.NewCore(enc, bws, zap.DebugLevel))

		samples := make([]time.Duration, latencySamples)
		runtime.GC()
		for i := 0; i < latencySamples; i++ {
			start := time.Now()
			logger.Info("request handled")
			samples[i] = time.Since(start)
		}
		reportLatency(t, "zap (sync, buffered)", samples)
	}

	// --- zerolog (sync, unbuffered) ---
	{
		ff, _ := os.CreateTemp("", "latency-zerolog-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		logger := zerolog.New(ff).With().Timestamp().Logger()

		samples := make([]time.Duration, latencySamples)
		runtime.GC()
		for i := 0; i < latencySamples; i++ {
			start := time.Now()
			logger.Info().Msg("request handled")
			samples[i] = time.Since(start)
		}
		reportLatency(t, "zerolog (sync, unbuffered)", samples)
	}

	// --- zerolog (sync, buffered) ---
	{
		ff, _ := os.CreateTemp("", "latency-zerologbuf-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		lbw := &lockedBufWriter{bw: bufio.NewWriterSize(ff, 4096)}
		defer lbw.Flush()
		logger := zerolog.New(lbw).With().Timestamp().Logger()

		samples := make([]time.Duration, latencySamples)
		runtime.GC()
		for i := 0; i < latencySamples; i++ {
			start := time.Now()
			logger.Info().Msg("request handled")
			samples[i] = time.Since(start)
		}
		reportLatency(t, "zerolog (sync, buffered)", samples)
	}

	// --- slog (sync, unbuffered) ---
	{
		ff, _ := os.CreateTemp("", "latency-slog-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		logger := slog.New(slog.NewJSONHandler(ff, &slog.HandlerOptions{Level: slog.LevelDebug}))

		samples := make([]time.Duration, latencySamples)
		runtime.GC()
		for i := 0; i < latencySamples; i++ {
			start := time.Now()
			logger.Info("request handled")
			samples[i] = time.Since(start)
		}
		reportLatency(t, "slog (sync, unbuffered)", samples)
	}

	// --- slog (sync, buffered) ---
	{
		ff, _ := os.CreateTemp("", "latency-slogbuf-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		bw := bufio.NewWriterSize(ff, 4096)
		defer bw.Flush()
		logger := slog.New(slog.NewJSONHandler(bw, &slog.HandlerOptions{Level: slog.LevelDebug}))

		samples := make([]time.Duration, latencySamples)
		runtime.GC()
		for i := 0; i < latencySamples; i++ {
			start := time.Now()
			logger.Info("request handled")
			samples[i] = time.Since(start)
		}
		reportLatency(t, "slog (sync, buffered)", samples)
	}

	// --- logrus ---
	{
		ff, _ := os.CreateTemp("", "latency-logrus-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		logger := &logrus.Logger{
			Out:       ff,
			Formatter: new(logrus.JSONFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.DebugLevel,
		}

		samples := make([]time.Duration, latencySamples)
		runtime.GC()
		for i := 0; i < latencySamples; i++ {
			start := time.Now()
			logger.Info("request handled")
			samples[i] = time.Since(start)
		}
		reportLatency(t, "logrus (sync, unbuffered)", samples)
	}
}

// =============================================================================
// Test 2: Slow I/O Simulation
// A writer that occasionally sleeps to simulate disk latency spikes.
// Shows logf's channel absorbs spikes while sync loggers block callers.
// =============================================================================

type slowWriter struct {
	mu      sync.Mutex
	w       *os.File
	slowPct float64 // fraction of writes that are slow (0.0 - 1.0)
	delay   time.Duration
}

func (s *slowWriter) Write(p []byte) (int, error) {
	if rand.Float64() < s.slowPct {
		time.Sleep(s.delay)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

func TestSlowIOLatency(t *testing.T) {
	t.Logf("Simulating slow I/O: 2%% of writes sleep 1ms")
	t.Logf("Parallel: %d goroutines, %d samples total", runtime.NumCPU(), latencySamples)
	t.Logf("")

	// --- logf async ---
	{
		ff, _ := os.CreateTemp("", "slowio-logf-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		sw := &slowWriter{w: ff, slowPct: 0.02, delay: 1 * time.Millisecond}
		w, cl := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
			Appender: logf.NewWriteAppender(sw, logf.NewJSONEncoder.Default()),
		})
		defer cl()
		logger := logf.NewLogger(w).WithCaller(false)

		samples := collectParallelLatency(func() {
			logger.Info(context.Background(), "request handled")
		})
		reportLatency(t, "logf async (channel)", samples)
	}

	// --- zap sync unbuffered ---
	{
		ff, _ := os.CreateTemp("", "slowio-zap-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		sw := &slowWriter{w: ff, slowPct: 0.02, delay: 1 * time.Millisecond}
		ec := zap.NewProductionEncoderConfig()
		ec.EncodeDuration = zapcore.NanosDurationEncoder
		ec.EncodeTime = zapcore.ISO8601TimeEncoder
		enc := zapcore.NewJSONEncoder(ec)
		logger := zap.New(zapcore.NewCore(enc, zapcore.AddSync(sw), zap.DebugLevel))

		samples := collectParallelLatency(func() {
			logger.Info("request handled")
		})
		reportLatency(t, "zap (sync, unbuffered)", samples)
	}

	// --- zap sync buffered ---
	{
		ff, _ := os.CreateTemp("", "slowio-zapbuf-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		sw := &slowWriter{w: ff, slowPct: 0.02, delay: 1 * time.Millisecond}
		ec := zap.NewProductionEncoderConfig()
		ec.EncodeDuration = zapcore.NanosDurationEncoder
		ec.EncodeTime = zapcore.ISO8601TimeEncoder
		enc := zapcore.NewJSONEncoder(ec)
		bws := &zapcore.BufferedWriteSyncer{WS: zapcore.AddSync(sw)}
		defer bws.Stop()
		logger := zap.New(zapcore.NewCore(enc, bws, zap.DebugLevel))

		samples := collectParallelLatency(func() {
			logger.Info("request handled")
		})
		reportLatency(t, "zap (sync, buffered)", samples)
	}

	// --- zerolog sync ---
	{
		ff, _ := os.CreateTemp("", "slowio-zerolog-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		sw := &slowWriter{w: ff, slowPct: 0.02, delay: 1 * time.Millisecond}
		logger := zerolog.New(sw).With().Timestamp().Logger()

		samples := collectParallelLatency(func() {
			logger.Info().Msg("request handled")
		})
		reportLatency(t, "zerolog (sync, unbuffered)", samples)
	}

	// --- slog sync ---
	{
		ff, _ := os.CreateTemp("", "slowio-slog-*.log")
		defer os.Remove(ff.Name())
		defer ff.Close()

		sw := &slowWriter{w: ff, slowPct: 0.02, delay: 1 * time.Millisecond}
		logger := slog.New(slog.NewJSONHandler(sw, &slog.HandlerOptions{Level: slog.LevelDebug}))

		samples := collectParallelLatency(func() {
			logger.Info("request handled")
		})
		reportLatency(t, "slog (sync, unbuffered)", samples)
	}
}

func collectParallelLatency(logFn func()) []time.Duration {
	numG := runtime.NumCPU()
	perG := latencySamples / numG
	allSamples := make([][]time.Duration, numG)

	runtime.GC()
	var wg sync.WaitGroup
	wg.Add(numG)
	for g := 0; g < numG; g++ {
		g := g
		go func() {
			defer wg.Done()
			s := make([]time.Duration, perG)
			for i := 0; i < perG; i++ {
				start := time.Now()
				logFn()
				s[i] = time.Since(start)
			}
			allSamples[g] = s
		}()
	}
	wg.Wait()

	var merged []time.Duration
	for _, s := range allSamples {
		merged = append(merged, s...)
	}
	return merged
}

// =============================================================================
// Test 3: Goroutine Scalability
// Measures throughput (logs/sec) as goroutine count increases.
// logf async should scale better because goroutines don't contend on write(fd).
// =============================================================================

func TestGoroutineScalability(t *testing.T) {
	goroutineCounts := []int{1, 2, 4, 8, 16, 32, 64}
	const totalOps = 200000

	t.Logf("Throughput (logs/sec) by goroutine count, %d total ops, file I/O", totalOps)
	t.Logf("")

	header := fmt.Sprintf("%-30s", "Logger")
	for _, g := range goroutineCounts {
		header += fmt.Sprintf(" %8s", fmt.Sprintf("%dG", g))
	}
	t.Log(header)

	// logf async
	{
		results := make([]float64, len(goroutineCounts))
		for gi, numG := range goroutineCounts {
			ff, _ := os.CreateTemp("", "scale-logf-*.log")
			w, cl := logf.NewChannelWriter(logf.LevelDebug, logf.ChannelWriterConfig{
				Appender: logf.NewWriteAppender(ff, logf.NewJSONEncoder.Default()),
			})
			logger := logf.NewLogger(w).WithCaller(false)

			results[gi] = measureThroughput(numG, totalOps, func() {
				logger.Info(context.Background(), "request handled")
			})
			cl()
			ff.Close()
			os.Remove(ff.Name())
		}
		logResults(t, "logf async", goroutineCounts, results)
	}

	// zap unbuffered
	{
		results := make([]float64, len(goroutineCounts))
		for gi, numG := range goroutineCounts {
			ff, _ := os.CreateTemp("", "scale-zap-*.log")
			ec := zap.NewProductionEncoderConfig()
			ec.EncodeDuration = zapcore.NanosDurationEncoder
			ec.EncodeTime = zapcore.ISO8601TimeEncoder
			enc := zapcore.NewJSONEncoder(ec)
			logger := zap.New(zapcore.NewCore(enc, zapcore.AddSync(ff), zap.DebugLevel))

			results[gi] = measureThroughput(numG, totalOps, func() {
				logger.Info("request handled")
			})
			ff.Close()
			os.Remove(ff.Name())
		}
		logResults(t, "zap (unbuffered)", goroutineCounts, results)
	}

	// zap buffered
	{
		results := make([]float64, len(goroutineCounts))
		for gi, numG := range goroutineCounts {
			ff, _ := os.CreateTemp("", "scale-zapbuf-*.log")
			ec := zap.NewProductionEncoderConfig()
			ec.EncodeDuration = zapcore.NanosDurationEncoder
			ec.EncodeTime = zapcore.ISO8601TimeEncoder
			enc := zapcore.NewJSONEncoder(ec)
			bws := &zapcore.BufferedWriteSyncer{WS: zapcore.AddSync(ff)}
			logger := zap.New(zapcore.NewCore(enc, bws, zap.DebugLevel))

			results[gi] = measureThroughput(numG, totalOps, func() {
				logger.Info("request handled")
			})
			bws.Stop()
			ff.Close()
			os.Remove(ff.Name())
		}
		logResults(t, "zap (buffered)", goroutineCounts, results)
	}

	// zerolog unbuffered
	{
		results := make([]float64, len(goroutineCounts))
		for gi, numG := range goroutineCounts {
			ff, _ := os.CreateTemp("", "scale-zerolog-*.log")
			logger := zerolog.New(ff).With().Timestamp().Logger()

			results[gi] = measureThroughput(numG, totalOps, func() {
				logger.Info().Msg("request handled")
			})
			ff.Close()
			os.Remove(ff.Name())
		}
		logResults(t, "zerolog (unbuffered)", goroutineCounts, results)
	}

	// slog unbuffered
	{
		results := make([]float64, len(goroutineCounts))
		for gi, numG := range goroutineCounts {
			ff, _ := os.CreateTemp("", "scale-slog-*.log")
			logger := slog.New(slog.NewJSONHandler(ff, &slog.HandlerOptions{Level: slog.LevelDebug}))

			results[gi] = measureThroughput(numG, totalOps, func() {
				logger.Info("request handled")
			})
			ff.Close()
			os.Remove(ff.Name())
		}
		logResults(t, "slog (unbuffered)", goroutineCounts, results)
	}

	// zerolog buffered
	{
		results := make([]float64, len(goroutineCounts))
		for gi, numG := range goroutineCounts {
			ff, _ := os.CreateTemp("", "scale-zerologbuf-*.log")
			lbw := &lockedBufWriter{bw: bufio.NewWriterSize(ff, 4096)}
			logger := zerolog.New(lbw).With().Timestamp().Logger()

			results[gi] = measureThroughput(numG, totalOps, func() {
				logger.Info().Msg("request handled")
			})
			lbw.Flush()
			ff.Close()
			os.Remove(ff.Name())
		}
		logResults(t, "zerolog (buffered)", goroutineCounts, results)
	}
}

func measureThroughput(numG, totalOps int, logFn func()) float64 {
	perG := totalOps / numG
	runtime.GC()

	var wg sync.WaitGroup
	wg.Add(numG)

	start := time.Now()
	for g := 0; g < numG; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				logFn()
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	return float64(numG*perG) / elapsed.Seconds()
}

func logResults(t *testing.T, name string, goroutines []int, results []float64) {
	line := fmt.Sprintf("%-30s", name)
	for _, r := range results {
		if r >= 1_000_000 {
			line += fmt.Sprintf(" %7.1fM", r/1_000_000)
		} else {
			line += fmt.Sprintf(" %7.0fK", r/1_000)
		}
	}
	t.Log(line)
}
