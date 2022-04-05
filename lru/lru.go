package lru

import "container/list"

type Cache struct {
	// 键值对的最大数量，为0则无限
	MaxEntries int
	// 链表用以维护最近最少使用的信息，最新访问过的数据被移动到队列尾部
	ll *list.List
	// 从key到链表元素的映射
	cache map[string]*list.Element
}

// 键值对
type entry struct {
	key   string
	value interface{}
}

func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[string]*list.Element),
	}
}

func (c *Cache) Get(key string) (value interface{}, ok bool) {
	// key存在，将对应的键值对移动到链表尾部
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) Put(key string, value interface{}) {
	// key已经存在，更新值
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*entry)
		kv.value = value
		c.ll.MoveToBack(ele)
		return
	}
	// key不存在
	// 容量已满删除链表头部元素
	for c.MaxEntries != 0 && len(c.cache) >= c.MaxEntries {
		e := c.ll.Front()
		c.ll.Remove(e)
		delete(c.cache, e.Value.(*entry).key)
	}
	// 插入链表尾部
	e := c.ll.PushBack(&entry{
		key:   key,
		value: value,
	})
	c.cache[key] = e
}
