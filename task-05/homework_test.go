package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Map[T any, R any](data []T, action func(T) R) []R {
	if data == nil {
		return nil
	}

	result := make([]R, len(data))
	for i, v := range data {
		result[i] = action(v)
	}

	return result
}

func Filter[T any](data []T, action func(T) bool) []T {
	if data == nil {
		return nil
	}

	result := make([]T, 0)
	for _, v := range data {
		if action(v) {
			result = append(result, v)
		}
	}

	return result
}

func Reduce[T any, R any](data []T, initial R, action func(R, T) R) R {
	result := initial
	for _, v := range data {
		result = action(result, v)
	}

	return result
}

func TestMap(t *testing.T) {
	t.Run("int to int", makeMapTests(map[string]struct {
		data   []int
		action func(int) int
		result []int
	}{
		"nil numbers": {
			action: func(number int) int {
				return -number
			},
		},
		"empty numbers": {
			data: []int{},
			action: func(number int) int {
				return -number
			},
			result: []int{},
		},
		"inc numbers": {
			data: []int{1, 2, 3, 4, 5},
			action: func(number int) int {
				return number + 1
			},
			result: []int{2, 3, 4, 5, 6},
		},
		"double numbers": {
			data: []int{1, 2, 3, 4, 5},
			action: func(number int) int {
				return number * number
			},
			result: []int{1, 4, 9, 16, 25},
		},
	}))

	t.Run("float32 to int", makeMapTests(map[string]struct {
		data   []float32
		action func(float32) int
		result []int
	}{
		"rubles to kopecks": {
			data: []float32{1.0, 5.5, 10.99, 0.01},
			action: func(rub float32) int {
				return int(rub * 100)
			},
			result: []int{100, 550, 1099, 1},
		},
	}))

	t.Run("string to int", makeMapTests(map[string]struct {
		data   []string
		action func(string) int
		result []int
	}{
		"lengths": {
			data: []string{"go", "lang"},
			action: func(s string) int {
				return len(s)
			},
			result: []int{2, 4},
		},
	}))
}

func TestFilter(t *testing.T) {
	t.Run("int", makeFilterTests(map[string]struct {
		data   []int
		action func(int) bool
		result []int
	}{
		"nil numbers": {
			action: func(number int) bool {
				return number == 0
			},
		},
		"empty numbers": {
			data: []int{},
			action: func(number int) bool {
				return number == 1
			},
			result: []int{},
		},
		"even numbers": {
			data: []int{1, 2, 3, 4, 5},
			action: func(number int) bool {
				return number%2 == 0
			},
			result: []int{2, 4},
		},
		"positive numbers": {
			data: []int{-1, -2, 1, 2},
			action: func(number int) bool {
				return number > 0
			},
			result: []int{1, 2},
		},
	}))

	t.Run("string", makeFilterTests(map[string]struct {
		data   []string
		action func(string) bool
		result []string
	}{
		"non-empty strings": {
			data: []string{"", "go", "", "lang", ""},
			action: func(s string) bool {
				return s != ""
			},
			result: []string{"go", "lang"},
		},
	}))
}

func TestReduce(t *testing.T) {
	t.Run("int data", makeReduceTests(map[string]struct {
		initial int
		data    []int
		action  func(int, int) int
		result  int
	}{
		"nil numbers": {
			action: func(lhs, rhs int) int {
				return 0
			},
		},
		"empty numbers": {
			data: []int{},
			action: func(lhs, rhs int) int {
				return 0
			},
		},
		"sum of numbers": {
			data: []int{1, 2, 3, 4, 5},
			action: func(lhs, rhs int) int {
				return lhs + rhs
			},
			result: 15,
		},
		"sum of numbers with initial value": {
			initial: 10,
			data:    []int{1, 2, 3, 4, 5},
			action: func(lhs, rhs int) int {
				return lhs + rhs
			},
			result: 25,
		},
	}))

	t.Run("string data", makeReduceTests(map[string]struct {
		initial int
		data    []string
		action  func(int, string) int
		result  int
	}{
		"sum of lengths": {
			initial: 0,
			data:    []string{"go", "lang"},
			action: func(l int, s string) int {
				return l + len(s)
			},
			result: 6,
		},
	}))
}

func makeMapTests[T any, R any](tests map[string]struct {
	data   []T
	action func(T) R
	result []R
}) func(t *testing.T) {
	return func(t *testing.T) {
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				result := Map(test.data, test.action)
				assert.Equal(t, test.result, result)
			})
		}
	}
}

func makeFilterTests[T any](tests map[string]struct {
	data   []T
	action func(T) bool
	result []T
}) func(t *testing.T) {
	return func(t *testing.T) {
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				result := Filter(test.data, test.action)
				assert.Equal(t, test.result, result)
			})
		}
	}
}

func makeReduceTests[T any, R any](tests map[string]struct {
	initial R
	data    []T
	action  func(R, T) R
	result  R
}) func(t *testing.T) {
	return func(t *testing.T) {
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				result := Reduce(test.data, test.initial, test.action)
				assert.Equal(t, test.result, result)
			})
		}
	}
}
