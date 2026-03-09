package logf

import (
	"context"
	"sync/atomic"
)

var nextVersion int32

// Bag is an immutable collection of Fields with a version for cache keying.
// Bag is safe to share across goroutines.
type Bag struct {
	fields  []Field
	version int32
}

// NewBag creates a new Bag with the given fields.
func NewBag(fs ...Field) *Bag {
	return &Bag{
		fields:  fs,
		version: atomic.AddInt32(&nextVersion, 1),
	}
}

// With returns a new Bag that contains both the existing fields and the
// given additional fields. The original Bag is not modified.
func (b *Bag) With(fs ...Field) *Bag {
	if b == nil {
		return NewBag(fs...)
	}

	merged := make([]Field, len(b.fields)+len(fs))
	copy(merged, b.fields)
	copy(merged[len(b.fields):], fs)

	return &Bag{
		fields:  merged,
		version: atomic.AddInt32(&nextVersion, 1),
	}
}

// Fields returns the fields stored in the Bag.
func (b *Bag) Fields() []Field {
	if b == nil {
		return nil
	}

	return b.fields
}

// Version returns the Bag's version, usable as a cache key.
func (b *Bag) Version() int32 {
	if b == nil {
		return 0
	}

	return b.version
}

// HasField reports whether the Bag contains a field with the given key.
func (b *Bag) HasField(key string) bool {
	if b == nil {
		return false
	}

	for i := range b.fields {
		if b.fields[i].Key == key {
			return true
		}
	}

	return false
}

type bagKey struct{}

// ContextWithBag returns a new context with the given Bag associated.
func ContextWithBag(ctx context.Context, bag *Bag) context.Context {
	return context.WithValue(ctx, bagKey{}, bag)
}

// BagFromContext returns the Bag associated with the context, or nil.
func BagFromContext(ctx context.Context) *Bag {
	bag, _ := ctx.Value(bagKey{}).(*Bag)

	return bag
}
