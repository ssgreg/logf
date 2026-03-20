package logf

import (
	"context"
	"errors"
	"io"
	"sync"
)

// NewRouter returns a RouterBuilder for constructing a fan-out Handler
// that sends log entries to multiple destinations. Each route groups
// outputs that share an encoder, so one Encode call serves all outputs
// in the group.
func NewRouter() *RouterBuilder {
	return &RouterBuilder{}
}

// RouterBuilder accumulates routes and builds a fan-out Handler. Add as
// many routes as you need — each route has an encoder and one or more
// outputs with independent level filters.
//
// Usage:
//
//	router, close, err := NewRouter().
//	    Route(jsonEnc,
//	        Output(LevelDebug, kibana),
//	        Output(LevelInfo, stderr),
//	    ).
//	    Build()
type RouterBuilder struct {
	groups []encoderGroup
	err    error
}

// Route adds an encoder group with the given outputs. All outputs in the
// same route share a single Encode call per entry — so sending JSON to
// both a file and a network socket costs exactly one encode, not two.
func (b *RouterBuilder) Route(enc Encoder, opts ...RouteOption) *RouterBuilder {
	g := encoderGroup{enc: enc}
	for _, opt := range opts {
		opt(&g)
	}
	b.groups = append(b.groups, g)
	return b
}

// Build validates the configuration and returns the Router as a Handler,
// plus a close function that flushes and syncs all writers. Always defer
// the close function to ensure data reaches its destination. Build returns
// an error if the configuration is invalid (no routes or no outputs).
func (b *RouterBuilder) Build() (Handler, func() error, error) {
	if b.err != nil {
		return nil, nil, b.err
	}

	var totalOutputs int
	for _, g := range b.groups {
		totalOutputs += len(g.outputs)
	}
	if totalOutputs == 0 {
		return nil, nil, errors.New("logf: router has no outputs")
	}

	r := &router{}

	for _, g := range b.groups {
		if len(g.outputs) == 0 {
			continue
		}
		if g.enc == nil {
			return nil, nil, errors.New("logf: route has nil encoder")
		}

		eg := routerEncoderGroup{
			enc:      g.enc,
			broadestLevel: g.outputs[0].level,
		}

		for _, o := range g.outputs {
			if o.level > eg.broadestLevel {
				eg.broadestLevel = o.level
			}
			ro := &routerOutput{
				level:   o.level,
				w:       o.w,
				closeFn: o.closeFn,
			}
			eg.outputs = append(eg.outputs, ro)
			r.allOutputs = append(r.allOutputs, ro)
		}

		r.groups = append(r.groups, eg)
	}

	// Compute the most permissive level across all groups for fast Enabled check.
	r.broadestLevel = r.groups[0].broadestLevel
	for _, g := range r.groups[1:] {
		if g.broadestLevel > r.broadestLevel {
			r.broadestLevel = g.broadestLevel
		}
	}

	return r, r.close, nil
}

type encoderGroup struct {
	enc     Encoder
	outputs []output
}

type output struct {
	level   Level
	w       Writer
	closeFn func() error // non-nil if the writer should be closed
}

// RouteOption configures a destination within a Route's encoder group.
// Use Output and OutputCloser to create RouteOptions.
type RouteOption func(*encoderGroup)

// Output returns a RouteOption that adds a destination with the given
// level filter and writer. Writes happen directly in the caller's
// goroutine — no channel, no background goroutine, zero per-message
// allocations. The Writer must be safe for concurrent use.
//
// For async I/O with batching and spike tolerance, wrap the writer in
// a SlabWriter before passing it to Output:
//
//	sw := logf.NewSlabWriter(conn).SlabSize(64*1024).SlabCount(8).FlushInterval(100*time.Millisecond).Build()
//	defer sw.Close()
//	router, close, _ := logf.NewRouter().
//	    Route(enc, logf.Output(logf.LevelDebug, sw)).
//	    Build()
func Output(level Level, w io.Writer) RouteOption {
	return func(g *encoderGroup) {
		g.outputs = append(g.outputs, output{level: level, w: WriterFromIO(w)})
	}
}

// OutputCloser is like Output but transfers ownership of the writer to
// the router — the router's close function will call Close on w after
// flushing. Perfect for SlabWriters and other resources you want the
// router to manage:
//
//	sw := logf.NewSlabWriter(conn).SlabSize(64*1024).SlabCount(8).Build()
//	router, close, _ := logf.NewRouter().
//	    Route(enc, logf.OutputCloser(logf.LevelDebug, sw)).
//	    Build()
//	defer close() // flushes and closes sw
func OutputCloser(level Level, w io.WriteCloser) RouteOption {
	return func(g *encoderGroup) {
		g.outputs = append(g.outputs, output{level: level, w: WriterFromIO(w), closeFn: w.Close})
	}
}

// router is the fan-out Handler built by RouterBuilder.
type router struct {
	groups     []routerEncoderGroup
	allOutputs []*routerOutput
	broadestLevel   Level
	closed     sync.Once
}

type routerEncoderGroup struct {
	enc      Encoder
	outputs  []*routerOutput
	broadestLevel Level
}

// routerOutput writes directly from the caller goroutine.
type routerOutput struct {
	level   Level
	w       Writer
	closeFn func() error
}

func (r *router) Enabled(_ context.Context, lvl Level) bool {
	return r.broadestLevel.Enabled(lvl)
}

func (r *router) Handle(_ context.Context, e Entry) error {
	var writeErr error

	for i := range r.groups {
		g := &r.groups[i]

		if !g.broadestLevel.Enabled(e.Level) {
			continue
		}

		buf, err := g.enc.Encode(e)
		if err != nil {
			writeErr = errors.Join(writeErr, err)
			continue
		}
		encoded := buf.Bytes()

		for _, o := range g.outputs {
			if !o.level.Enabled(e.Level) {
				continue
			}
			_, err := o.w.Write(encoded)
			if err != nil {
				writeErr = errors.Join(writeErr, err)
			}
		}

		buf.Free()
	}

	return writeErr
}

func (r *router) close() error {
	var err error
	r.closed.Do(func() {
		for _, o := range r.allOutputs {
			err = errors.Join(err, o.w.Flush())
			err = errors.Join(err, o.w.Sync())
			if o.closeFn != nil {
				err = errors.Join(err, o.closeFn())
			}
		}
	})
	return err
}
