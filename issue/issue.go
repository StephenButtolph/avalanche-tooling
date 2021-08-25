// (c) 2021, Stephen Buttolph. All rights reserved.
// See the file LICENSE for licensing terms.

package issue

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

const (
	maxOutputsPerTx = 1000
)

var (
	errDuplicatedAddresses = errors.New("duplicated addresses")
	errSpendOverflow       = errors.New("spent amount overflows uint64")
)

func SendOutputsOtherToP(
	networkID uint32,
	chainID ids.ID,
	sourceChainID ids.ID,
	pClient *platformvm.Client,
	keychain *secp256k1fx.Keychain,
	txOuts [][]*avax.TransferableOutput,
	feeAssetID ids.ID,
	feeAmount uint64,
) error {
	numSent := 0
	for _, outs := range txOuts {
		cost, err := GetCost(outs, feeAssetID, feeAmount)
		if err != nil {
			return err
		}

		utxos, err := GetPChainAtomicUTXOs(networkID, sourceChainID, pClient, keychain)
		if err != nil {
			return err
		}

		spent, ins, keys, err := BuildInputs(utxos, keychain, cost)
		if err != nil {
			return err
		}

		changeAddr := keychain.Keys[0].PublicKey().Address()

		changeOutputs := GetChangeOutputs(changeAddr, cost, spent)

		newOuts := []*avax.TransferableOutput(nil)
		newOuts = append(newOuts, outs...)
		newOuts = append(newOuts, changeOutputs...)
		avax.SortTransferableOutputs(outs, platformvm.Codec)

		tx, err := BuildImportTx(networkID, chainID, sourceChainID, newOuts, ins, keys)
		if err != nil {
			return err
		}

		{
			// txID := tx.ID()
			// txJSON, err := json.MarshalIndent(tx, "", "  ")
			// if err != nil {
			// 	return err
			// }
			// log.Printf("txID = %s\n%s", txID, string(txJSON))
		}

		txBytes := tx.Bytes()

		txID, err := pClient.IssueTx(txBytes)
		if err != nil {
			return err
		}

		txStatus, err := ConfirmTx(pClient, txID, 100, 100*time.Millisecond)
		if err != nil {
			return err
		}

		addrStr, err := formatting.FormatAddress("P", constants.GetHRP(networkID), changeAddr[:])
		if err != nil {
			return err
		}

		numSent += len(outs)
		log.Printf("%s - %s - %s - %d", txID, txStatus, addrStr, numSent)

		time.Sleep(1 * time.Second)
	}
	return nil
}

func BuildImportTx(
	networkID uint32,
	chainID ids.ID,
	sourceChainID ids.ID,
	outs []*avax.TransferableOutput,
	ins []*avax.TransferableInput,
	keys [][]*crypto.PrivateKeySECP256K1R,
) (
	*platformvm.Tx,
	error,
) {
	tx := &platformvm.Tx{UnsignedTx: &platformvm.UnsignedImportTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    networkID,
			BlockchainID: chainID,
			Outs:         outs,
		}},
		SourceChain:    sourceChainID,
		ImportedInputs: ins,
	}}
	return tx, tx.Sign(platformvm.Codec, keys)
}

func SendOutputsXToOther(
	networkID uint32,
	chainID ids.ID,
	destinationChainID ids.ID,
	xClient *avm.Client,
	keychain *secp256k1fx.Keychain,
	txOuts [][]*avax.TransferableOutput,
	feeAssetID ids.ID,
	feeAmount uint64,
) error {
	for _, outs := range txOuts {
		cost, err := GetCost(outs, feeAssetID, feeAmount)
		if err != nil {
			return err
		}

		utxos, err := GetXChainUTXOs(networkID, xClient, keychain)
		if err != nil {
			return err
		}

		spent, ins, keys, err := BuildInputs(utxos, keychain, cost)
		if err != nil {
			return err
		}

		changeAddr := keychain.Keys[0].PublicKey().Address()

		changeOutputs := GetChangeOutputs(changeAddr, cost, spent)

		avax.SortTransferableOutputs(outs, c)
		avax.SortTransferableOutputs(changeOutputs, c)

		tx, err := BuildExportTx(networkID, chainID, destinationChainID, outs, changeOutputs, ins, keys)
		if err != nil {
			return err
		}
		txBytes := tx.Bytes()

		{
			// txID := tx.ID()
			// txJSON, err := json.MarshalIndent(tx, "", "  ")
			// if err != nil {
			// 	return err
			// }
			// log.Printf("txID = %s\n%s", txID, string(txJSON))
		}

		txID, err := xClient.IssueTx(txBytes)
		if err != nil {
			return err
		}

		txStatus, err := xClient.ConfirmTx(txID, 100, 100*time.Millisecond)
		if err != nil {
			return err
		}

		addrStr, err := formatting.FormatAddress("X", constants.GetHRP(networkID), changeAddr[:])
		if err != nil {
			return err
		}

		log.Printf("%s - %s - %s", txID, txStatus, addrStr)

		time.Sleep(1 * time.Second)
	}
	return nil
}

