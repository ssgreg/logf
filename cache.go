package logf

import (
	"container/list"
)

type element struct {
	key   int32
	bytes []byte
}

type Cache struct {
	m map[int32]*list.Element
	l *list.List

	limit int
}

const (
	Unlimited int = 0
)

func NewCache(limit int) *Cache {
	return &Cache{
		m:     make(map[int32]*list.Element, limit),
		l:     list.New(),
		limit: limit,
	}
}

func (c *Cache) Set(k int32, bytes []byte) {
	e := c.l.PushFront(&element{k, bytes})
	c.m[k] = e
	if c.limit != Unlimited {
		if c.l.Len() > c.limit {
			c.removeBack()
		}
	}
}

func (c *Cache) removeBack() {
	e := c.l.Remove(c.l.Back())
	delete(c.m, e.(*element).key)
}

func (c *Cache) Get(k int32) ([]byte, bool) {
	if e, ok := c.m[k]; ok {
		c.l.MoveToFront(e)
		return e.Value.(*element).bytes, true
	}

	return nil, false
}

func (c *Cache) Len() int {
	return len(c.m)
}

func (c *Cache) Clean() {
	c.l = list.New()
	c.m = make(map[int32]*list.Element, c.limit)
}
