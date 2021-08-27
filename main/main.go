// (c) 2021, Stephen Buttolph. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"log"
	"os"
	"time"

	"github.com/ava-labs/avalanchego/vms/platformvm"

	"github.com/StephenButtolph/avalanche-tooling/uptime"
)

const (
	apiTimeout = 3 * time.Second
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("expected api endpoint to be provided as an argument")
	}

	// infoClient := info.NewClient(os.Args[1], apiTimeout)
	platformClient := platformvm.NewClient(os.Args[1], apiTimeout)

	err := uptime.DisplayDown(platformClient) //benched.DisplayBenched(infoClient, platformClient)
	if err != nil {
		log.Fatal(err)
	}
}
