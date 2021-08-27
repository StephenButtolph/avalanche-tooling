// (c) 2021, Stephen Buttolph. All rights reserved.
// See the file LICENSE for licensing terms.

package uptime

import (
	"encoding/json"
	"fmt"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

func GetDownedNodesWithWeight(client *platformvm.Client) (map[string]uint64, error) {
	currentValidators, err := client.GetCurrentValidators(constants.PrimaryNetworkID, nil)
	if err != nil {
		return nil, err
	}

	down := map[string]uint64{}
	for _, validatorMap := range currentValidators {
		validatorBytes, err := json.Marshal(validatorMap)
		if err != nil {
			return nil, err
		}

		validator := platformvm.APIPrimaryValidator{}
		err = json.Unmarshal(validatorBytes, &validator)
		if err != nil {
			return nil, err
		}

		if *validator.Connected {
			continue
		}

		stake := uint64(*validator.StakeAmount)
		for _, delegator := range validator.Delegators {
			stake += uint64(*delegator.StakeAmount)
		}
		down[validator.NodeID] = stake
	}
	return down, nil
}

func DisplayDown(client *platformvm.Client) error {
	down, err := GetDownedNodesWithWeight(client)
	if err != nil {
		return err
	}

	for nodeID, stake := range down {
		if stake < 50*units.KiloAvax {
			continue
		}
		fmt.Printf("%-40s with %d\n", nodeID, stake/units.Avax)
	}
	return nil
}
