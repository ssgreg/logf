package logf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWith(t *testing.T) {
	ctx := context.Background()

	ctx = With(ctx, String("k1", "v1"))
	assert.True(t, HasField(ctx, "k1"))
	assert.False(t, HasField(ctx, "k2"))

	ctx = With(ctx, String("k2", "v2"))
	assert.True(t, HasField(ctx, "k1"))
	assert.True(t, HasField(ctx, "k2"))
	assert.Equal(t, 2, len(Fields(ctx)))
}

func TestWithEmptyContext(t *testing.T) {
	ctx := context.Background()

	assert.False(t, HasField(ctx, "k1"))
	assert.Nil(t, Fields(ctx))
}

func TestContextHandler(t *testing.T) {
	sink := &testHandler{}
	cw := NewContextHandler(sink)

	ctx := With(context.Background(), String("request_id", "abc"))
	_ = cw.Handle(ctx, Entry{Level: LevelInfo, Text: "hello"})

	assert.NotNil(t, sink.Entry.Bag)
	assert.True(t, sink.Entry.Bag.HasField("request_id"))
	assert.Equal(t, "hello", sink.Entry.Text)
}

func TestContextHandlerNoBag(t *testing.T) {
	sink := &testHandler{}
	cw := NewContextHandler(sink)

	_ = cw.Handle(context.Background(), Entry{Level: LevelInfo, Text: "no bag"})

	assert.Nil(t, sink.Entry.Bag)
	assert.Equal(t, "no bag", sink.Entry.Text)
}

func TestContextHandlerPreservesLoggerBag(t *testing.T) {
	sink := &testHandler{}
	cw := NewContextHandler(sink)

	loggerBag := NewBag(String("service", "api"))
	ctx := With(context.Background(), String("request_id", "abc"))

	e := Entry{LoggerBag: loggerBag, Level: LevelInfo, Text: "both bags"}
	_ = cw.Handle(ctx, e)

	assert.NotNil(t, sink.Entry.LoggerBag)
	assert.NotNil(t, sink.Entry.Bag)
	assert.True(t, sink.Entry.LoggerBag.HasField("service"))
	assert.True(t, sink.Entry.Bag.HasField("request_id"))
}

func TestContextHandlerEncodesContextBag(t *testing.T) {
	ctx := With(context.Background(), String("rid", "123"))
	bag := BagFromContext(ctx)

	e := Entry{
		Bag:   bag,
		Level: LevelInfo,
		Text:  "test",
	}

	enc := JSON().Build()
	buf, _ := enc.Encode(e)

	assert.Contains(t, buf.String(), `"rid":"123"`)
	buf.Free()
}

func TestContextBagBeforeLoggerBag(t *testing.T) {
	e := Entry{
		Bag:       NewBag(String("rid", "123")),
		LoggerBag: NewBag(String("service", "api")),
		Level:     LevelInfo,
		Text:      "order",
	}

	enc := JSON().Build()
	buf, _ := enc.Encode(e)

	s := buf.String()
	buf.Free()
	assert.Contains(t, s, `"rid":"123"`)
	assert.Contains(t, s, `"service":"api"`)

	// Context bag (rid) should come before logger bag (service).
	ridPos := indexOf(s, `"rid"`)
	svcPos := indexOf(s, `"service"`)
	assert.True(t, ridPos < svcPos, "context bag should come before logger bag: rid@%d service@%d in %s", ridPos, svcPos, s)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestContextHandlerFieldSource(t *testing.T) {
	sink := &testHandler{}

	traceSource := FieldSource(func(ctx context.Context) []Field {
		if v := ctx.Value(traceKey{}); v != nil {
			return []Field{String("trace_id", v.(string))}
		}
		return nil
	})

	cw := NewContextHandler(sink, traceSource)

	ctx := context.WithValue(context.Background(), traceKey{}, "abc-123")
	ctx = With(ctx, String("rid", "r1"))

	_ = cw.Handle(ctx, Entry{Level: LevelInfo, Text: "ext", Fields: []Field{Int("x", 1)}})

	// Bag is set from context.
	assert.NotNil(t, sink.Entry.Bag)
	assert.True(t, sink.Entry.Bag.HasField("rid"))

	// External source fields are prepended to Entry.Fields.
	assert.Equal(t, 2, len(sink.Entry.Fields))
	assert.Equal(t, "trace_id", sink.Entry.Fields[0].Key)
	assert.Equal(t, "x", sink.Entry.Fields[1].Key)
}

func TestContextHandlerFieldSourceNoFields(t *testing.T) {
	sink := &testHandler{}

	emptySource := FieldSource(func(ctx context.Context) []Field {
		return nil
	})

	cw := NewContextHandler(sink, emptySource)
	_ = cw.Handle(context.Background(), Entry{Level: LevelInfo, Text: "noop", Fields: []Field{Int("x", 1)}})

	assert.Equal(t, 1, len(sink.Entry.Fields))
	assert.Equal(t, "x", sink.Entry.Fields[0].Key)
}

func TestContextHandlerMultipleSources(t *testing.T) {
	sink := &testHandler{}

	src1 := FieldSource(func(ctx context.Context) []Field {
		return []Field{String("a", "1")}
	})
	src2 := FieldSource(func(ctx context.Context) []Field {
		return []Field{String("b", "2")}
	})

	cw := NewContextHandler(sink, src1, src2)
	_ = cw.Handle(context.Background(), Entry{Level: LevelInfo, Text: "multi", Fields: []Field{String("c", "3")}})

	// Order: src1 fields, src2 fields, original fields.
	assert.Equal(t, 3, len(sink.Entry.Fields))
	assert.Equal(t, "a", sink.Entry.Fields[0].Key)
	assert.Equal(t, "b", sink.Entry.Fields[1].Key)
	assert.Equal(t, "c", sink.Entry.Fields[2].Key)
}

type traceKey struct{}

func TestContextBagCaching(t *testing.T) {
	ctx := With(context.Background(), String("rid", "123"))
	bag := BagFromContext(ctx)

	e := Entry{
		Bag:   bag,
		Level: LevelInfo,
		Text:  "test",
	}

	enc := JSON().Build()

	buf1, _ := enc.Encode(e)
	buf2, _ := enc.Encode(e)

	// Same bag, same version — second encode should hit cache.
	assert.Equal(t, buf1.String(), buf2.String())
	assert.Contains(t, buf1.String(), `"rid":"123"`)
	buf1.Free()
	buf2.Free()
}
