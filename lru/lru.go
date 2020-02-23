package lru

import (
	"container/list"
)

// Cache is LRU cache, It is not safe for concurrent access 
type Cache struct {
	maxBytes int64   // 最大使用内存
	nBytes 	int64        //  当前已使用内存
	ll *list.List   
	cache  map[string]*list.Element   //双链表对应节点的指针
	// option and executed when an entry is purged
    OnEvicted  func(key string, value Value)   //某条记录被删除时候的回调函数
}

type entry struct {
	key string
	value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

// New is the Constructor of cache
func New(maxBytes int64, OnEvicted func(string,Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		ll: list.New(),
		cache: make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

// Get look ups a key's value
func (c *Cache) Get(key string) (value Value,ok bool) {
	if ele,ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)     // 节点移到队头，表示热度为最新
		kv := ele.Value.(*entry)
		return kv.value, true  
	}
	return
}


// RemoveOldest removes the oldest item
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()    // 找到队尾，删除	
	if ele != nil {
		c.ll.Remove(ele) 
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)   // 删除 该节点的 映射关系
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add adds a value to the cache  or modify
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key];ok {   // 若存在，则更新对应节点的值，并移动到最前方，更新节点内容
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		return
	}
    // 若不存在则添加
	ele := c.ll.PushFront(&entry{key, value})    
	c.cache[key] = ele
	c.nBytes += int64(len(key)) + int64(value.Len())

	// 若内存不够，需要删除掉最老的
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}




