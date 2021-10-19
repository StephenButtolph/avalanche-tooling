// (c) 2021, Stephen Buttolph. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"log"
	"os"

	"github.com/StephenButtolph/avalanche-tooling/checksum"
)

func main() {
	// if len(os.Args) != 4 {
	// 	log.Fatalf("expected secret key, input file, and output file to be provided as arguments")
	// }

	// err := signer.Sign(os.Args[2], os.Args[3], os.Args[1])
	// if err != nil {
	// 	log.Fatal(err)
	// }

	if len(os.Args) != 3 {
		log.Fatalf("expected input file, and output file to be provided as arguments")
	}

	err := checksum.AddChecksum(os.Args[1], os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
}
