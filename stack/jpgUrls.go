package stack

import "sync"

type (
	Stack struct {
		top    *node
		length int
		lock   sync.Mutex
	}
	node struct {
		value string
		prev  *node
	}
)

// Create a new stack
func New() *Stack {
	return &Stack{nil, 0, sync.Mutex{}}
}

// Return the number of items in the stack
func (this *Stack) Len() int {
	return this.length
}

// View the top item on the stack
func (this *Stack) Peek() string {
	if this.length == 0 {
		return ""
	}
	return this.top.value
}

// Pop the top item of the stack and return it
func (this *Stack) Pop() string {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.length == 0 {
		return ""
	}

	n := this.top
	this.top = n.prev
	this.length--
	return n.value
}

// Push a value onto the top of the stack
func (this *Stack) Push(value string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	n := &node{value, this.top}
	this.top = n
	this.length++
}
