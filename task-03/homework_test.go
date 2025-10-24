package main

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

// COWBuffer is not thread-safe.
type COWBuffer struct {
	data []byte
	refs *int
}

func newCOWBuffer(data []byte, refs *int) *COWBuffer {
	b := &COWBuffer{
		data: data,
		refs: refs,
	}

	runtime.SetFinalizer(b, func(b *COWBuffer) {
		b.Close()
	})

	return b
}

func NewCOWBuffer(data []byte) *COWBuffer {
	refs := new(int)
	*refs = 1

	return newCOWBuffer(data, refs)
}

func (b *COWBuffer) Clone() *COWBuffer {
	if b.refs == nil {
		return nil
	}

	*b.refs++

	return newCOWBuffer(b.data, b.refs)
}

func (b *COWBuffer) Close() {
	if b.refs != nil && *b.refs > 0 {
		*b.refs--
	}

	b.refs = nil
	b.data = nil
}

func (b *COWBuffer) Update(index int, value byte) bool {
	if b.refs == nil || len(b.data) == 0 {
		return false
	}

	if index < 0 || index >= len(b.data) {
		return false
	}

	if *b.refs > 1 {
		*b.refs--
		newData := make([]byte, len(b.data))
		copy(newData, b.data)
		*b = *NewCOWBuffer(newData)
	}

	b.data[index] = value

	return true
}

func (b *COWBuffer) String() string {
	return *(*string)(unsafe.Pointer(&b.data))
}

func TestCOWBuffer(t *testing.T) {
	data := []byte{'a', 'b', 'c', 'd'}
	buffer := NewCOWBuffer(data)
	defer buffer.Close()

	copy1 := buffer.Clone()
	copy2 := buffer.Clone()

	assert.NotEqual(t, unsafe.Pointer(&buffer), unsafe.Pointer(&copy1))
	assert.NotEqual(t, unsafe.Pointer(&copy1), unsafe.Pointer(&copy2))

	assert.Equal(t, unsafe.SliceData(data), unsafe.SliceData(buffer.data))
	assert.Equal(t, unsafe.SliceData(buffer.data), unsafe.SliceData(copy1.data))
	assert.Equal(t, unsafe.SliceData(copy1.data), unsafe.SliceData(copy2.data))

	assert.True(t, (*byte)(unsafe.SliceData(data)) == unsafe.StringData(buffer.String()))
	assert.True(t, (*byte)(unsafe.StringData(buffer.String())) == unsafe.StringData(copy1.String()))
	assert.True(t, (*byte)(unsafe.StringData(copy1.String())) == unsafe.StringData(copy2.String()))

	assert.True(t, buffer.Update(0, 'g'))
	assert.False(t, buffer.Update(-1, 'g'))
	assert.False(t, buffer.Update(4, 'g'))

	assert.True(t, reflect.DeepEqual([]byte{'g', 'b', 'c', 'd'}, buffer.data))
	assert.True(t, reflect.DeepEqual([]byte{'a', 'b', 'c', 'd'}, copy1.data))
	assert.True(t, reflect.DeepEqual([]byte{'a', 'b', 'c', 'd'}, copy2.data))

	assert.NotEqual(t, unsafe.SliceData(buffer.data), unsafe.SliceData(copy1.data))
	assert.Equal(t, unsafe.SliceData(copy1.data), unsafe.SliceData(copy2.data))

	copy1.Close()

	previous := copy2.data
	copy2.Update(0, 'f')
	current := copy2.data

	// 1 reference - don't need to copy buffer during update
	assert.Equal(t, unsafe.SliceData(previous), unsafe.SliceData(current))

	copy2.Close()
}

func TestCOWBuffer_CloseClose(t *testing.T) {
	data := []byte{'a', 'b', 'c', 'd'}
	buffer := NewCOWBuffer(data)

	assert.NotNil(t, buffer.refs)
	assert.NotNil(t, buffer.data)

	buffer.Close()

	assert.Nil(t, buffer.refs)
	assert.Nil(t, buffer.data)

	assert.NotPanics(t, func() {
		buffer.Close()
	})

	assert.Nil(t, buffer.refs)
	assert.Nil(t, buffer.data)
}

func TestCOWBuffer_NilSafety(t *testing.T) {
	var b COWBuffer

	assert.Equal(t, "", b.String())

	assert.False(t, b.Update(1, 'a'))
	assert.False(t, b.Update(0, 'b'))
	assert.False(t, b.Update(-1, 'c'))

	assert.NotPanics(t, func() {
		b.Close()
	})

	assert.NotPanics(t, func() {
		assert.Nil(t, b.Clone())
	})
}

func TestCOWBuffer_Finalizer(t *testing.T) {
	runtime.GC()

	var refs *int

	{
		data := []byte{'a', 'b', 'c', 'd'}
		copy1 := NewCOWBuffer(data)
		refs = copy1.refs

		assert.Equal(t, 1, *copy1.refs)

		{
			copy2 := copy1.Clone()

			assert.Equal(t, 2, *copy1.refs)
			assert.Equal(t, 2, *copy2.refs)

			copy3 := copy2.Clone()

			assert.Equal(t, 3, *copy1.refs)
			assert.Equal(t, 3, *copy2.refs)
			assert.Equal(t, 3, *copy3.refs)
		}

		runtime.GC()

		assert.Equal(t, 1, *copy1.refs)
	}

	runtime.GC()

	assert.NotNil(t, refs)
	assert.Equal(t, 0, *refs)
}
