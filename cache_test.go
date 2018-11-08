package logf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	goldenFirst  = []byte("first")
	goldenSecond = []byte("second")
	goldenThird  = []byte("third")
)

func TestCacheNew(t *testing.T) {
	c := NewCache(2)
	require.EqualValues(t, 0, c.Len())
}

func TestCacheCleanEmpty(t *testing.T) {
	c := NewCache(2)

	c.Clean()
	require.EqualValues(t, 0, c.Len())
}

func TestCacheGetAbsentBuffer(t *testing.T) {
	c := NewCache(2)

	_, ok := c.Get(0)
	require.False(t, ok)
}

func TestCacheSetOneBuffer(t *testing.T) {
	c := NewCache(2)

	c.Set(0, goldenFirst)
	require.EqualValues(t, 1, c.Len())

	_, ok := c.Get(1)
	require.False(t, ok)

	buf, ok := c.Get(0)
	require.True(t, ok)
	require.EqualValues(t, goldenFirst, buf)

	// Try to get the buffer for the second time.
	buf, ok = c.Get(0)
	require.True(t, ok)
	require.EqualValues(t, goldenFirst, buf)
}

func TestCacheDefaultInsertionOrder(t *testing.T) {
	c := NewCache(2)

	c.Set(0, goldenFirst)
	require.EqualValues(t, 1, c.Len())
	c.Set(1, goldenSecond)
	require.EqualValues(t, 2, c.Len())
	c.Set(2, goldenThird)
	require.EqualValues(t, 2, c.Len())

	// The buffer with ID 0 should be removed from the cache.
	_, ok := c.Get(0)
	require.False(t, ok)

	// Others should exist.
	buf, ok := c.Get(1)
	require.True(t, ok)
	require.EqualValues(t, goldenSecond, buf)

	buf, ok = c.Get(2)
	require.True(t, ok)
	require.EqualValues(t, goldenThird, buf)
}

func TestCacheGetChangesOrder(t *testing.T) {
	c := NewCache(2)

	// Fill the cache.
	c.Set(0, goldenFirst)
	require.EqualValues(t, 1, c.Len())
	c.Set(1, goldenSecond)
	require.EqualValues(t, 2, c.Len())

	// Change LRU order. 0 is LRU now.
	buf, ok := c.Get(0)
	require.True(t, ok)
	require.EqualValues(t, goldenFirst, buf)

	// Add a new one buffer. Last buffer should be removed.
	c.Set(2, goldenThird)
	require.EqualValues(t, 2, c.Len())

	// The buffer with ID 1 should be removed from the cache.
	_, ok = c.Get(1)
	require.False(t, ok)

	// Others should exist.
	buf, ok = c.Get(0)
	require.True(t, ok)
	require.EqualValues(t, goldenFirst, buf)

	buf, ok = c.Get(2)
	require.True(t, ok)
	require.EqualValues(t, goldenThird, buf)
}

func TestCacheSetOneBufferTwoTimes(t *testing.T) {
	c := NewCache(2)

	// Add one buffer.
	c.Set(0, goldenFirst)
	require.EqualValues(t, 1, c.Len())

	// Check it.
	buf, ok := c.Get(0)
	require.True(t, ok)
	require.EqualValues(t, goldenFirst, buf)

	// Set another buffer with the same ID.
	c.Set(0, goldenSecond)
	require.EqualValues(t, 1, c.Len())

	// Try to get the buffer for the second time.
	buf, ok = c.Get(0)
	require.True(t, ok)
	require.EqualValues(t, goldenSecond, buf)
}

func BenchmarkAddWithRotation(b *testing.B) {
	c := NewCache(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(int32(i), goldenFirst)
	}
}

func BenchmarkGetWith1000(b *testing.B) {
	count := 1000
	c := NewCache(count)

	for i := 0; i < count; i++ {
		c.Set(int32(i), goldenFirst)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(int32(i % count))
	}
}
