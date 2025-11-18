package main

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/constraints"
)

// go test -v homework_test.go

// NOTE: `pointers` must be sorted in ascending order according to their addresses within `memory`.
func Defragment[T constraints.Unsigned](memory []T, pointers []unsafe.Pointer) {
	if len(memory) == 0 {
		return
	}

	var zero T
	for i, ptr := range pointers {
		dst := unsafe.Pointer(&memory[i])
		if ptr != dst {
			memory[i] = *(*T)(ptr)
			*(*T)(ptr) = zero
			pointers[i] = dst
		}
	}
}

func TestDefragmentation(t *testing.T) {
	var fragmentedMemory = []byte{
		0xF1, 0x00, 0x00, 0x00,
		0x00, 0xF2, 0x00, 0x00,
		0x00, 0x00, 0xF3, 0x00,
		0x00, 0x00, 0x00, 0xF4,
	}

	var fragmentedPointers = []unsafe.Pointer{
		unsafe.Pointer(&fragmentedMemory[0]),
		unsafe.Pointer(&fragmentedMemory[5]),
		unsafe.Pointer(&fragmentedMemory[10]),
		unsafe.Pointer(&fragmentedMemory[15]),
	}

	var defragmentedPointers = []unsafe.Pointer{
		unsafe.Pointer(&fragmentedMemory[0]),
		unsafe.Pointer(&fragmentedMemory[1]),
		unsafe.Pointer(&fragmentedMemory[2]),
		unsafe.Pointer(&fragmentedMemory[3]),
	}

	var defragmentedMemory = []byte{
		0xF1, 0xF2, 0xF3, 0xF4,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	Defragment(fragmentedMemory, fragmentedPointers)
	assert.Equal(t, defragmentedMemory, fragmentedMemory)
	assert.Equal(t, defragmentedPointers, fragmentedPointers)
}

// go test -v homework_test.go -fuzz=FuzzDefragmentByte
func FuzzDefragmentByte(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x00, 0x00, 0x00})
	f.Add([]byte{0xFF, 0x00, 0xFF})
	f.Add([]byte{0xFF, 0xFF, 0xFF})

	f.Fuzz(func(t *testing.T, mem []byte) {
		pointers := make([]unsafe.Pointer, len(mem))
		expectedPtrs := make([]unsafe.Pointer, len(mem))
		expectedMem := make([]byte, len(mem))

		writeIndex := 0
		for i := range mem {
			if mem[i] != 0 {
				pointers[writeIndex] = unsafe.Pointer(&mem[i])
				expectedPtrs[writeIndex] = unsafe.Pointer(&mem[writeIndex])
				expectedMem[writeIndex] = mem[i]
				writeIndex++
			}
		}

		pointers = pointers[:writeIndex]
		expectedPtrs = expectedPtrs[:writeIndex]

		Defragment(mem, pointers)

		assert.Equal(t, expectedMem, mem)
		assert.Equal(t, expectedPtrs, pointers)
	})
}
