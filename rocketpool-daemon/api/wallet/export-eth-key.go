package wallet

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type walletExportEthKeyContextFactory struct {
	handler *WalletHandler
}

func (f *walletExportEthKeyContextFactory) Create(args url.Values) (*walletExportEthKeyContext, error) {
	c := &walletExportEthKeyContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *walletExportEthKeyContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletExportEthKeyContext, api.WalletExportEthKeyData](
		router, "export-eth-key", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletExportEthKeyContext struct {
	handler *WalletHandler
}

func (c *walletExportEthKeyContext) PrepareData(data *api.WalletExportEthKeyData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	// Requirements
	err := sp.RequireWalletReady()
	if err != nil {
		return err
	}

	// Get password
	pw, isSet := w.GetPassword()
	if !isSet {
		return fmt.Errorf("password has not been set; cannot decrypt wallet keystore without it")
	}

	privateKeyBytes := w.GetNodePrivateKeyBytes()
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return fmt.Errorf("error decoding wallet private key: %w", err)
	}

	// Create an ETH key from the data
	key := &keystore.Key{
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
		Id:         uuid.UUID(w.GetWalletUuidBytes()),
	}
	ethkey, err := keystore.EncryptKey(key, string(pw), keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return fmt.Errorf("error converting wallet private key: %w", err)
	}
	data.EthKeyJson = ethkey
	return nil
}
