package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

var hash = crc32.ChecksumIEEE

type Map struct {
	// 在节点数量比较少的时候，容易出现节点映射不均衡的现象，可以插入多个虚拟节点
	replicas int
	// 虚拟节点hash值对应的实际节点
	hashMap map[int]string
	// 虚拟节点的哈希值，需要排序
	// 通过二分查找更快找到 目标key 所归属的节点
	keys []int
}

func New(replicas int) *Map {
	return &Map{
		replicas: replicas,
		hashMap:  make(map[int]string),
		keys:     []int{},
	}
}

func (m *Map) AddNode(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 拼接编号和节点名称，获得hash值
			h := int(hash([]byte(strconv.Itoa(i) + key)))
			// 插入map和keys中
			m.hashMap[h] = key
			m.keys = append(m.keys, h)
		}
	}
	// 排序
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	h := int(hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= h })
	// 因为是个环，下标等于长度时，就折回到0，归属0号节点
	if idx == len(m.keys) {
		idx = 0
	}
	return m.hashMap[m.keys[idx]]
}
