// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package typedmapcodec

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// JSONMarshaler encodes and decodes JSON and is the default used in this
// codec.
var JSONMarshaler = new(jsonMarshaler)

// New creates a new Coded with some optional options.
func New(opts ...Option) (*Codec, error) {
	c := &Codec{
		typeSep:               "|",
		marshaler:             JSONMarshaler,
		typeAdapters:          DefaultTypeAdapters,
		typeAdaptersMap:       make(map[reflect.Type]Adapter),
		typeAdaptersStringMap: make(map[string]Adapter),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return c, err
		}
	}

	for _, w := range c.typeAdapters {
		tp := w.Type()
		c.typeAdaptersMap[tp] = w
		c.typeAdaptersStringMap[tp.String()] = w
	}

	return c, nil
}

// Option configures the Codec.
type Option func(c *Codec) error

// WithTypeSep sets the separator to use before the type information encoded in
// the key field. Default is "|".
func WithTypeSep(sep string) func(c *Codec) error {
	return func(c *Codec) error {
		if sep == "" {
			return errors.New("separator cannot be empty")
		}
		c.typeSep = sep
		return nil
	}
}

// WithMarshalUnmarshaler sets the MarshalUnmarshaler to use.
// Default is JSONMarshaler.
func WithMarshalUnmarshaler(marshaler MarshalUnmarshaler) func(c *Codec) error {
	return func(c *Codec) error {
		c.marshaler = marshaler
		return nil
	}
}

// WithTypeAdapters sets the type adapters to use. Note that if more than one
// adapter exists for the same type, the last one will win. This means that
// if you want to use the default adapters, but override some of them, you
// can do:
//
//   adapters := append(typedmapcodec.DefaultTypeAdapters, mycustomAdapters ...)
//   codec := typedmapcodec.New(WithTypeAdapters(adapters))
//
func WithTypeAdapters(typeAdapters []Adapter) func(c *Codec) error {
	return func(c *Codec) error {
		c.typeAdapters = typeAdapters
		return nil
	}
}

// Codec provides methods to marshal and unmarshal a Go map while preserving
// type information.
type Codec struct {
	typeSep               string
	marshaler             MarshalUnmarshaler
	typeAdapters          []Adapter
	typeAdaptersMap       map[reflect.Type]Adapter
	typeAdaptersStringMap map[string]Adapter
}

// Marshal accepts a Go map and marshals it to the configured marshaler
// anntated with type information.
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	m, err := c.toTypedMap(v)
	if err != nil {
		return nil, err
	}
	return c.marshaler.Marshal(m)
}

// Unmarshal unmarshals the given data to the given Go map, using
// any annotated type information found to preserve the type information
// stored in Marshal.
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	if err := c.marshaler.Unmarshal(data, v); err != nil {
		return err
	}
	return c.fromTypedMap(v)
}

func (c *Codec) newKey(key reflect.Value, a Adapter) reflect.Value {
	return reflect.ValueOf(fmt.Sprintf("%s%s%s", key, c.typeSep, a.Type()))
}

func (c *Codec) fromTypedMap(v interface{}) error {
	mv := reflect.ValueOf(v)
	if mv.Kind() == reflect.Ptr {
		mv = mv.Elem()
	}

	if mv.Kind() != reflect.Map {
		return errors.New("must be a Map")
	}

	for _, key := range mv.MapKeys() {
		v := indirectInterface(mv.MapIndex(key))

		if v.Type().Kind() == reflect.Map {
			if err := c.fromTypedMap(v.Interface()); err != nil {
				return nil
			}
			continue
		}

		if key.Kind() != reflect.String {
			continue
		}

		keyStr := key.String()

		sepIdx := strings.LastIndex(keyStr, c.typeSep)
		if sepIdx == -1 {
			continue
		}

		keyPlain := keyStr[:sepIdx]
		keyType := keyStr[sepIdx+len(c.typeSep):]

		if wrapper, found := c.typeAdaptersStringMap[keyType]; found {
			nv, err := wrapper.FromString(v.String())
			if err != nil {
				return err
			}
			mv.SetMapIndex(reflect.ValueOf(keyPlain), reflect.ValueOf(nv))
			mv.SetMapIndex(key, reflect.Value{})
		}

	}

	return nil
}

func (c *Codec) toTypedMap(v interface{}) (interface{}, error) {
	mv := reflect.ValueOf(v)
	if mv.Kind() != reflect.Map {
		return nil, errors.New("must be a Map")
	}

	mcopy := reflect.MakeMap(mv.Type())

	for _, key := range mv.MapKeys() {
		v := indirectInterface(mv.MapIndex(key))

		if v.Type().Kind() == reflect.Map {
			nested, err := c.toTypedMap(v.Interface())
			if err != nil {
				return nil, err
			}
			mcopy.SetMapIndex(key, reflect.ValueOf(nested))
			continue
		}

		if key.Kind() != reflect.String {
			continue
		}

		if wrapper, found := c.typeAdaptersMap[v.Type()]; found {
			mcopy.SetMapIndex(c.newKey(key, wrapper), reflect.ValueOf(wrapper.Wrap(v.Interface())))
		} else {
			mcopy.SetMapIndex(key, v)
		}
	}

	return mcopy.Interface(), nil
}

// MarshalUnmarshaler is the interface that must be implemented if you want to
// add support for more than JSON to this codec.
type MarshalUnmarshaler interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(b []byte, v interface{}) error
}

type jsonMarshaler int

func (jsonMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonMarshaler) Unmarshal(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}

// Based on: https://github.com/golang/go/blob/178a2c42254166cffed1b25fb1d3c7a5727cada6/src/text/template/exec.go#L931
func indirectInterface(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Interface {
		return v
	}
	if v.IsNil() {
		return reflect.Value{}
	}
	return v.Elem()
}
