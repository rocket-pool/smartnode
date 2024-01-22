package utils

import "github.com/rocket-pool/smartnode/shared/types"

// Check if the node wallet is ready for transacting
func IsWalletReady(status types.WalletStatus) bool {
	return status.HasAddress &&
		status.HasKeystore &&
		status.HasPassword &&
		status.NodeAddress == status.KeystoreAddress
}
