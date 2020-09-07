package glog

import "sort"

type KV struct {
	Key   string
	Value string
}

// SortedMap 有序的kv结构,按照key排序,保证map的排序稳定,查询时使用二分查找
// 通常用于Tag列表
type SortedMap struct {
	items []KV
}

// Fill 构建map
func (m *SortedMap) Fill(dict map[string]string) {
	if len(dict) == 0 {
		return
	}
	for k, v := range dict {
		m.items = append(m.items, KV{Key: k, Value: v})
	}

	sort.Slice(m.items, func(i, j int) bool {
		return m.items[i].Key < m.items[j].Key
	})
}

// Get 通过Key查询Value
func (m *SortedMap) Get(k string) (string, bool) {
	index := sort.Search(len(m.items), func(i int) bool {
		return m.items[i].Key >= k
	})

	if index < len(m.items) && m.items[index].Key == k {
		return m.items[index].Value, true
	}

	return "", false
}

func (m *SortedMap) GetAt(index int) (string, string) {
	item := &(m.items[index])
	return item.Key, item.Value
}

func (m *SortedMap) GetKey(idx int) string {
	return m.items[idx].Key
}

func (m *SortedMap) GetValue(idx int) string {
	return m.items[idx].Value
}

func (m *SortedMap) Empty() bool {
	return len(m.items) == 0
}

func (m *SortedMap) Len() int {
	return len(m.items)
}
