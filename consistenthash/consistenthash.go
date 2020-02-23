package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map constans all hashed keys
type Map struct {
	hash 		Hash
	replicas 	int    // 虚拟节点倍数
	keys 		[]int  // Sorted hash环
	hashMap 	map[int]string  // 虚拟节点与真实节点的映射表
}

// New creates a Map instance
func New(replicas int,fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash: fn,
		hashMap: make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) {
	for _,key := range keys {    // 每个key 对应 m.replicas 个虚拟节点，虚拟节点的名称是 strconv.Itoa(i) + key
		for i := 0;i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)   // 加到环上
			m.hashMap[hash] = key   // 虚拟环到真实环上的映射
		}
	}
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))        // 计算key的hash 值
	// Binary search for appropriate replica
	idx := sort.Search(len(m.keys),func(i int) bool {  // 顺时针找到第一个匹配的虚拟节点的下标 idx
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx % len(m.keys)]]   // 通过 hashMap 映射到真实的节点
}




