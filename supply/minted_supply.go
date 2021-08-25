// (c) 2021, Stephen Buttolph. All rights reserved.
// See the file LICENSE for licensing terms.

package supply

import (
	"encoding/json"
	"errors"

	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

const (
	initialSupply = 360 * units.MegaAvax
)

var (
	errMissingValidatorReward = errors.New("expected validator's potential reward to be present")
	errMissingDelegatorReward = errors.New("expected delegator's potential reward to be present")
)

func GetAmountMinted(client *platformvm.Client) (uint64, error) {
	currentAllocatedSupply, err := client.GetCurrentSupply()
	if err != nil {
		return 0, err
	}

	currentValidators, err := client.GetCurrentValidators(constants.PrimaryNetworkID, nil)
	if err != nil {
		return 0, err
	}

	var newlyAllocatedSupply uint64
	for _, validatorMap := range currentValidators {
		validatorBytes, err := json.Marshal(validatorMap)
		if err != nil {
			return 0, err
		}

		validator := platformvm.APIPrimaryValidator{}
		err = json.Unmarshal(validatorBytes, &validator)
		if err != nil {
			return 0, err
		}

		if validator.PotentialReward == nil {
			return 0, errMissingValidatorReward
		}

		newlyAllocatedSupply += uint64(*validator.PotentialReward)

		for _, delegator := range validator.Delegators {
			if delegator.PotentialReward == nil {
				return 0, errMissingDelegatorReward
			}

			newlyAllocatedSupply += uint64(*delegator.PotentialReward)
		}
	}

	return currentAllocatedSupply - initialSupply - newlyAllocatedSupply, nil
}
