// (c) 2021, Stephen Buttolph. All rights reserved.
// See the file LICENSE for licensing terms.

package checksum

import (
	"bufio"
	"encoding/hex"
	"os"

	"github.com/ava-labs/avalanchego/utils/formatting"
)

func AddChecksum(inFilePath, outFilePath string) error {
	inFile, err := os.Open(inFilePath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(inFile)
	const maxCapacity = 500_000_000
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	for scanner.Scan() {
		rawHex := scanner.Text()
		bytes, err := hex.DecodeString(rawHex)
		if err != nil {
			return err
		}

		checkedHex, err := formatting.EncodeWithChecksum(formatting.Hex, bytes)
		if err != nil {
			return err
		}

		_, err = outFile.WriteString(checkedHex)
		if err != nil {
			return err
		}

		_, err = outFile.WriteString("\n")
		if err != nil {
			return err
		}
	}

	return scanner.Err()
}
