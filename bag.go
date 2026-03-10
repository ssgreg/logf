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

// AllocEncoderSlot returns a unique slot index for an encoder to use
// with Bag.LoadCache / Bag.StoreCache. Slots are 1-based; zero value
// means "no slot" and cache methods become no-ops.
//
// Call once per encoder at creation time. If all slots are taken,
// returns 0 (no caching, graceful degradation).
func AllocEncoderSlot() int {
	s := int(nextSlot.Add(1))
	if s > maxCacheSlots {
		return 0
	}
	return s
}

// Bag is an immutable linked list of Fields.
// Each With creates a new node pointing to the parent — O(1), no copies.
// Bag is safe to share across goroutines.
type Bag struct {
	fields []Field
	parent *Bag
	group  string // group name; empty = no group
	cache  atomic.Pointer[bagCache]
}

// NewBag creates a new Bag with the given fields.
func NewBag(fs ...Field) *Bag {
	return &Bag{fields: fs}
}

// With returns a new Bag that contains both the existing fields and the
// given additional fields. The original Bag is not modified.
// O(1): no field copy, new node points to parent.
func (b *Bag) With(fs ...Field) *Bag {
	return &Bag{fields: fs, parent: b}
}

// WithGroup returns a new Bag that opens a named group.
// All fields added to descendant nodes (via With) will be logically
// nested under this group when encoded. The original Bag is not modified.
func (b *Bag) WithGroup(name string) *Bag {
	return &Bag{group: name, parent: b}
}

// Group returns the group name for this Bag node, or empty string.
func (b *Bag) Group() string {
	if b == nil {
		return ""
	}
	return b.group
}

// Fields returns all fields in the Bag chain, parent-first order.
// This allocates a new slice when the chain has more than one node.
// For cache-aware encoding, use OwnFields + Parent instead.
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

// OwnFields returns only the fields stored in this Bag node,
// not including parent fields.
func (b *Bag) OwnFields() []Field {
	if b == nil {
		return nil
	}
	return b.fields
}

// Parent returns the parent Bag, or nil for a root node.
func (b *Bag) Parent() *Bag {
	if b == nil {
		return nil
	}
	return b.parent
}

// LoadCache returns cached encoded bytes for the given encoder slot,
// or nil on cache miss. Slot 0 (uninitialized) always returns nil.
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

// StoreCache stores encoded bytes in the cache for the given encoder
// slot. Slot 0 (uninitialized) is a no-op. The bagCache is allocated
// lazily on first store.
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

// HasField reports whether the Bag chain contains a field with the given key.
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

// ContextWithBag returns a new context with the given Bag associated.
func ContextWithBag(ctx context.Context, bag *Bag) context.Context {
	return context.WithValue(ctx, bagKey{}, bag)
}

// BagFromContext returns the Bag associated with the context, or nil.
func BagFromContext(ctx context.Context) *Bag {
	bag, _ := ctx.Value(bagKey{}).(*Bag)

	return bag
}
