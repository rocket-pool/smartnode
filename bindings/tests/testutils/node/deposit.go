package node

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
	"github.com/rocket-pool/rocketpool-go/tests/testutils/validator"
)

var salt int64 = 0

// Returns a unique salt for minipool address generation
func GetSalt() *big.Int {
	salt += 1
	return big.NewInt(salt)
}

// Call deposit on the node using the validator test values
func Deposit(t *testing.T, rp *rocketpool.RocketPool, nodeAccount *accounts.Account, depositAmount *big.Int, pubkey int) (common.Address, *types.Receipt, error) {

	// Get the next salt
	salt := GetSalt()

	// Get validator & deposit data
	depositType, err := node.GetDepositType(rp, depositAmount, nil)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("Error getting deposit type: %w", err)
	}
	validatorPubkey, err := validator.GetValidatorPubkey(pubkey)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("Error getting validator pubkey: %w", err)
	}
	expectedMinipoolAddress, err := utils.GenerateAddress(rp, nodeAccount.Address, depositType, salt, nil)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("Error generating minipool address: %w", err)
	}
	withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(rp, expectedMinipoolAddress, nil)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("Error getting minipool withdrawal credentials: %w", err)
	}
	validatorSignature, err := validator.GetValidatorSignature(pubkey)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("Error getting validator signature: %w", err)
	}
	depositDataRoot, err := validator.GetDepositDataRoot(validatorPubkey, withdrawalCredentials, validatorSignature)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("Error getting deposit data root: %w", err)
	}

	// Make node deposit
	opts := nodeAccount.GetTransactor()
	opts.Value = depositAmount

	minNodeFee := 0.0
	//t.Logf("Deposit:\n\tMin Node Fee: %f\n\tValidator Pubkey: %s\n\tValidator Signature: %s\n\tDeposit Data Root: %s\n\tNode Address: %s\n\tSalt: %s\n\tExpected Minipool: %s\n",
	//    minNodeFee, validatorPubkey.Hex(), validatorSignature.Hex(), depositDataRoot.Hex(), nodeAccount.Address.Hex(), GetDefaultSalt().String(), expectedMinipoolAddress.Hex())
	tx, err := node.Deposit(rp, minNodeFee, validatorPubkey, validatorSignature, depositDataRoot, salt, expectedMinipoolAddress, opts)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("Error executing deposit: %w", err)
	}
	txReceipt, err := utils.WaitForTransaction(rp.Client, tx.Hash())
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("Error waiting for deposit transaction: %w", err)
	}

	return expectedMinipoolAddress, txReceipt, nil
}
