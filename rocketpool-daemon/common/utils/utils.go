package utils

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	ens "github.com/wealdtech/go-ens/v3"
)

func GetMaxApproval() *big.Int {
	// Calculate max uint256 value
	maxApproval := big.NewInt(2)
	maxApproval = maxApproval.Exp(maxApproval, big.NewInt(256), nil)
	maxApproval = maxApproval.Sub(maxApproval, big.NewInt(1))
	return maxApproval
}

// Get a formatting string containing the ENS name for an address (if it exists)
func GetFormattedAddress(ec eth.IExecutionClient, address common.Address) string {
	name, err := ens.ReverseResolve(ec, address)
	if err != nil {
		return address.Hex()
	}
	return fmt.Sprintf("%s (%s)", name, address.Hex())
}
