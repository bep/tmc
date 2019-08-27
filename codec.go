// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package typedmapcodec

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

const typeSep = "|"

// JSONMarshaler encodes and decodes JSON and is the default used in this
// codec.
var JSONMarshaler = new(jsonMarshaler)

func New() Codec {
	return Codec{
		marshaler: JSONMarshaler,
	}
}

type Codec struct {
	marshaler MarshalUnmarshaler
}

func (c Codec) Marshal(v interface{}) ([]byte, error) {
	m, err := c.toTypedMap(v)
	if err != nil {
		return nil, err
	}
	return c.marshaler.Marshal(m)
}

func (c Codec) Unmarshal(data []byte, v interface{}) error {
	if err := c.marshaler.Unmarshal(data, v); err != nil {
		return err
	}
	return c.fromTypedMap(v)
}

func (c Codec) fromTypedMap(v interface{}) error {
	mv := reflect.ValueOf(v)
	if mv.Kind() == reflect.Ptr {
		mv = mv.Elem()
	}

	if mv.Kind() != reflect.Map {
		return errors.New("must be a Map")
	}

	for _, key := range mv.MapKeys() {
		if key.Kind() != reflect.String {
			continue
		}

		keyStr := key.String()

		sepIdx := strings.LastIndex(keyStr, typeSep)
		if sepIdx == -1 {
			continue
		}

		keyPlain := keyStr[:sepIdx]
		keyType := keyStr[sepIdx+len(typeSep):]

		if wrapper, found := typeAdaptersStringMap[keyType]; found {
			ov := indirectInterface(mv.MapIndex(key))
			nv, err := wrapper.FromString(ov.String())
			if err != nil {
				return err
			}
			mv.SetMapIndex(reflect.ValueOf(keyPlain), reflect.ValueOf(nv))
			mv.SetMapIndex(key, reflect.Value{})
		}

	}

	return nil
}

func (c Codec) toTypedMap(v interface{}) (interface{}, error) {
	mv := reflect.ValueOf(v)
	if mv.Kind() != reflect.Map {
		return nil, errors.New("must be a Map")
	}

	mcopy := reflect.MakeMap(mv.Type())

	for _, key := range mv.MapKeys() {
		if key.Kind() != reflect.String {
			continue
		}

		v := indirectInterface(mv.MapIndex(key))

		if wrapper, found := typeAdaptersMap[v.Type()]; found {
			mcopy.SetMapIndex(newKey(key, wrapper), reflect.ValueOf(wrapper.Wrap(v.Interface())))
		} else {
			mcopy.SetMapIndex(key, v)
		}
	}

	return mcopy.Interface(), nil
}

type MarshalUnmarshaler interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(b []byte, v interface{}) error
}

type jsonMarshaler int

func (j jsonMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j jsonMarshaler) Unmarshal(b []byte, v interface{}) error {
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

func init() {
	for _, w := range DefaultTypeAdapters {
		tp := w.Type()
		typeAdaptersMap[tp] = w
		typeAdaptersStringMap[tp.String()] = w
	}
}
