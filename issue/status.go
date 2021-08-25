// (c) 2021, Stephen Buttolph. All rights reserved.
// See the file LICENSE for licensing terms.

package issue

import (
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

// ConfirmTx attempts to confirm [txID] by checking its status [attempts] times
// with a [delay] in between each attempt. If the transaction has not been decided
// by the final attempt, it returns the status of the last attempt.
// Note: ConfirmTx will block until either the last attempt finishes or the client
// returns a decided status.
func ConfirmTx(c *platformvm.Client, txID ids.ID, attempts int, delay time.Duration) (platformvm.Status, error) {
	for i := 0; i < attempts-1; i++ {
		resp, err := c.GetTxStatus(txID, true)
		if err != nil {
			return platformvm.Unknown, err
		}
		status := resp.Status

		switch status {
		case platformvm.Committed, platformvm.Aborted:
			return status, nil
		}

		time.Sleep(delay)
	}

	resp, err := c.GetTxStatus(txID, false)
	if err != nil {
		return platformvm.Unknown, err
	}
	return resp.Status, nil
}
