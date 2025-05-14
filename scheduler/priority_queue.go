// scheduler/priority_queue.go
package scheduler

import (
	"container/heap"
	"sync"
)

// TaskItem 表示优先级队列中的任务项
type TaskItem struct {
	task     *Task
	priority Priority
	index    int // 在堆中的索引，由 heap.Interface 维护
}

// PriorityQueue 实现了一个基于优先级的任务队列
type PriorityQueue struct {
	items []*TaskItem
	mutex sync.Mutex
}

// Len 返回队列长度
func (pq *PriorityQueue) Len() int {
	return len(pq.items)
}

// Less 比较两个任务的优先级
// 注意：我们希望 Pop 返回最高优先级的任务，所以使用 > 而不是 <
func (pq *PriorityQueue) Less(i, j int) bool {
	return pq.items[i].priority > pq.items[j].priority
}

// Swap 交换两个任务的位置
func (pq *PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

// Push 添加任务到队列
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item := x.(*TaskItem)
	item.index = n
	pq.items = append(pq.items, item)
}

// Pop 从队列中移除并返回最高优先级的任务
func (pq *PriorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // 避免内存泄漏
	item.index = -1 // 标记为已移除
	pq.items = old[0 : n-1]
	return item
}

// NewPriorityQueue 创建一个新的优先级队列
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		items: make([]*TaskItem, 0),
	}
	heap.Init(pq)
	return pq
}

// Enqueue 将任务添加到队列
func (pq *PriorityQueue) Enqueue(task *Task) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	
	item := &TaskItem{
		task:     task,
		priority: task.priority,
	}
	heap.Push(pq, item)
}

// Dequeue 从队列中取出最高优先级的任务
func (pq *PriorityQueue) Dequeue() *Task {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	
	if pq.Len() == 0 {
		return nil
	}
	
	item := heap.Pop(pq).(*TaskItem)
	return item.task
}

// IsEmpty 检查队列是否为空
func (pq *PriorityQueue) IsEmpty() bool {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	return pq.Len() == 0
}