func BuildExportTx(
	networkID uint32,
	chainID ids.ID,
	destinationChainID ids.ID,
	exportedOuts []*avax.TransferableOutput,
	returnedOuts []*avax.TransferableOutput,
	ins []*avax.TransferableInput,
	keys [][]*crypto.PrivateKeySECP256K1R,
) (
	*avm.Tx,
	error,
) {
	tx := &avm.Tx{UnsignedTx: &avm.ExportTx{
		BaseTx: avm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    networkID,
			BlockchainID: chainID,
			Outs:         returnedOuts,
			Ins:          ins,
		}},
		DestinationChain: destinationChainID,
		ExportedOuts:     exportedOuts,
	}}
	return tx, tx.SignSECP256K1Fx(c, keys)
}

func SendOutputsXToX(
	networkID uint32,
	chainID ids.ID,
	xClient *avm.Client,
	keychain *secp256k1fx.Keychain,
	txOuts [][]*avax.TransferableOutput,
	feeAssetID ids.ID,
	feeAmount uint64,
) error {
	for _, outs := range txOuts {
		cost, err := GetCost(outs, feeAssetID, feeAmount)
		if err != nil {
			return err
		}

		utxos, err := GetXChainUTXOs(networkID, xClient, keychain)
		if err != nil {
			return err
		}

		spent, ins, keys, err := BuildInputs(utxos, keychain, cost)
		if err != nil {
			return err
		}

		changeAddr := keychain.Keys[0].PublicKey().Address()

		changeOutputs := GetChangeOutputs(changeAddr, cost, spent)

		newOuts := []*avax.TransferableOutput(nil)
		newOuts = append(newOuts, outs...)
		newOuts = append(newOuts, changeOutputs...)
		avax.SortTransferableOutputs(outs, c)

		tx, err := BuildBaseTx(networkID, chainID, newOuts, ins, keys)
		if err != nil {
			return err
		}
		txBytes := tx.Bytes()

		txID, err := xClient.IssueTx(txBytes)
		if err != nil {
			return err
		}

		txStatus, err := xClient.ConfirmTx(txID, 100, 100*time.Millisecond)
		if err != nil {
			return err
		}

		addrStr, err := formatting.FormatAddress("X", "fuji", changeAddr[:])
		if err != nil {
			return err
		}

		log.Printf("%s - %s - %s", txID, txStatus, addrStr)

		time.Sleep(1 * time.Second)
	}
	return nil
}

func BuildBaseTx(
	networkID uint32,
	chainID ids.ID,
	outs []*avax.TransferableOutput,
	ins []*avax.TransferableInput,
	keys [][]*crypto.PrivateKeySECP256K1R,
) (
	*avm.Tx,
	error,
) {
	tx := &avm.Tx{UnsignedTx: &avm.BaseTx{BaseTx: avax.BaseTx{
		NetworkID:    networkID,
		BlockchainID: chainID,
		Outs:         outs,
		Ins:          ins,
	}}}
	return tx, tx.SignSECP256K1Fx(c, keys)
}

func GetChangeOutputs(
	changeAddress ids.ShortID,
	produced,
	consumed map[ids.ID]uint64,
) []*avax.TransferableOutput {
	var outs []*avax.TransferableOutput
	// Add the required change outputs
	for assetID, amountConsumed := range consumed {
		amountProduced := produced[assetID]

		if amountConsumed > amountProduced {
			outs = append(outs, &avax.TransferableOutput{
				Asset: avax.Asset{ID: assetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: amountConsumed - amountProduced,
					OutputOwners: secp256k1fx.OutputOwners{
						Locktime:  0,
						Threshold: 1,
						Addrs:     []ids.ShortID{changeAddress},
					},
				},
			})
		}
	}
	return outs
}

