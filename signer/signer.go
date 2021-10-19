// (c) 2021, Stephen Buttolph. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/platformvm"
)

func Sign(inFilePath, outFilePath, secretKey string) error {
	secretKeyBytes, err := formatting.Decode(formatting.CB58, secretKey)
	if err != nil {
		return err
	}

	secp := crypto.FactorySECP256K1R{}
	skIntf, err := secp.ToPrivateKey(secretKeyBytes)
	if err != nil {
		return err
	}
	sk := skIntf.(*crypto.PrivateKeySECP256K1R)

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
	for scanner.Scan() {
		unsignedTxHex := scanner.Text()
		unsignedTxBytes, err := hex.DecodeString(unsignedTxHex)
		if err != nil {
			return err
		}

		var unsignedTx platformvm.UnsignedTx
		version, err := platformvm.Codec.Unmarshal(unsignedTxBytes, &unsignedTx)
		if err != nil {
			return err
		}
		if version != 0 {
			return fmt.Errorf("expected codec version 0 but got %d", version)
		}

		tx := platformvm.Tx{
			UnsignedTx: unsignedTx,
		}
		err = tx.Sign(platformvm.Codec, [][]*crypto.PrivateKeySECP256K1R{{sk}})
		if err != nil {
			return err
		}

		signedTxBytes := tx.Bytes()
		signedTxHex, err := formatting.EncodeWithChecksum(formatting.Hex, signedTxBytes)
		if err != nil {
			return err
		}

		_, err = outFile.WriteString(signedTxHex)
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
