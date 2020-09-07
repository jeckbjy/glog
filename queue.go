package glog

import (
	"sync"
)

var gNodePool = sync.Pool{
	New: func() interface{} {
		return &Node{}
	},
}

// NewNode 创建Node,使用完需要Free
func NewNode(e *Entry) *Node {
	n := gNodePool.Get().(*Node)
	n.entry = e
	return n
}

type Node struct {
	entry *Entry
	next  *Node
}

func (n *Node) Free() {
	gNodePool.Put(n)
}

// Iter Queue迭代器
// 使用方式类似java
// iter := q.Iterator()
// for iter.HasNext() {
//	entry := iter.Next()
//  process entry
// }
type Iter struct {
	node *Node
}

func (i *Iter) HasNext() bool {
	return i.node != nil
}

func (i *Iter) Next() *Entry {
	e := i.node.entry
	i.node = i.node.next
	return e
}

// Queue 单向非循环队列,非线程安全
type Queue struct {
	head *Node
	tail *Node
	size int
}

func (q *Queue) Len() int {
	return q.size
}

func (q *Queue) Empty() bool {
	return q.size == 0
}

func (q *Queue) Clear() {
	// free all
	for n := q.head; n != nil; {
		t := n
		n = n.next
		t.entry.Free()
		t.Free()
	}

	q.head = nil
	q.tail = nil
	q.size = 0
}

func (q *Queue) Iterator() Iter {
	return Iter{node: q.head}
}

func (q *Queue) Push(e *Entry) {
	e.Obtain()
	n := NewNode(e)
	if q.size == 0 {
		q.head = n
		q.tail = n
	} else {
		q.tail.next = n
		q.tail = n
	}
	q.size++
}

func (q *Queue) Pop() *Entry {
	if q.size == 0 {
		return nil
	}

	node := q.head

	entry := node.entry
	q.head = node.next
	q.size--
	if q.size == 0 {
		q.head = nil
		q.tail = nil
	}

	node.Free()
	return entry
}
