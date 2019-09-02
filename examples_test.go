// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tmc_test

import (
	"fmt"
	"log"

	"github.com/bep/tmc"
	yaml "gopkg.in/yaml.v2"
)

func Example() {
	m1 := map[string]interface{}{"num": 42}
	c, err := tmc.New()
	if err != nil {
		log.Fatal(err)
	}

	data, err := c.Marshal(m1)
	if err != nil {
		log.Fatal(err)
	}

	m2 := make(map[string]interface{})
	err = c.Unmarshal(data, &m2)
	if err != nil {
		log.Fatal(err)
	}
	num := m2["num"]

	fmt.Printf("%v (%T)", num, num)
	// Output: 42 (int)

}

func ExampleWithMarshalUnmarshaler() {
	m1 := map[string]interface{}{"num": 42}
	c, err := tmc.New(tmc.WithMarshalUnmarshaler(new(yamlMarshaler)))
	if err != nil {
		log.Fatal(err)
	}

	data, err := c.Marshal(m1)
	if err != nil {
		log.Fatal(err)
	}

	m2 := make(map[string]interface{})
	err = c.Unmarshal(data, &m2)
	if err != nil {
		log.Fatal(err)
	}
	num := m2["num"]

	fmt.Printf("%v (%T)", num, num)
	// Output: 42 (int)
}

type yamlMarshaler int

func (yamlMarshaler) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)

}

func (yamlMarshaler) Unmarshal(b []byte, v interface{}) error {
	return yaml.Unmarshal(b, v)
}
