package pdao

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
)

// ===============
// === Factory ===
// ===============

type protocolDaoClearSignallingAddressFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoClearSignallingAddressFactory) Create(args url.Values) (*protocolDaoClearSignallingAddressContext, error) {
	c := &protocolDaoClearSignallingAddressContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoClearSignallingAddressFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*protocolDaoClearSignallingAddressContext, types.TxInfoData](
		router, "clear-signalling-address", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoClearSignallingAddressContext struct {
	handler *ProtocolDaoHandler
	rp      *rocketpool.RocketPool

	signallingAddress common.Address
}

func (c *protocolDaoClearSignallingAddressContext) PrepareData(data *types.TxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()
	cfg := sp.GetConfig()
	network := cfg.GetNetworkResources().Network

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}
	registry := sp.GetRocketSignerRegistry()
	if registry == nil {
		return types.ResponseStatus_Error, fmt.Errorf("Network [%v] does not have a signer registry contract", network)
	}

	// Query registry contract
	err = c.rp.Query(func(mc *batch.MultiCaller) error {
		registry.NodeToSigner(mc, &c.signallingAddress, nodeAddress)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("Error getting the registry contract: %w", err)
	}

	// Return if there if no signalling address is set
	if c.signallingAddress == (common.Address{}) {
		return types.ResponseStatus_Error, fmt.Errorf("No signalling address set")
	}
	data.TxInfo, err = registry.ClearSigner(opts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("Error getting the TX info for ClearSigner: %w", err)

	}
	return types.ResponseStatus_Success, nil
}
