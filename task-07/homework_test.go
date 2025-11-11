package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

type UserService struct {
	// not need to implement
	NotEmptyStruct bool
}
type MessageService struct {
	// not need to implement
	NotEmptyStruct bool
}

var (
	ErrNotRegistered      = errors.New("not registered")
	ErrTypeMismatch       = errors.New("type mismatch")
	ErrInvalidConstructor = errors.New("invalid constructor")
)

// Container is not thread-safe.
type Container struct {
	types map[string]any
}

func NewContainer() *Container {
	return &Container{
		types: make(map[string]any),
	}
}

func (c *Container) RegisterType(name string, constructor any) {
	fn, err := getConstructor(constructor)
	if err != nil {
		c.types[name] = err
		return
	}

	c.types[name] = fn
}

func (c *Container) RegisterSingletonType(name string, constructor any) {
	fn, err := getConstructor(constructor)
	if err != nil {
		c.types[name] = err
		return
	}

	var (
		instance any
		once     bool
	)

	singleton := func() any {
		if !once {
			instance = fn()
			once = true
		}

		return instance
	}

	c.types[name] = singleton
}

func (c *Container) Resolve(name string) (any, error) {
	raw, ok := c.types[name]

	if !ok {
		return nil, fmt.Errorf("%w: %v", ErrNotRegistered, name)
	}

	constructor, ok := raw.(func() any)
	if !ok {
		if err, ok := raw.(error); ok {
			return nil, err
		} else {
			panic("invalid constructor")
		}
	}

	return constructor(), nil
}

func ResolveAs[T any](c *Container, name string) (T, error) {
	val, err := c.Resolve(name)
	if err != nil {
		var zero T

		return zero, err
	}

	result, ok := val.(T)
	if !ok {
		var zero T

		return zero, fmt.Errorf("%w: %T", ErrTypeMismatch, val)
	}

	return result, nil
}

func getConstructor(raw any) (func() any, error) {
	fn, ok := raw.(func() any)

	if !ok {
		return nil, fmt.Errorf("%w: %T", ErrInvalidConstructor, raw)
	}

	return fn, nil
}

func TestDIContainer(t *testing.T) {
	container := NewContainer()
	container.RegisterType("UserService", func() any {
		return &UserService{}
	})
	container.RegisterType("MessageService", func() any {
		return &MessageService{}
	})
	container.RegisterSingletonType("UserServiceSingleton", func() any {
		return &UserService{}
	})
	container.RegisterSingletonType("MessageServiceSingleton", func() any {
		return &MessageService{}
	})
	container.RegisterType("Broken", 123)
	container.RegisterSingletonType("BrokenSingleton", false)

	t.Run("Resolve returns registered service", func(t *testing.T) {
		messageService, err := container.Resolve("MessageService")
		assert.NoError(t, err)
		assert.NotNil(t, messageService)
	})

	t.Run("Resolve returns error for unregistered service", func(t *testing.T) {
		paymentService, err := container.Resolve("PaymentService")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotRegistered)
		assert.Nil(t, paymentService)
	})

	t.Run("Resolve returns error for invalid constructor in RegisterType", func(t *testing.T) {
		val, err := container.Resolve("Broken")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidConstructor)
		assert.Nil(t, val)
	})

	t.Run("Resolve returns new instance each time", func(t *testing.T) {
		userService1, err := container.Resolve("UserService")
		assert.NoError(t, err)
		userService2, err := container.Resolve("UserService")
		assert.NoError(t, err)

		u1 := userService1.(*UserService)
		u2 := userService2.(*UserService)
		assert.False(t, u1 == u2)
	})

	t.Run("Resolve returns same instance for singleton", func(t *testing.T) {
		s1, err := container.Resolve("UserServiceSingleton")
		assert.NoError(t, err)
		s2, err := container.Resolve("UserServiceSingleton")
		assert.NoError(t, err)

		assert.True(t, s1 == s2)
	})

	t.Run("ResolveAs returns typed service", func(t *testing.T) {
		messageService, err := ResolveAs[*MessageService](container, "MessageService")
		assert.NoError(t, err)
		assert.NotNil(t, messageService)
	})

	t.Run("ResolveAs returns error for missing service", func(t *testing.T) {
		messageService, err := ResolveAs[*MessageService](container, "MessageService2")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotRegistered)
		assert.Nil(t, messageService)
	})

	t.Run("ResolveAs returns error for invalid constructor in RegisterSingletonType", func(t *testing.T) {
		val, err := ResolveAs[*MessageService](container, "BrokenSingleton")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidConstructor)
		assert.Nil(t, val)
	})

	t.Run("ResolveAs returns type mismatch error", func(t *testing.T) {
		messageService, err := ResolveAs[*MessageService](container, "UserService")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrTypeMismatch)
		assert.Nil(t, messageService)
	})

	t.Run("ResolveAs returns new instance each time", func(t *testing.T) {
		u1, err := ResolveAs[*UserService](container, "UserService")
		assert.NoError(t, err)
		u2, err := ResolveAs[*UserService](container, "UserService")
		assert.NoError(t, err)

		assert.False(t, u1 == u2)
	})

	t.Run("ResolveAs returns same typed instance for singleton", func(t *testing.T) {
		m1, err := ResolveAs[*MessageService](container, "MessageServiceSingleton")
		assert.NoError(t, err)
		m2, err := ResolveAs[*MessageService](container, "MessageServiceSingleton")
		assert.NoError(t, err)

		assert.True(t, m1 == m2)
	})

	t.Run("RegisterType overrides previous registration", func(t *testing.T) {
		container.RegisterType("ServiceRewrite", func() any {
			return 1
		})
		container.RegisterType("ServiceRewrite", func() any {
			return 2
		})

		val, err := container.Resolve("ServiceRewrite")
		assert.NoError(t, err)
		assert.Equal(t, 2, val)
	})

	t.Run("RegisterSingletonType overrides previous registration", func(t *testing.T) {
		container.RegisterSingletonType("ServiceSingletonRewrite", func() any {
			return 1
		})
		container.RegisterSingletonType("ServiceSingletonRewrite", func() any {
			return 2
		})

		val, err := container.Resolve("ServiceSingletonRewrite")
		assert.NoError(t, err)
		assert.Equal(t, 2, val)
	})
}
