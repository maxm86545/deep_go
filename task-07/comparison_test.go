package main

import (
	"fmt"
	"testing"
)

// go test -bench=. comparison_test.go

// ===== ContainerResolveCheck BEGIN

func NewResolveCheck() *ContainerResolveCheck {
	return &ContainerResolveCheck{types: make(map[string]any)}
}

type ContainerResolveCheck struct {
	types map[string]any
}

func (c *ContainerResolveCheck) RegisterType(name string, constructor any) {
	c.types[name] = constructor
}

func (c *ContainerResolveCheck) Resolve(name string) (any, error) {
	raw, ok := c.types[name]
	if !ok {
		return nil, fmt.Errorf("not registered")
	}
	fn, ok := raw.(func() any)
	if !ok {
		return nil, fmt.Errorf("invalid constructor")
	}
	return fn(), nil
}

// ===== ContainerResolveCheck END

// ===== ContainerRegisterCheck BEGIN

func NewRegisterCheck() *ContainerRegisterCheck {
	return &ContainerRegisterCheck{types: make(map[string]registration)}
}

type registration struct {
	constructor func() any
	err         error
}

type ContainerRegisterCheck struct {
	types map[string]registration
}

func (c *ContainerRegisterCheck) RegisterType(name string, constructor any) {
	fn, ok := constructor.(func() any)
	if !ok {
		c.types[name] = registration{err: fmt.Errorf("invalid constructor")}
		return
	}
	c.types[name] = registration{constructor: fn}
}

func (c *ContainerRegisterCheck) Resolve(name string) (any, error) {
	reg, ok := c.types[name]
	if !ok {
		return nil, fmt.Errorf("not registered")
	}
	if reg.err != nil {
		return nil, reg.err
	}
	return reg.constructor(), nil
}

// ===== ContainerRegisterCheck END

// ===== ContainerErrorStored BEGIN

func NewErrorStored() *ContainerErrorStored {
	return &ContainerErrorStored{types: make(map[string]any)}
}

type ContainerErrorStored struct {
	types map[string]any
}

func (c *ContainerErrorStored) RegisterType(name string, constructor any) {
	fn, ok := constructor.(func() any)
	if !ok {
		c.types[name] = fmt.Errorf("invalid constructor")
		return
	}
	c.types[name] = fn
}

func (c *ContainerErrorStored) Resolve(name string) (any, error) {
	raw, ok := c.types[name]
	if !ok {
		return nil, fmt.Errorf("not registered")
	}
	fn, ok := raw.(func() any)
	if !ok {
		if err, isErr := raw.(error); isErr {
			return nil, err
		}
		return nil, fmt.Errorf("invalid constructor")
	}
	return fn(), nil
}

// ===== ContainerErrorStored END

var constructorFuncAny = func() any { return 123 }
var constructorInt = 123

const nameType = "test"

func BenchmarkResolveCheck(b *testing.B) {
	b.Run("func() any", func(b *testing.B) {
		c := NewResolveCheck()
		c.RegisterType(nameType, constructorFuncAny)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = c.Resolve(nameType)
		}
	})

	b.Run("int", func(b *testing.B) {
		c := NewResolveCheck()
		c.RegisterType(nameType, constructorInt)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = c.Resolve(nameType)
		}
	})
}

func BenchmarkRegisterCheck(b *testing.B) {
	b.Run("func() any", func(b *testing.B) {
		c := NewRegisterCheck()
		c.RegisterType(nameType, constructorFuncAny)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = c.Resolve(nameType)
		}
	})

	b.Run("int", func(b *testing.B) {
		c := NewRegisterCheck()
		c.RegisterType(nameType, constructorInt)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = c.Resolve(nameType)
		}
	})
}

func BenchmarkErrorStored(b *testing.B) {
	b.Run("func() any", func(b *testing.B) {
		c := NewErrorStored()
		c.RegisterType(nameType, constructorFuncAny)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = c.Resolve(nameType)
		}
	})

	b.Run("int", func(b *testing.B) {
		c := NewErrorStored()
		c.RegisterType(nameType, constructorInt)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = c.Resolve(nameType)
		}
	})
}
