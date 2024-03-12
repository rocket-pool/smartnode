package data

import (
	"io/fs"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet/keystore"
)

const (
	walletAddressFileMode  fs.FileMode = 0664
	walletKeystoreFileMode fs.FileMode = 0600
	passwordFileMode       fs.FileMode = 0600
)

// Creates a new wallet address manager
func NewAddressManager(walletAddressPath string) *DataManager[common.Address] {
	return NewDataManager[common.Address]("wallet address", walletAddressPath, walletAddressFileMode, walletAddressSerializer{})
}

// Creates a new wallet keystore manager
func NewKeystoreManager(walletKeystorePath string) *DataManager[*keystore.WalletKeystore] {
	return NewDataManager[*keystore.WalletKeystore]("wallet keystore", walletKeystorePath, walletKeystoreFileMode, walletKeystoreSerializer{})
}

// Creates a new wallet keystore password manager
func NewPasswordManager(passwordFilePath string) *DataManager[[]byte] {
	return NewDataManager[[]byte]("password", passwordFilePath, passwordFileMode, passwordSerializer{})
}
