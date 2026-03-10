package logf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBag(t *testing.T) {
	bag := NewBag(String("k1", "v1"), Int("k2", 42))

	assert.Equal(t, 2, len(bag.Fields()))
	assert.Equal(t, "k1", bag.Fields()[0].Key)
	assert.Equal(t, "k2", bag.Fields()[1].Key)
}

func TestNewBagEmpty(t *testing.T) {
	bag := NewBag()

	assert.Equal(t, 0, len(bag.Fields()))
}

func TestBagWith(t *testing.T) {
	bag1 := NewBag(String("k1", "v1"))
	bag2 := bag1.With(String("k2", "v2"))

	// Original unchanged.
	assert.Equal(t, 1, len(bag1.Fields()))

	// New bag has both fields.
	assert.Equal(t, 2, len(bag2.Fields()))
	assert.Equal(t, "k1", bag2.Fields()[0].Key)
	assert.Equal(t, "k2", bag2.Fields()[1].Key)

}

func TestBagWithNil(t *testing.T) {
	var bag *Bag
	bag2 := bag.With(String("k1", "v1"))

	require.NotNil(t, bag2)
	assert.Equal(t, 1, len(bag2.Fields()))
	assert.Equal(t, "k1", bag2.Fields()[0].Key)
}

func TestBagFieldsNil(t *testing.T) {
	var bag *Bag
	assert.Nil(t, bag.Fields())
}

func TestBagHasField(t *testing.T) {
	bag := NewBag(String("request_id", "abc"), Int("user_id", 42))

	assert.True(t, bag.HasField("request_id"))
	assert.True(t, bag.HasField("user_id"))
	assert.False(t, bag.HasField("missing"))
}

func TestBagHasFieldNil(t *testing.T) {
	var bag *Bag
	assert.False(t, bag.HasField("any"))
}

func TestBagImmutable(t *testing.T) {
	bag1 := NewBag(String("k1", "v1"))
	_ = bag1.With(String("k2", "v2"))

	// bag1 must not be affected.
	assert.Equal(t, 1, len(bag1.Fields()))
}

func TestBagLinkedList(t *testing.T) {
	bag1 := NewBag(String("a", "1"))
	bag2 := bag1.With(String("b", "2"))
	bag3 := bag2.With(String("c", "3"))

	// Each node only has its own fields.
	assert.Equal(t, 1, len(bag1.Fields()))
	assert.Equal(t, 2, len(bag2.Fields()))
	assert.Equal(t, 3, len(bag3.Fields()))

	// Parent-first order.
	assert.Equal(t, "a", bag3.Fields()[0].Key)
	assert.Equal(t, "b", bag3.Fields()[1].Key)
	assert.Equal(t, "c", bag3.Fields()[2].Key)
}

func TestBagHasFieldLinkedList(t *testing.T) {
	bag := NewBag(String("a", "1")).With(String("b", "2")).With(String("c", "3"))

	assert.True(t, bag.HasField("a"))
	assert.True(t, bag.HasField("b"))
	assert.True(t, bag.HasField("c"))
	assert.False(t, bag.HasField("d"))
}

func TestContextWithBag(t *testing.T) {
	bag := NewBag(String("k", "v"))
	ctx := ContextWithBag(context.Background(), bag)

	got := BagFromContext(ctx)
	assert.Equal(t, bag, got)
}

func TestBagFromContextNil(t *testing.T) {
	got := BagFromContext(context.Background())
	assert.Nil(t, got)
}

func TestBagWithGroup(t *testing.T) {
	bag := NewBag(String("a", "1")).WithGroup("http").With(String("method", "GET"))

	assert.Equal(t, "", bag.Group())
	assert.Equal(t, "http", bag.Parent().Group())
	assert.Equal(t, "", bag.Parent().Parent().Group())

	// OwnFields: only the child node has fields.
	assert.Equal(t, 1, len(bag.OwnFields()))
	assert.Equal(t, "method", bag.OwnFields()[0].Key)
}

func TestBagWithGroupNil(t *testing.T) {
	var bag *Bag
	bag2 := bag.WithGroup("g")

	require.NotNil(t, bag2)
	assert.Equal(t, "g", bag2.Group())
	assert.Nil(t, bag2.Parent())
}

func TestBagGroupNil(t *testing.T) {
	var bag *Bag
	assert.Equal(t, "", bag.Group())
}