func GetCost(
	outs []*avax.TransferableOutput,
	feeAssetID ids.ID,
	feeAmount uint64,
) (map[ids.ID]uint64, error) {
	amounts := make(map[ids.ID]uint64)
	for _, out := range outs {
		assetID := out.AssetID()
		newAmount, err := math.Add64(amounts[assetID], out.Out.Amount())
		if err != nil {
			return nil, err
		}
		amounts[assetID] = newAmount
	}

	// Add the fee
	amountWithFee, err := math.Add64(amounts[feeAssetID], feeAmount)
	if err != nil {
		return nil, err
	}
	amounts[feeAssetID] = amountWithFee
	return amounts, nil
}

func BuildInputs(
	utxos map[ids.ID]*avax.UTXO,
	keychain *secp256k1fx.Keychain,
	amounts map[ids.ID]uint64,
) (
	map[ids.ID]uint64,
	[]*avax.TransferableInput,
	[][]*crypto.PrivateKeySECP256K1R,
	error,
) {
	amountsSpent := make(map[ids.ID]uint64, len(amounts))
	time := uint64(time.Now().Unix())

	ins := []*avax.TransferableInput{}
	keys := [][]*crypto.PrivateKeySECP256K1R{}
	for _, utxo := range utxos {
		assetID := utxo.AssetID()
		amount := amounts[assetID]
		amountSpent := amountsSpent[assetID]

		if amountSpent >= amount {
			// we already have enough inputs allocated to this asset
			continue
		}

		inputIntf, signers, err := keychain.Spend(utxo.Out, time)
		if err != nil {
			// this utxo can't be spent with the current keys right now
			continue
		}
		input, ok := inputIntf.(avax.TransferableIn)
		if !ok {
			// this input doesn't have an amount, so I don't care about it here
			continue
		}
		newAmountSpent, err := math.Add64(amountSpent, input.Amount())
		if err != nil {
			// there was an error calculating the consumed amount, just error
			return nil, nil, nil, errSpendOverflow
		}
		amountsSpent[assetID] = newAmountSpent

		// add the new input to the array
		ins = append(ins, &avax.TransferableInput{
			UTXOID: utxo.UTXOID,
			Asset:  avax.Asset{ID: assetID},
			In:     input,
		})
		// add the required keys to the array
		keys = append(keys, signers)
	}

	for asset, amount := range amounts {
		if amountsSpent[asset] < amount {
			return nil, nil, nil, fmt.Errorf("want to spend %d of asset %s but only have %d",
				amount,
				asset,
				amountsSpent[asset],
			)
		}
	}

	avax.SortTransferableInputsWithSigners(ins, keys)
	return amountsSpent, ins, keys, nil
}

func BuildOutputs(
	addresses []ids.ShortID,
	numUTXOsPerAddress int,
	assetID ids.ID,
	amountPerUTXO uint64,
) ([][]*avax.TransferableOutput, error) {
	ids.SortShortIDs(addresses)
	if !ids.IsSortedAndUniqueShortIDs(addresses) {
		return nil, errDuplicatedAddresses
	}

	var (
		outs               [][]*avax.TransferableOutput
		currentSpentAmount uint64
		currentOuts        []*avax.TransferableOutput
	)
	for _, addr := range addresses {
		addrStr, _ := formatting.FormatAddress("P", "local", addr[:])
		log.Println(addrStr, addr)

		for i := 0; i < numUTXOsPerAddress; i++ {
			newSpentAmount, err := math.Add64(currentSpentAmount, amountPerUTXO)
			if err != nil || len(currentOuts) >= maxOutputsPerTx {
				outs = append(outs, currentOuts)
				currentSpentAmount = amountPerUTXO
				currentOuts = nil
			} else {
				currentSpentAmount = newSpentAmount
			}

			currentOuts = append(currentOuts, &avax.TransferableOutput{
				Asset: avax.Asset{ID: assetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: amountPerUTXO,
					OutputOwners: secp256k1fx.OutputOwners{
						Locktime:  0,
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				},
			})
		}
	}
	if len(currentOuts) > 0 {
		outs = append(outs, currentOuts)
	}
	return outs, nil
}
