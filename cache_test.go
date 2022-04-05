package simple_groupcache

import (
	"log"
	"sync"
	"testing"
)

func TestCache(t *testing.T) {
	var testDate = []struct {
		op    string
		key   string
		value string
	}{
		{"add", "1", "1"},
		{"get", "1", "1"},
		{"add", "1", "2"},
		{"get", "1", "2"},
	}
	c := cache{
		mu:         sync.Mutex{},
		lru:        nil,
		maxEntries: 0,
	}
	t.Parallel()
	for _, v := range testDate {
		if v.op == "add" {
			c.add(v.key, ByteView{data: []byte(v.value)})
		} else if v.op == "get" {
			val, ok := c.get(v.key)
			if !ok {
				log.Fatalln("get error")
			}
			if string(val.ByteSlice()) != v.value {
				log.Fatalln("value error")
			}
		}
	}
}
