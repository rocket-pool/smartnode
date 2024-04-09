package wallet

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
	ens "github.com/wealdtech/go-ens/v3"
)

// ===============
// === Factory ===
// ===============

type walletSetEnsNameContextFactory struct {
	handler *WalletHandler
}

func (f *walletSetEnsNameContextFactory) Create(args url.Values) (*walletSetEnsNameContext, error) {
	c := &walletSetEnsNameContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.GetStringFromVars("name", args, &c.name),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletSetEnsNameContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletSetEnsNameContext, api.WalletSetEnsNameData](
		router, "set-ens-name", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletSetEnsNameContext struct {
	handler *WalletHandler
	name    string
}

func (c *walletSetEnsNameContext) PrepareData(data *api.WalletSetEnsNameData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	ec := sp.GetEthClient()
	txMgr := sp.GetTransactionManager()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}

	// Name validation
	if c.name == "" {
		return types.ResponseStatus_InvalidArguments, fmt.Errorf("name cannot be blank")
	}

	// The ENS name must resolve to the wallet address
	resolvedAddress, err := ens.Resolve(ec, c.name)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error resolving '%s' to an address: %w", c.name, err)
	}

	if resolvedAddress != nodeAddress {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("%s currently resolves to the address %s instead of the node wallet address %s", c.name, resolvedAddress.Hex(), nodeAddress.Hex())
	}

	// Check if the name is already in use
	resolvedName, err := ens.ReverseResolve(ec, nodeAddress)
	if err != nil && err.Error() != "not a resolver" {
		// Handle errors unrelated to the address not being an ENS resolver
		return types.ResponseStatus_Error, fmt.Errorf("error reverse resolving %s to an ENS name: %w", nodeAddress.Hex(), err)
	} else if resolvedName == c.name {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("the ENS record already points to the name '%s'", c.name)
	}

	// Get the raw TX from the ENS lib
	registrar, err := ens.NewReverseRegistrar(ec)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating reverse registrar binding: %w", err)
	}
	opts.NoSend = true
	tx, err := registrar.SetName(opts, c.name)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error constructing SetName TX: %w", err)
	}

	// Derive the TXInfo
	txInfo := txMgr.CreateTransactionInfoRaw(*tx.To(), tx.Data(), opts)
	data.TxInfo = txInfo
	return types.ResponseStatus_Success, nil
}
