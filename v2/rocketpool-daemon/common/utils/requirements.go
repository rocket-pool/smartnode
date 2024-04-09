package utils

import (
	"errors"

	"github.com/rocket-pool/node-manager-core/wallet"
)

func CheckIfWalletReady(status wallet.WalletStatus) error {
	if !status.Address.HasAddress {
		return errors.New("The node currently does not have an address set. Please run 'rocketpool wallet init' and try again.")
	}
	if !status.Wallet.IsLoaded {
		if status.Wallet.IsOnDisk {
			if !status.Password.IsPasswordSaved {
				return errors.New("The node has a node wallet on disk but does not have the password for it loaded. Please run `rocketpool wallet set-password` to load it.")
			}
			return errors.New("The node has a node wallet and a password on disk but there was an error loading it - perhaps the password is incorrect? Please check the node logs for more information.")
		}
		return errors.New("The node currently does not have a node wallet keystore. Please run 'rocketpool wallet init' and try again.")
	}
	if status.Wallet.WalletAddress != status.Address.NodeAddress {
		return errors.New("The node's wallet keystore does not match the node address. This node is currently in read-only mode.")
	}
	return nil
}
