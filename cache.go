package logf

import (
	"container/list"
)

// Cache is the simple implementation of LRU cache. The Cache is not
// goroutine safe.
type Cache struct {
	m map[int32]*list.Element
	l *list.List

	limit int
}

// NewCache returns a new Cache with the given limit.
func NewCache(limit int) *Cache {
	return &Cache{
		m:     make(map[int32]*list.Element, limit),
		l:     list.New(),
		limit: limit,
	}
}

// Set adds the given buffer with the given key to the cache or replaces
// the existing one with the same key.
func (c *Cache) Set(k int32, bytes []byte) {
	e := c.l.PushFront(&element{k, bytes})
	c.m[k] = e

	if c.l.Len() > c.limit {
		c.removeBack()
	}
}

// Get returns cached buffer for the given key.
func (c *Cache) Get(k int32) ([]byte, bool) {
	if e, ok := c.m[k]; ok {
		c.l.MoveToFront(e)

		return e.Value.(*element).bytes, true
	}

	return nil, false
}

// Len returns the count of cached buffers.
func (c *Cache) Len() int {
	return len(c.m)
}

// Clean removes all buffers from the cache.
func (c *Cache) Clean() {
	c.l = list.New()
	c.m = make(map[int32]*list.Element, c.limit)
}

func (c *Cache) removeBack() {
	e := c.l.Remove(c.l.Back())
	delete(c.m, e.(*element).key)
}

type element struct {
	key   int32
	bytes []byte
}
