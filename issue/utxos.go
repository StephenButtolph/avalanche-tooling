// (c) 2021, Stephen Buttolph. All rights reserved.
// See the file LICENSE for licensing terms.

package issue

import (
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

const (
	utxoPageSize = 1024
)

func GetXChainUTXOs(
	networkID uint32,
	xClient *avm.Client,
	keychain *secp256k1fx.Keychain,
) (map[ids.ID]*avax.UTXO, error) {
	hrp := constants.GetHRP(networkID)

	ownedAddresses := []string(nil)
	for ownedAddr := range keychain.Addrs {
		ownedAddress, err := formatting.FormatAddress("X", hrp, ownedAddr[:])
		if err != nil {
			return nil, err
		}
		ownedAddresses = append(ownedAddresses, ownedAddress)
	}

	utxos := make(map[ids.ID]*avax.UTXO)
	index := api.Index{}
	for {
		rawUTXOs, newIndex, err := xClient.GetUTXOs(ownedAddresses, utxoPageSize, index.Address, index.UTXO)
		if err != nil {
			return nil, err
		}
		index = newIndex

		for _, rawUTXO := range rawUTXOs {
			utxo := avax.UTXO{}
			_, err := c.Unmarshal(rawUTXO, &utxo)
			if err != nil {
				return nil, err
			}
			utxos[utxo.InputID()] = &utxo
		}

		if len(rawUTXOs) != utxoPageSize {
			return utxos, nil
		}
	}
}

func GetPChainAtomicUTXOs(
	networkID uint32,
	sourceChain ids.ID,
	pClient *platformvm.Client,
	keychain *secp256k1fx.Keychain,
) (map[ids.ID]*avax.UTXO, error) {
	hrp := constants.GetHRP(networkID)

	ownedAddresses := []string(nil)
	for ownedAddr := range keychain.Addrs {
		ownedAddress, err := formatting.FormatAddress("P", hrp, ownedAddr[:])
		if err != nil {
			return nil, err
		}
		ownedAddresses = append(ownedAddresses, ownedAddress)
	}

	utxos := make(map[ids.ID]*avax.UTXO)
	index := api.Index{}
	for {
		rawUTXOs, newIndex, err := pClient.GetAtomicUTXOs(ownedAddresses, sourceChain.String(), utxoPageSize, index.Address, index.UTXO)
		if err != nil {
			return nil, err
		}
		index = newIndex

		for _, rawUTXO := range rawUTXOs {
			utxo := avax.UTXO{}
			_, err := c.Unmarshal(rawUTXO, &utxo)
			if err != nil {
				return nil, err
			}
			utxos[utxo.InputID()] = &utxo
		}

		if len(rawUTXOs) != utxoPageSize {
			return utxos, nil
		}
	}
}
