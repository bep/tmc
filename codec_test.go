// Copyright © 2019 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package typedmapcodec_test

import (
	"math/big"
	"testing"
	"time"

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
	}

	codec := typedmapcodec.New()

	data, err := codec.Marshal(src)
	c.Assert(err, qt.IsNil)

	dst := make(map[string]interface{})
	c.Assert(codec.Unmarshal(data, &dst), qt.IsNil)

	c.Assert(dst, eq, src)
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
