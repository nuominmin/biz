package topnheap

import (
	"container/heap"
	"sort"
)

type Item interface {
	Less(other Item) bool
}

type TopNHeap[T Item] interface {
	Size() int
	First() (T, bool)
	SortedDesc() []T
	Add(item T)
}

type heapItem[T Item] struct {
	data []T
	size int
}

func New[T Item](capacity int) TopNHeap[T] {
	return &heapItem[T]{
		data: make([]T, 0, capacity),
		size: capacity,
	}
}

func (h *heapItem[T]) Len() int           { return len(h.data) }
func (h *heapItem[T]) Less(i, j int) bool { return h.data[i].Less(h.data[j]) }
func (h *heapItem[T]) Swap(i, j int)      { h.data[i], h.data[j] = h.data[j], h.data[i] }
func (h *heapItem[T]) Push(x interface{}) {
	item, ok := x.(T)
	if !ok {
		panic("Push: unexpected type")
	}
	h.data = append(h.data, item)
}
func (h *heapItem[T]) Pop() interface{} {
	n := len(h.data)
	item := h.data[n-1]
	h.data = h.data[:n-1]
	return item
}

// 插入元素并保持 topN
func (h *heapItem[T]) Add(item T) {
	if len(h.data) < h.size {
		h.data = append(h.data, item)
		return
	}
	if item.Less(h.data[0]) {
		return
	}
	h.data[0] = item
	heap.Fix(h, 0)
}

// 返回降序排列结果
func (h *heapItem[T]) SortedDesc() []T {
	result := make([]T, len(h.data))
	copy(result, h.data)
	sort.Slice(result, func(i, j int) bool {
		return result[j].Less(result[i])
	})
	return result
}

// 返回第一个元素
func (h *heapItem[T]) First() (T, bool) {
	if len(h.data) == 0 {
		var zero T
		return zero, false
	}
	return h.data[0], true
}

func (h *heapItem[T]) Size() int {
	return len(h.data)
}
