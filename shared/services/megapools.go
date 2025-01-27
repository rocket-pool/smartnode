package services

import (
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	"github.com/urfave/cli"
)

func GetStakeValidatorInfo(c *cli.Context, wallet *wallet.Wallet, eth2Config beacon.Eth2Config, mp megapool.Megapool, validatorPubkey types.ValidatorPubkey) (types.ValidatorSignature, common.Hash, megapool.ValidatorProof, error) {
	// Get validator private key
	validatorKey, err := wallet.GetValidatorKeyByPubkey(validatorPubkey)
	if err != nil {
		return types.ValidatorSignature{}, common.Hash{}, megapool.ValidatorProof{}, err
	}

	withdrawalCredentials, err := mp.GetWithdrawalCredentials(nil)
	if err != nil {
		return types.ValidatorSignature{}, common.Hash{}, megapool.ValidatorProof{}, err
	}

	depositAmount := uint64(31e9) // 31 ETH in gwei

	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
	if err != nil {
		return types.ValidatorSignature{}, common.Hash{}, megapool.ValidatorProof{}, err
	}
	signature := types.BytesToValidatorSignature(depositData.Signature)

	bc, err := GetBeaconClient(c)
	if err != nil {
		return types.ValidatorSignature{}, common.Hash{}, megapool.ValidatorProof{}, err
	}

	// Get the validator index on the beacon chain
	validatorIndex, err := bc.GetValidatorIndex(validatorPubkey)
	if err != nil {
		return types.ValidatorSignature{}, common.Hash{}, megapool.ValidatorProof{}, err
	}

	validatorIndex64, err := strconv.ParseUint(validatorIndex, 10, 64)
	if err != nil {
		return types.ValidatorSignature{}, common.Hash{}, megapool.ValidatorProof{}, err
	}

	// Get the finalized block
	block, _, err := bc.GetBeaconBlock("finalized")
	if err != nil {
		return types.ValidatorSignature{}, common.Hash{}, megapool.ValidatorProof{}, err
	}

	// Get the beacon state for that slot
	beaconState, err := bc.GetBeaconState(block.Slot)
	if err != nil {
		return types.ValidatorSignature{}, common.Hash{}, megapool.ValidatorProof{}, err
	}

	proofBytes, err := beaconState.ValidatorProof(validatorIndex64)
	if err != nil {
		return types.ValidatorSignature{}, common.Hash{}, megapool.ValidatorProof{}, err
	}

	// Convert [][]byte to [][32]byte
	proofWithFixedSize := convertToFixedSize(proofBytes)

	proof := megapool.ValidatorProof{
		Slot:                  block.Slot,
		ValidatorIndex:        new(big.Int).SetUint64(validatorIndex64),
		Pubkey:                validatorPubkey[:],
		WithdrawalCredentials: withdrawalCredentials,
		Witnesses:             proofWithFixedSize,
	}

	return signature, depositDataRoot, proof, err
}

func convertToFixedSize(proofBytes [][]byte) [][32]byte {
	var proofWithFixedSize [][32]byte
	for _, b := range proofBytes {
		if len(b) != 32 {
			panic("each byte slice must be exactly 32 bytes long")
		}
		var arr [32]byte
		copy(arr[:], b)
		proofWithFixedSize = append(proofWithFixedSize, arr)
	}
	return proofWithFixedSize
}
