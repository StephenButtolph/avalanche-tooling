// (c) 2021, Stephen Buttolph. All rights reserved.
// See the file LICENSE for licensing terms.

package benched

import (
	"encoding/json"
	"fmt"

	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/network"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

func GetBenched(client *info.Client) ([]network.PeerID, error) {
	peers, err := client.Peers()
	if err != nil {
		return nil, err
	}

	var benchedPeers []network.PeerID
	for _, peer := range peers {
		if len(peer.Benched) > 0 {
			benchedPeers = append(benchedPeers, peer)
		}
	}
	return benchedPeers, nil
}

func DisplayBenched(infoClient *info.Client, platformClient *platformvm.Client) error {
	nodes, err := GetBenched(infoClient)
	if err != nil {
		return err
	}

	currentValidators, err := platformClient.GetCurrentValidators(constants.PrimaryNetworkID, nil)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		var stake uint64
		for _, validatorMap := range currentValidators {
			validatorBytes, err := json.Marshal(validatorMap)
			if err != nil {
				return err
			}

			validator := platformvm.APIPrimaryValidator{}
			err = json.Unmarshal(validatorBytes, &validator)
			if err != nil {
				return err
			}

			if validator.NodeID != node.ID {
				continue
			}

			stake = uint64(*validator.StakeAmount)
			for _, delegator := range validator.Delegators {
				stake += uint64(*delegator.StakeAmount)
			}
		}

		fmt.Printf("%-40s at %-20s on %s with %d\n", node.ID, node.IP, node.Version, stake/units.Avax)
	}
	return nil
}
