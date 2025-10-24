package main

import (
	"math/rand"
	"reflect"

	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/constraints"
)

// go test -v homework_test.go

type node[K constraints.Ordered, V any] struct {
	key   K
	value V
	left  *node[K, V]
	right *node[K, V]
}

// OrderedMap is not thread-safe.
type OrderedMap[K constraints.Ordered, V any] struct {
	root *node[K, V]
	size int
}

func NewOrderedMap[K constraints.Ordered, V any]() OrderedMap[K, V] {
	return OrderedMap[K, V]{}
}

func (m *OrderedMap[K, V]) Insert(key K, value V) {
	nodeRef := m.findNodeRef(key)
	if *nodeRef == nil {
		*nodeRef = &node[K, V]{key: key, value: value}
		m.size++
	} else {
		(*nodeRef).value = value
	}
}

func (m *OrderedMap[K, V]) Erase(key K) {
	nodeRef := m.findNodeRef(key)
	if *nodeRef == nil {
		return
	}

	if (*nodeRef).left == nil {
		*nodeRef = (*nodeRef).right
	} else if (*nodeRef).right == nil {
		*nodeRef = (*nodeRef).left
	} else {
		minLeftNodeRef := &(*nodeRef).right
		for (*minLeftNodeRef).left != nil {
			minLeftNodeRef = &(*minLeftNodeRef).left
		}

		(*nodeRef).key = (*minLeftNodeRef).key
		(*nodeRef).value = (*minLeftNodeRef).value

		*minLeftNodeRef = (*minLeftNodeRef).right
	}

	m.size--
}

func (m *OrderedMap[K, V]) Contains(key K) bool {
	return *m.findNodeRef(key) != nil
}

func (m *OrderedMap[K, V]) findNodeRef(key K) **node[K, V] {
	n := &m.root

	for *n != nil {
		if key < (*n).key {
			n = &(*n).left
		} else if key > (*n).key {
			n = &(*n).right
		} else {
			break
		}
	}

	return n
}

func (m *OrderedMap[K, V]) Size() int {
	return m.size
}

func (m *OrderedMap[K, V]) ForEach(action func(K, V)) {
	const defaultStackCap = 32

	stack := make([]*node[K, V], 0, min(defaultStackCap, m.size))
	n := m.root

	for n != nil || len(stack) > 0 {
		for n != nil {
			stack = append(stack, n)
			n = n.left
		}

		n = stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		action(n.key, n.value)
		n = n.right
	}
}

func TestCircularQueue(t *testing.T) {
	data := NewOrderedMap[int, int]()
	assert.Zero(t, data.Size())

	data.Insert(10, 10)
	data.Insert(5, 5)
	data.Insert(15, 15)
	data.Insert(2, 2)
	data.Insert(4, 4)
	data.Insert(12, 12)
	data.Insert(14, 14)

	assert.Equal(t, 7, data.Size())
	assert.True(t, data.Contains(4))
	assert.True(t, data.Contains(12))
	assert.False(t, data.Contains(3))
	assert.False(t, data.Contains(13))

	var keys []int
	expectedKeys := []int{2, 4, 5, 10, 12, 14, 15}
	data.ForEach(func(key, _ int) {
		keys = append(keys, key)
	})

	assert.True(t, reflect.DeepEqual(expectedKeys, keys))

	data.Erase(15)
	data.Erase(14)
	data.Erase(2)

	assert.Equal(t, 4, data.Size())
	assert.True(t, data.Contains(4))
	assert.True(t, data.Contains(12))
	assert.False(t, data.Contains(2))
	assert.False(t, data.Contains(14))

	keys = nil
	expectedKeys = []int{4, 5, 10, 12}
	data.ForEach(func(key, _ int) {
		keys = append(keys, key)
	})

	assert.True(t, reflect.DeepEqual(expectedKeys, keys))
}

// go test -v homework_test.go -fuzz=FuzzOrderedMap
func FuzzOrderedMap(f *testing.F) {
	const eraseLen = 3
	const minLen = 5
	const maxLen = 500
	const keySpaceSize = 10000

	check := func(t *testing.T, m *OrderedMap[int, struct{}], expectedLen int) {
		if size := m.Size(); size != expectedLen {
			t.Errorf("expected %d unique keys, got %d from Size", expectedLen, size)
		}

		var count int
		var prev *int

		m.ForEach(func(k int, _ struct{}) {
			if prev != nil && *prev >= k {
				t.Errorf("keys not strictly increasing: prev=%d, curr=%d", *prev, k)
			}

			prev = &k
			count++
		})

		if count != expectedLen {
			t.Errorf("expected %d unique keys, got %d from ForEach", expectedLen, count)
		}
	}

	f.Add(int64(0), int64(0))

	f.Fuzz(func(t *testing.T, seedLen int64, seedVal int64) {
		rndLen := rand.New(rand.NewSource(seedLen))
		rndVal := rand.New(rand.NewSource(seedVal))

		m := NewOrderedMap[int, struct{}]()
		needLen := rndLen.Intn(maxLen-minLen+1) + minLen

		seen := make(map[int]struct{})
		for len(seen) < needLen {
			k := rndVal.Intn(keySpaceSize)
			m.Insert(k, struct{}{})
			seen[k] = struct{}{}
		}

		check(t, &m, needLen)

		{
			i := 0
			for key := range seen {
				i++
				m.Erase(key)

				if i >= eraseLen {
					break
				}
			}
		}

		seen = nil

		check(t, &m, needLen-eraseLen)
	})
}
