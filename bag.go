package logf

import (
	"context"
	"sync/atomic"
)

// maxCacheSlots is the maximum number of encoder cache slots per Bag.
// Typical setup: 1 encoder (JSON to file). Two slots provide headroom
// for a second destination (e.g., text to stderr). Encoders beyond
// maxCacheSlots work correctly but without caching.
const maxCacheSlots = 2

// bagCache holds per-encoder cached encoded bytes.
// Allocated lazily on first cache store via Bag.StoreCache.
type bagCache struct {
	slots [maxCacheSlots][]byte
}

// nextSlot is the global counter for encoder slot allocation.
// Slots are 1-based: 0 means "uninitialized / no caching".
var nextSlot atomic.Int32

// AllocEncoderSlot returns a unique 1-based slot index for an encoder to
// use with Bag.LoadCache and Bag.StoreCache. Call this once when you
// create an encoder — the slot lets the Bag cache encoded bytes per
// encoder format so repeated encoding is nearly free.
//
// If all slots are taken, returns 0 (no caching, graceful degradation —
// everything still works, just without the cache speedup).
func AllocEncoderSlot() int {
	s := int(nextSlot.Add(1))
	if s > maxCacheSlots {
		return 0
	}
	return s
}

// Bag is an immutable, goroutine-safe linked list of Fields — the backbone
// of logf's zero-copy field accumulation. Every call to With or WithGroup
// creates a new node pointing to the parent in O(1) time with no field
// copies. The encoder walks the chain at encoding time, and results are
// cached per encoder so repeated encoding of the same Bag is essentially
// free.
type Bag struct {
	fields []Field
	parent *Bag
	group  string // group name; empty = no group
	cache  atomic.Pointer[bagCache]
}

// NewBag creates a root Bag node with the given fields. Most of the time
// you will not call this directly — Logger.With and logf.With handle Bag
// creation for you.
func NewBag(fs ...Field) *Bag {
	return &Bag{fields: fs}
}

// With returns a new Bag that includes the given additional fields. The
// original Bag is not modified — the new node simply points to the parent.
// O(1) time, zero copies.
func (b *Bag) With(fs ...Field) *Bag {
	return &Bag{fields: fs, parent: b}
}

// WithGroup returns a new Bag that opens a named group. All fields added
// to descendant nodes via subsequent With calls will be logically nested
// under this group name when encoded (e.g., as a nested JSON object).
// The original Bag is not modified.
func (b *Bag) WithGroup(name string) *Bag {
	return &Bag{group: name, parent: b}
}

// Group returns the group name for this Bag node, or an empty string if
// it is a regular field node.
func (b *Bag) Group() string {
	if b == nil {
		return ""
	}
	return b.group
}

// Fields collects all fields across the entire Bag chain in parent-first
// order. This allocates a new slice when the chain has more than one node,
// so for hot-path encoding prefer walking OwnFields + Parent directly.
func (b *Bag) Fields() []Field {
	if b == nil {
		return nil
	}
	if b.parent == nil {
		return b.fields
	}

	// Count total fields.
	n := 0
	for node := b; node != nil; node = node.parent {
		n += len(node.fields)
	}

	// Collect in reverse (child→parent), then reverse to get parent-first.
	all := make([]Field, n)
	i := n
	for node := b; node != nil; node = node.parent {
		i -= len(node.fields)
		copy(all[i:], node.fields)
	}

	return all
}

// OwnFields returns only the fields stored directly in this Bag node,
// without walking up to parents. Useful for cache-aware encoding.
func (b *Bag) OwnFields() []Field {
	if b == nil {
		return nil
	}
	return b.fields
}

// Parent returns the parent Bag in the linked list, or nil if this is
// the root node.
func (b *Bag) Parent() *Bag {
	if b == nil {
		return nil
	}
	return b.parent
}

// LoadCache returns previously cached encoded bytes for the given encoder
// slot, or nil on a cache miss. Slot 0 (no caching) always returns nil.
func (b *Bag) LoadCache(slot int) []byte {
	if slot == 0 || b == nil {
		return nil
	}
	c := b.cache.Load()
	if c == nil {
		return nil
	}
	return c.slots[slot-1]
}

// StoreCache saves encoded bytes for the given encoder slot so future
// Encode calls can skip re-encoding this Bag. Slot 0 (no caching) is a
// no-op. The internal cache structure is allocated lazily on first store.
func (b *Bag) StoreCache(slot int, data []byte) {
	if slot == 0 || b == nil {
		return
	}
	c := b.cache.Load()
	if c == nil {
		c = &bagCache{}
		if !b.cache.CompareAndSwap(nil, c) {
			c = b.cache.Load()
		}
	}
	c.slots[slot-1] = data
}

// HasField reports whether any node in the Bag chain contains a field with
// the given key. Walks the full chain from this node up to the root.
func (b *Bag) HasField(key string) bool {
	for node := b; node != nil; node = node.parent {
		for i := range node.fields {
			if node.fields[i].Key == key {
				return true
			}
		}
	}

	return false
}

type bagKey struct{}

// ContextWithBag returns a new context carrying the given Bag. This is
// the low-level API — most callers should use logf.With(ctx, fields...)
// instead, which handles Bag creation and chaining automatically.
func ContextWithBag(ctx context.Context, bag *Bag) context.Context {
	return context.WithValue(ctx, bagKey{}, bag)
}

// BagFromContext returns the Bag stored in the context, or nil if none
// was set. Safe to call on any context.
func BagFromContext(ctx context.Context) *Bag {
	bag, _ := ctx.Value(bagKey{}).(*Bag)

	return bag
}
