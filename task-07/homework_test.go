package main

import (
	"errors"
	"fmt"
	"sync"
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

type Container struct {
	mu    sync.RWMutex
	types map[string]any
}

func NewContainer() *Container {
	return &Container{
		types: make(map[string]any),
	}
}

func (c *Container) RegisterType(name string, constructor any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.types[name] = constructor
}

func (c *Container) RegisterSingletonType(name string, constructor any) {
	var _constructor any

	if fn, ok := constructor.(func() any); ok {
		var (
			instance any
			once     sync.Once
		)

		_constructor = func() any {
			once.Do(func() {
				instance = fn()
			})

			return instance
		}
	} else {
		_constructor = constructor
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.types[name] = _constructor
}

func (c *Container) Resolve(name string) (any, error) {
	c.mu.RLock()
	raw, ok := c.types[name]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w: %v", ErrNotRegistered, name)
	}

	constructor, ok := raw.(func() any)
	if !ok {
		return nil, fmt.Errorf("%w: %T", ErrInvalidConstructor, raw)
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
}
