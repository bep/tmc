// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package typedmapcodec

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
)

func TestRoundtrip(t *testing.T) {
	c := qt.New(t)

	src := newTestMap()

	for _, test := range []struct {
		name       string
		options    []Option
		dataAssert func(c *qt.C, data string)
	}{
		{"Default", nil,
			func(c *qt.C, data string) { c.Assert(data, qt.Contains, `{"nested":{"vduration|time.Duration":"5s"`) }},
		{"Custom type separator",
			[]Option{
				WithTypeSep("TYPE:"),
			},
			func(c *qt.C, data string) { c.Assert(data, qt.Contains, "TYPE:") },
		},
		{"YAML",
			[]Option{
				WithMarshalUnmarshaler(new(yamlMarshaler)),
			},
			func(c *qt.C, data string) { c.Assert(data, qt.Contains, "vduration|time.Duration: 3s") },
		},
		{"JSON indent",
			[]Option{
				WithMarshalUnmarshaler(new(jsonMarshalerIndent)),
			},
			func(c *qt.C, data string) {

			},
		},
	} {

		test := test
		c.Run(test.name, func(c *qt.C) {
			c.Parallel()

			codec, err := New(test.options...)
			c.Assert(err, qt.IsNil)

			data, err := codec.Marshal(src)
			c.Assert(err, qt.IsNil)
			c.Log(string(data))
			if test.dataAssert != nil {
				test.dataAssert(c, string(data))
			}

			dst := make(map[string]interface{})
			c.Assert(codec.Unmarshal(data, &dst), qt.IsNil)

			c.Assert(dst, eq, src)
		})

	}

}

func TestErrors(t *testing.T) {
	c := qt.New(t)
	codec, err := New()
	c.Assert(err, qt.IsNil)
	marshal := func(v interface{}) error {
		_, err := codec.Marshal(v)
		return err
	}

	// OK
	c.Assert(marshal(map[string]interface{}{"32": "a"}), qt.IsNil)
	c.Assert(marshal(map[string]int{"32": 32}), qt.IsNil)

	// Should fail
	c.Assert(marshal([]string{"a"}), qt.Not(qt.IsNil))
	c.Assert(marshal(map[int]interface{}{32: "a"}), qt.Not(qt.IsNil))
	c.Assert(marshal(map[string]interface{}{"a": map[int]string{32: "32"}}), qt.Not(qt.IsNil))
}

func BenchmarkCodec(b *testing.B) {
	b.Run("JSON regular", func(b *testing.B) {
		b.StopTimer()
		mi := newTestMap()
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			data, err := json.Marshal(mi)
			if err != nil {
				b.Fatal(err)
			}
			m := make(map[string]interface{})
			if err := json.Unmarshal(data, &m); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("JSON typed", func(b *testing.B) {
		b.StopTimer()
		mi := newTestMap()
		c, err := New()
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			data, err := c.Marshal(mi)
			if err != nil {
				b.Fatal(err)
			}
			m := make(map[string]interface{})
			if err := c.Unmarshal(data, &m); err != nil {
				b.Fatal(err)
			}
		}
	})

}

func newTestMap() map[string]interface{} {
	return map[string]interface{}{
		"vstring":     "Hello1",
		"vstring|":    "Hello2",
		"vstring|foo": "Hello3",
		// Numbers
		"vint":     32,
		"vfloat32": float32(3.14159),
		"vfloat64": float64(3.14159),
		"vrat":     big.NewRat(1, 2),
		// Time
		"vtime":     time.Now(),
		"vduration": 3 * time.Second,
		"vsliceint": []int{1, 3, 4},
		"nested": map[string]interface{}{
			"vint":      55,
			"vduration": 5 * time.Second,
		},
		"nested-typed-int": map[string]int{
			"vint": 42,
		},
		"nested-typed-duration": map[string]time.Duration{
			"v1": 5 * time.Second,
			"v2": 10 * time.Second,
		},
	}
}

var eq = qt.CmpEquals(
	cmp.Comparer(
		func(v1, v2 *big.Rat) bool {
			return v1.RatString() == v2.RatString()
		},
	),
	cmp.Comparer(func(v1, v2 time.Time) bool {
		// UnmarshalText always create times with no monotonic clock reading,
		// so we cannot compare with ==.
		// TODO(bep) improve this
		return v1.Unix() == v2.Unix()
	}),
)

type yamlMarshaler int

func (yamlMarshaler) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)

}

func (yamlMarshaler) Unmarshal(b []byte, v interface{}) error {
	return yaml.Unmarshal(b, v)
}

// Useful for debugging
type jsonMarshalerIndent int

func (jsonMarshalerIndent) Marshal(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func (jsonMarshalerIndent) Unmarshal(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}
