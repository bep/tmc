// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package typedmapcodec_test

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/bep/typedmapcodec"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
)

func TestRoundtrip(t *testing.T) {
	c := qt.New(t)

	src := map[string]interface{}{
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
	}

	for _, test := range []struct {
		name       string
		options    []typedmapcodec.Option
		dataAssert func(c *qt.C, data string)
	}{
		{"Default", nil,
			func(c *qt.C, data string) { c.Assert(data, qt.Contains, `{"vduration|time.Duration":"3s"`) }},
		{"Custom type separator",
			[]typedmapcodec.Option{
				typedmapcodec.WithTypeSep("TYPE:"),
			},
			func(c *qt.C, data string) { c.Assert(data, qt.Contains, "TYPE:") },
		},
		{"YAML",
			[]typedmapcodec.Option{
				typedmapcodec.WithMarshalUnmarshaler(new(yamlMarshaler)),
			},
			func(c *qt.C, data string) { c.Assert(data, qt.Contains, "vduration|time.Duration: 3s") },
		},
		{"JSON indent",
			[]typedmapcodec.Option{
				typedmapcodec.WithMarshalUnmarshaler(new(jsonMarshalerIndent)),
			},
			func(c *qt.C, data string) {
				c.Log(data)
				c.Assert(data, qt.Contains, `vduration`)
			},
		},
	} {

		test := test
		c.Run(test.name, func(c *qt.C) {
			c.Parallel()

			codec, err := typedmapcodec.New(test.options...)
			c.Assert(err, qt.IsNil)

			data, err := codec.Marshal(src)
			c.Assert(err, qt.IsNil)
			if test.dataAssert != nil {
				test.dataAssert(c, string(data))
			}

			dst := make(map[string]interface{})
			c.Assert(codec.Unmarshal(data, &dst), qt.IsNil)

			c.Assert(dst, eq, src)
		})

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

type jsonMarshalerIndent int

func (jsonMarshalerIndent) Marshal(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func (jsonMarshalerIndent) Unmarshal(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}
