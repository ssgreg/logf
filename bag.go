package logf

import "context"

// Bag is an immutable linked list of Fields.
// Each With creates a new node pointing to the parent — O(1), no copies.
// Bag is safe to share across goroutines.
type Bag struct {
	fields []Field
	parent *Bag
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

// Fields returns all fields in the Bag chain, parent-first order.
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
