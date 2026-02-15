// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tmc

import (
	"encoding"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"time"
)

// Adapter wraps a type to preserve type information when encoding and decoding
// a map.
//
// The simples way to create new adapters is via the NewAdapter function.
type Adapter interface {
	FromString(s string) (any, error)
	MarshalText() (text []byte, err error)
	Type() reflect.Type
	Wrap(v any) Adapter
}

// DefaultTypeAdapters contains the default set of type adapters.
var DefaultTypeAdapters = []Adapter{
	// Time
	NewAdapter(time.Now(), nil, nil),
	NewAdapter(
		3*time.Hour,
		func(s string) (any, error) { return time.ParseDuration(s) },
		func(v any) (string, error) { return v.(time.Duration).String(), nil },
	),

	// Numbers
	NewAdapter(big.NewRat(1, 2), nil, nil),
	NewAdapter(
		int(32),
		func(s string) (any, error) {
			return strconv.Atoi(s)
		},
		func(v any) (string, error) {
			return strconv.Itoa(v.(int)), nil
		},
	),
}

// NewAdapter creates a new adapter that wraps the target type.
//
// fromString can be omitted if target implements encoding.TextUnmarshaler.
// toString can be omitted if target implements encoding.TextMarshaler.
//
// It will panic if it can not be created.
func NewAdapter(
	target any,
	fromString func(s string) (any, error),
	toString func(v any) (string, error),
) Adapter {
	targetValue := reflect.ValueOf(target)
	targetType := targetValue.Type()

	wasPointer := targetType.Kind() == reflect.Ptr
	if !wasPointer {
		// Need the pointer to see the TextUnmarshaler implementation.
		v := targetValue
		targetValue = reflect.New(targetType)
		targetValue.Elem().Set(v)
	}

	if fromString == nil {
		if _, ok := targetValue.Interface().(encoding.TextUnmarshaler); ok {
			fromString = func(s string) (any, error) {
				typ := targetType
				if typ.Kind() == reflect.Ptr {
					typ = typ.Elem()
				}
				v := reflect.New(typ)

				err := v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
				if err != nil {
					return nil, err
				}

				if !wasPointer {
					v = v.Elem()
				}
				return v.Interface(), nil
			}
		} else {
			panic(fmt.Sprintf("%T can not be unmarshaled", target))
		}
	}

	var marshalText func(v any) ([]byte, error)

	if toString != nil {
		marshalText = func(v any) ([]byte, error) {
			s, err := toString(v)
			return []byte(s), err
		}
	} else if _, ok := target.(encoding.TextMarshaler); ok {
		marshalText = func(v any) ([]byte, error) {
			return v.(encoding.TextMarshaler).MarshalText()
		}
	} else {
		panic(fmt.Sprintf("%T can not be marshaled", target))
	}

	return &adapter{
		targetType:  targetType,
		fromString:  fromString,
		marshalText: marshalText,
	}
}

var _ Adapter = (*adapter)(nil)

type adapter struct {
	fromString  func(s string) (any, error)
	marshalText func(v any) (text []byte, err error)
	targetType  reflect.Type

	target any
}

func (a *adapter) FromString(s string) (any, error) {
	return a.fromString(s)
}

func (a *adapter) MarshalText() (text []byte, err error) {
	return a.marshalText(a.target)
}

func (a adapter) Type() reflect.Type {
	return a.targetType
}

func (a adapter) Wrap(v any) Adapter {
	a.target = v
	return &a
}
