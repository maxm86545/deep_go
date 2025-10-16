package main

import (
	"golang.org/x/exp/constraints"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

// CircularQueue is not thread-safe.
type CircularQueue[T constraints.Signed] struct {
	values []T
	size   int
	head   int
	tail   int
	count  int
}

func NewCircularQueue[T constraints.Signed](size int) CircularQueue[T] {
	return CircularQueue[T]{
		values: make([]T, size),
		size:   size,
	}
}

func (q *CircularQueue[T]) Push(value T) bool {
	if q.Full() {
		return false
	}

	q.values[q.tail] = value
	q.tail = (q.tail + 1) % q.size
	q.count++

	return true
}

func (q *CircularQueue[T]) Pop() bool {
	if q.Empty() {
		return false
	}

	q.head = (q.head + 1) % q.size
	q.count--

	return true
}

func (q *CircularQueue[T]) Front() T {
	if q.Empty() {
		return q.notFoundValue()
	}

	return q.values[q.head]
}

func (q *CircularQueue[T]) Back() T {
	if q.Empty() {
		return q.notFoundValue()
	}

	return q.values[(q.tail-1+q.size)%q.size]
}

func (q *CircularQueue[T]) Empty() bool {
	return q.count <= 0
}

func (q *CircularQueue[T]) Full() bool {
	return q.count >= q.size
}

func (q *CircularQueue[T]) notFoundValue() T {
	var zero T

	return zero - 1
}

func TestCircularQueueInt(t *testing.T) {
	const queueSize = 3
	queue := NewCircularQueue[int](queueSize)

	assert.True(t, queue.Empty())
	assert.False(t, queue.Full())

	assert.Equal(t, -1, queue.Front())
	assert.Equal(t, -1, queue.Back())
	assert.False(t, queue.Pop())

	assert.True(t, queue.Push(1))
	assert.True(t, queue.Push(2))
	assert.True(t, queue.Push(3))
	assert.False(t, queue.Push(4))

	assert.Equal(t, []int{1, 2, 3}, queue.values)

	assert.False(t, queue.Empty())
	assert.True(t, queue.Full())

	assert.Equal(t, 1, queue.Front())
	assert.Equal(t, 3, queue.Back())

	assert.True(t, queue.Pop())
	assert.False(t, queue.Empty())
	assert.False(t, queue.Full())
	assert.True(t, queue.Push(4))

	assert.Equal(t, []int{4, 2, 3}, queue.values)

	assert.Equal(t, 2, queue.Front())
	assert.Equal(t, 4, queue.Back())

	assert.True(t, queue.Pop())
	assert.True(t, queue.Pop())
	assert.True(t, queue.Pop())
	assert.False(t, queue.Pop())

	assert.True(t, queue.Empty())
	assert.False(t, queue.Full())
}

func TestCircularQueueInt8(t *testing.T) {
	const queueSize = 2
	queue := NewCircularQueue[int8](queueSize)

	assert.True(t, queue.Empty())
	assert.False(t, queue.Full())

	assert.Equal(t, int8(-1), queue.Front())
	assert.Equal(t, int8(-1), queue.Back())
	assert.False(t, queue.Pop())

	assert.True(t, queue.Push(10))
	assert.True(t, queue.Push(20))
	assert.False(t, queue.Push(30))

	assert.Equal(t, int8(10), queue.Front())
	assert.Equal(t, int8(20), queue.Back())

	assert.True(t, queue.Pop())
	assert.True(t, queue.Push(30))

	assert.Equal(t, int8(20), queue.Front())
	assert.Equal(t, int8(30), queue.Back())
}

func TestCircularQueueInt64(t *testing.T) {
	const queueSize = 2
	queue := NewCircularQueue[int64](queueSize)

	assert.True(t, queue.Empty())
	assert.False(t, queue.Full())

	assert.Equal(t, int64(-1), queue.Front())
	assert.Equal(t, int64(-1), queue.Back())
	assert.False(t, queue.Pop())

	assert.True(t, queue.Push(1<<40))
	assert.True(t, queue.Push(1<<41))
	assert.False(t, queue.Push(1<<42))

	assert.Equal(t, int64(1<<40), queue.Front())
	assert.Equal(t, int64(1<<41), queue.Back())

	assert.True(t, queue.Pop())
	assert.True(t, queue.Push(1<<42))

	assert.Equal(t, int64(1<<41), queue.Front())
	assert.Equal(t, int64(1<<42), queue.Back())
}

func TestCircularQueueOneSizeCycle(t *testing.T) {
	const queueSize = 1
	queue := NewCircularQueue[int](queueSize)

	for i := 0; i < 5; i++ {
		assert.True(t, queue.Push(i))
		assert.Equal(t, i, queue.Front())
		assert.Equal(t, i, queue.Back())
		assert.True(t, queue.Full())
		assert.False(t, queue.Empty())
		assert.True(t, queue.Pop())
		assert.Equal(t, -1, queue.Front())
		assert.Equal(t, -1, queue.Back())
		assert.False(t, queue.Full())
		assert.False(t, queue.Pop())
		assert.True(t, queue.Empty())
	}
}
