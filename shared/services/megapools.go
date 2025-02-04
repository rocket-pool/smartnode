package services

import (
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/core/signing"
	prdeposit "github.com/prysmaticlabs/prysm/v5/contracts/deposit"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/rocketpool-go/types"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	"github.com/urfave/cli"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

func GetStakeValidatorInfo(c *cli.Context, wallet *wallet.Wallet, eth2Config beacon.Eth2Config, megapoolAddress common.Address, validatorPubkey types.ValidatorPubkey) (types.ValidatorSignature, common.Hash, megapool.ValidatorProof, error) {
	// Get validator private key
	validatorKey, err := wallet.GetValidatorKeyByPubkey(validatorPubkey)
	if err != nil {
		return types.ValidatorSignature{}, common.Hash{}, megapool.ValidatorProof{}, err
	}

	withdrawalCredentials := CalculateMegapoolWithdrawalCredentials(megapoolAddress)

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

	err = validateDepositInfo(eth2Config, uint64(depositAmount), validatorPubkey, withdrawalCredentials, signature)
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

	proofBytes, err := beaconState.ValidatorCredentialsProof(validatorIndex64)
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

func validateDepositInfo(eth2Config beacon.Eth2Config, depositAmount uint64, pubkey rptypes.ValidatorPubkey, withdrawalCredentials common.Hash, signature rptypes.ValidatorSignature) error {

	// Get the deposit domain based on the eth2 config
	depositDomain, err := signing.ComputeDomain(eth2types.DomainDeposit, eth2Config.GenesisForkVersion, eth2types.ZeroGenesisValidatorsRoot)
	if err != nil {
		return err
	}

	// Create the deposit struct
	depositData := new(ethpb.Deposit_Data)
	depositData.Amount = depositAmount
	depositData.PublicKey = pubkey.Bytes()
	depositData.WithdrawalCredentials = withdrawalCredentials.Bytes()
	depositData.Signature = signature.Bytes()

	// Validate the signature
	err = prdeposit.VerifyDepositSignature(depositData, depositDomain)
	return err

}

func CalculateMegapoolWithdrawalCredentials(megapoolAddress common.Address) common.Hash {
	// Convert the address to a uint160 (20 bytes) and then to a uint256 (32 bytes)
	addressBigInt := new(big.Int)
	addressBigInt.SetString(megapoolAddress.Hex()[2:], 16) // Remove the "0x" prefix and convert from hex

	// Shift 0x01 left by 248 bits
	shiftedValue := new(big.Int).Lsh(big.NewInt(0x01), 248)

	// Perform the bitwise OR operation
	result := new(big.Int).Or(shiftedValue, addressBigInt)

	// Convert the result to a 32-byte array (bytes32)
	var bytes32 [32]byte
	resultBytes := result.Bytes()
	copy(bytes32[32-len(resultBytes):], resultBytes)

	return common.BytesToHash(resultBytes)

}
