package main

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

func Serialize(input any) string {
	const (
		tagKey          = "properties"
		tagValOmitempty = "omitempty"
	)

	t := reflect.TypeOf(input)
	if t.Kind() != reflect.Struct {
		panic("Serialize: expected struct type")
	}

	v := reflect.ValueOf(input)
	var sb strings.Builder

	for i := 0; i < t.NumField(); i++ {
		var key string
		var omitEmpty bool

		field := t.Field(i)
		tag := field.Tag.Get(tagKey)

		if tag == "" {
			key = field.Name
		} else {
			parts := strings.Split(tag, ",")
			key = parts[0]
			omitEmpty = slices.Contains(parts[1:], tagValOmitempty)
		}

		value := v.Field(i)

		if omitEmpty && isZero(value) {
			continue
		}

		valStr := formatValue(value)

		sb.WriteString(key)
		sb.WriteByte('=')
		sb.WriteString(valStr)
		sb.WriteByte('\n')
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Bool:
		return !v.Bool()
	default:
		panic("Serialize: isZero: unsupported kind " + v.Kind().String())
	}
}

// See encoding/json.newTypeEncoder
func formatValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	default:
		panic("Serialize: formatValue: unsupported kind " + v.Kind().String())
	}
}

func TestSerializationPerson(t *testing.T) {
	type Person struct {
		Name    string `properties:"name"`
		Address string `properties:"address,omitempty"`
		Age     int    `properties:"age"`
		Married bool   `properties:"married"`
	}

	tests := map[string]struct {
		person Person
		result string
	}{
		"test case with empty fields": {
			result: "name=\nage=0\nmarried=false",
		},
		"test case with fields": {
			person: Person{
				Name:    "John Doe",
				Age:     30,
				Married: true,
			},
			result: "name=John Doe\nage=30\nmarried=true",
		},
		"test case with omitempty field": {
			person: Person{
				Name:    "John Doe",
				Age:     30,
				Married: true,
				Address: "Paris",
			},
			result: "name=John Doe\naddress=Paris\nage=30\nmarried=true",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := Serialize(test.person)
			assert.Equal(t, test.result, result)
		})
	}
}

func TestSerializationAny(t *testing.T) {
	tests := map[string]struct {
		input  any
		result string
	}{
		"empty struct": {
			input:  struct{}{},
			result: "",
		},
		"struct with string and int": {
			input: struct {
				Foo string `properties:"foo,omitempty"`
				Bar int    `properties:"bar,omitempty"`
			}{
				Foo: "go",
				Bar: 1,
			},
			result: "foo=go\nbar=1",
		},
		"struct with omitempty and empty field": {
			input: struct {
				Name  string `properties:"mynameis"`
				Email string `properties:"email,omitempty"`
			}{
				Name: "go",
			},
			result: "mynameis=go",
		},
		"struct with no tags": {
			input: struct {
				Name   string
				Active bool
			}{
				Name: "go",
			},
			result: "Name=go\nActive=false",
		},
		"struct with mixed tags and no tags": {
			input: struct {
				ID    int `properties:"id"`
				Title string
				Skip  bool `properties:"ignored,omitempty"`
			}{
				ID:    1,
				Title: "Untitled",
			},
			result: "id=1\nTitle=Untitled",
		},
		"struct with newline-only string": {
			input: struct {
				Body string `properties:"body,omitempty"`
			}{
				Body: "\n\n\n",
			},
			result: "body=\n\n\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := Serialize(test.input)
			assert.Equal(t, test.result, result)
		})
	}
}
