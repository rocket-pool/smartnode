package validator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/eth2"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// Get the withdrawal private key for a validator based on its mnemonic, index, and path
func GetWithdrawalKey(mnemonic string, index uint, validatorKeyPath string) (*eth2types.BLSPrivateKey, error) {

	withdrawalKeyPath := strings.TrimSuffix(validatorKeyPath, "/0")
	withdrawalKey, err := GetPrivateKey(mnemonic, index, withdrawalKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error getting withdrawal private key: %w", err)
	}

	return withdrawalKey, nil

}

// Get a voluntary exit message signature for a given validator key and index
func GetSignedWithdrawalCredsChangeMessage(withdrawalKey *eth2types.BLSPrivateKey, validatorIndex string, newWithdrawalAddress common.Address, signatureDomain []byte) (types.ValidatorSignature, error) {

	// Get the withdrawal pubkey
	withdrawalPubkey := withdrawalKey.PublicKey().Marshal()
	withdrawalPubkeyBuffer := [48]byte{}
	copy(withdrawalPubkeyBuffer[:], withdrawalPubkey)

	// Convert the validator index to a uint
	indexNum, err := strconv.ParseUint(validatorIndex, 10, 64)
	if err != nil {
		return types.ValidatorSignature{}, fmt.Errorf("error parsing validator index (%s): %w", validatorIndex, err)
	}

	// Build withdrawal creds change message
	message := eth2.WithdrawalCredentialsChange{
		ValidatorIndex:     indexNum,
		FromBLSPubkey:      withdrawalPubkeyBuffer,
		ToExecutionAddress: newWithdrawalAddress,
	}

	// Get object root
	or, err := message.HashTreeRoot()
	if err != nil {
		return types.ValidatorSignature{}, err
	}

	// Get signing root
	sr := eth2.SigningRoot{
		ObjectRoot: or[:],
		Domain:     signatureDomain,
	}

	srHash, err := sr.HashTreeRoot()
	if err != nil {
		return types.ValidatorSignature{}, err
	}

	// Sign message
	signature := withdrawalKey.Sign(srHash[:]).Marshal()

	// Return
	return types.BytesToValidatorSignature(signature), nil

}
