package node

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
	"github.com/rocket-pool/rocketpool-go/tests/testutils/validator"
)

// Returns the default salt for minipool address generation
func GetDefaultSalt() *big.Int {
    return big.NewInt(204)
}


// Call deposit on the node using the validator test values
func Deposit(rp *rocketpool.RocketPool, nodeAccount *accounts.Account, depositAmount *big.Int) (common.Address, *types.Receipt, error) {

    // Get validator & deposit data
    validatorPubkey, err := validator.GetValidatorPubkey()
    if err != nil { return common.Address{}, nil, err }
    expectedMinipoolAddress, err := utils.GenerateAddress(rp, nodeAccount.Address, rptypes.Half, GetDefaultSalt(), nil)
    if err != nil { return common.Address{}, nil, err }
    withdrawalCredentials := utils.GetWithdrawalCredentials(expectedMinipoolAddress)
    validatorSignature, err := validator.GetValidatorSignature()
    if err != nil { return common.Address{}, nil, err }
    depositDataRoot, err := validator.GetDepositDataRoot(validatorPubkey, withdrawalCredentials, validatorSignature)
    if err != nil { return common.Address{}, nil, err }

    // Make node deposit
    opts := nodeAccount.GetTransactor()
    opts.Value = depositAmount
    hash, err := node.Deposit(rp, 0, validatorPubkey, validatorSignature, depositDataRoot, GetDefaultSalt(), expectedMinipoolAddress, opts)
    if err != nil { return common.Address{}, nil, err }
    txReceipt, err := utils.WaitForTransaction(rp.Client, hash)
    if err != nil { return common.Address{}, nil, err }

    return expectedMinipoolAddress, txReceipt, nil
}

