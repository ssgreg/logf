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
	assert.NotZero(t, bag.Version())
}

func TestNewBagEmpty(t *testing.T) {
	bag := NewBag()

	assert.Equal(t, 0, len(bag.Fields()))
	assert.NotZero(t, bag.Version())
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

	// Different versions.
	assert.NotEqual(t, bag1.Version(), bag2.Version())
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

func TestBagVersionNil(t *testing.T) {
	var bag *Bag
	assert.Equal(t, int32(0), bag.Version())
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

func TestBagVersionUnique(t *testing.T) {
	bag1 := NewBag(String("a", "1"))
	bag2 := NewBag(String("b", "2"))
	bag3 := bag1.With(String("c", "3"))

	assert.NotEqual(t, bag1.Version(), bag2.Version())
	assert.NotEqual(t, bag1.Version(), bag3.Version())
	assert.NotEqual(t, bag2.Version(), bag3.Version())
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
