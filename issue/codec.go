// (c) 2021, Stephen Buttolph. All rights reserved.
// See the file LICENSE for licensing terms.

package issue

import (
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/nftfx"
	"github.com/ava-labs/avalanchego/vms/propertyfx"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

var c codec.Manager

func init() {
	gc, _, err := avm.NewCodecs([]avm.Fx{
		&secp256k1fx.Fx{},
		&nftfx.Fx{},
		&propertyfx.Fx{},
	})
	if err != nil {
		panic(err)
	}
	c = gc
}
