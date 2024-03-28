package node

import "math/big"

func getMaxApproval() *big.Int {
	// Calculate max uint256 value
	maxApproval := big.NewInt(2)
	maxApproval = maxApproval.Exp(maxApproval, big.NewInt(256), nil)
	maxApproval = maxApproval.Sub(maxApproval, big.NewInt(1))
	return maxApproval
}
