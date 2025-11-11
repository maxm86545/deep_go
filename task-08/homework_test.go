package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

// MultiError is thread-safe.
type MultiError struct {
	errs []error
}

func (e *MultiError) Error() string {
	if len(e.errs) == 0 {
		return ""
	}

	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "%d errors occured:\n", len(e.errs))

	for _, err := range e.errs {
		_, _ = fmt.Fprintf(&b, "\t* %s", err.Error())
	}

	b.WriteString("\n")

	return b.String()
}

func Append(err error, errs ...error) *MultiError {
	var newErrs []error

	if err != nil {
		if me, ok := err.(*MultiError); ok {
			newErrs = append(newErrs, me.errs...)
		} else {
			newErrs = append(newErrs, err)
		}
	}

	for _, err := range errs {
		if err != nil {
			newErrs = append(newErrs, err)
		}
	}

	return &MultiError{errs: newErrs}
}

func (e *MultiError) Unwrap() []error {
	return e.errs
}

func (e *MultiError) Is(target error) bool {
	for _, err := range e.errs {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}

func (e *MultiError) As(target any) bool {
	if e == nil {
		return false
	}

	for _, err := range e.errs {
		if errors.As(err, target) {
			return true
		}
	}

	return false
}

func TestMultiError(t *testing.T) {
	var err error
	err = Append(err, errors.New("error 1"))
	err = Append(err, errors.New("error 2"))

	expectedMessage := "2 errors occured:\n\t* error 1\t* error 2\n"
	assert.EqualError(t, err, expectedMessage)
}

// go test -v homework_test.go -fuzz=FuzzMultiError
func FuzzMultiError(f *testing.F) {
	const maxErrors = 100
	const multiErrorMaxEvery = 10

	f.Add(int64(0), int64(0))

	f.Fuzz(func(t *testing.T, countSeed int64, byteSeed int64) {
		countRand := rand.New(rand.NewSource(countSeed))
		byteRand := rand.New(rand.NewSource(byteSeed))

		count := countRand.Intn(maxErrors) + 1
		multiErrPeriod := countRand.Intn(multiErrorMaxEvery + 1)

		var (
			err           error
			allErrs       []error
			totalBufLen   int
			totalErrCount int
			foreignErr    foreign
		)

		for i := 0; i < count; i++ {
			n := byteRand.Intn(100) + 1
			buf := make([]byte, n)
			_, readErr := byteRand.Read(buf)
			assert.NoError(t, readErr, "rand.Read failed")

			currentErr := customErr{msg: string(buf)}
			allErrs = append(allErrs, currentErr)
			totalBufLen += len(buf)

			if multiErrPeriod != 0 && i%multiErrPeriod == 0 {
				err = Append(&MultiError{errs: allErrs[:i]}, currentErr)
				totalErrCount = i + 1
			} else {
				err = Append(err, currentErr)
				totalErrCount++
			}

			me := err.(*MultiError)

			expectedLen := len(fmt.Sprintf("%d errors occured:\n", totalErrCount)) +
				totalErrCount*len("\t* ") +
				totalBufLen +
				len("\n")

			assert.Equalf(t, expectedLen, len(me.Error()), "Error() length mismatch at i=%d", i)
			assert.Equalf(t, totalErrCount, len(me.Unwrap()), "Unwrap count mismatch at i=%d", i)

			assert.True(t, me.Is(currentErr), "Is failed to match customErr at i=%d", i)
			assert.False(t, me.Is(foreignErr), "Is unexpectedly matched foreign error at i=%d", i)

			var foreignTarget *foreign
			assert.False(t, me.As(&foreignTarget), "As unexpectedly matched foreign type at i=%d", i)

			var target customErr
			assert.True(t, me.As(&target), "As failed to match customErr at i=%d", i)
		}
	})
}

type customErr struct {
	msg string
}

func (e customErr) Error() string {
	return e.msg
}

type foreign struct{}

func (foreign) Error() string {
	return "foreign"
}
