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

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}
	registry := sp.GetRocketSignerRegistry()
	if registry == nil {
		return types.ResponseStatus_ResourceNotFound, err
	}

	// Snapshot Registry
	err = c.rp.Query(func(mc *batch.MultiCaller) error {
		registry.NodeToSigner(mc, &c.signallingAddress, nodeAddress)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting the registry contract: %w", err)
	}

	// Return if there if no signalling address is set
	if c.signallingAddress == (common.Address{}) {
		return types.ResponseStatus_Error, nil
	} else {
		data.TxInfo, err = registry.ClearSigner(opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("Error getting the TX info for ClearSigner: %w", err)
		}
	}
	return types.ResponseStatus_Success, nil
}

// was initially trying it this way before I decided using server.RegisterQuerylessGet would be more appropriate

// package pdao

// import (
// 	"fmt"
// 	"net/url"

// 	"github.com/ethereum/go-ethereum/accounts/abi/bind"
// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/gorilla/mux"
// 	batch "github.com/rocket-pool/batch-query"
// 	"github.com/rocket-pool/node-manager-core/api/server"
// 	"github.com/rocket-pool/node-manager-core/api/types"
// 	"github.com/rocket-pool/rocketpool-go/v2/node"
// 	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
// 	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/contracts"
// 	"github.com/rocket-pool/smartnode/v2/shared/types/api"
// )

// // ===============
// // === Factory ===
// // ===============

// type protocolDaoClearSignallingAddressFactory struct {
// 	handler *ProtocolDaoHandler
// }

// func (f *protocolDaoClearSignallingAddressFactory) Create(args url.Values) (*protocolDaoClearSignallingAddressContext, error) {
// 	c := &protocolDaoClearSignallingAddressContext{
// 		handler: f.handler,
// 	}
// 	return c, nil
// }

// func (f *protocolDaoClearSignallingAddressFactory) RegisterRoute(router *mux.Router) {
// 	server.RegisterSingleStageRoute[*protocolDaoClearSignallingAddressContext, api.ProtocolDaoClearSignallingAddressResponse](
// 		router, "clear-signalling-address", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
// 	)
// }

// // ===============
// // === Context ===
// // ===============

// type protocolDaoClearSignallingAddressContext struct {
// 	handler  *ProtocolDaoHandler
// 	rp       *rocketpool.RocketPool
// 	registry *contracts.RocketSignerRegistry

// 	node              *node.Node
// 	signallingAddress common.Address
// }

// func (c *protocolDaoClearSignallingAddressContext) Initialize() (types.ResponseStatus, error) {
// 	sp := c.handler.serviceProvider
// 	c.rp = sp.GetRocketPool()
// 	nodeAddress, _ := sp.GetWallet().GetAddress()

// 	// Requirements
// 	err := sp.RequireNodeAddress()
// 	if err != nil {
// 		return types.ResponseStatus_AddressNotPresent, err
// 	}
// 	c.registry = sp.GetRocketSignerRegistry()
// 	if c.registry == nil {
// 		return types.ResponseStatus_ResourceNotFound, err
// 	}

// 	// Bindings
// 	c.node, err = node.NewNode(c.rp, nodeAddress)
// 	if err != nil {
// 		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
// 	}

// 	return types.ResponseStatus_Success, nil
// }

// func (c *protocolDaoClearSignallingAddressContext) GetState(mc *batch.MultiCaller) {

// 	// Snapshot Registry
// 	if c.registry != nil {
// 		c.registry.NodeToSigner(mc, &c.signallingAddress, c.node.Address)
// 	}

// }

// func (c *protocolDaoClearSignallingAddressContext) PrepareData(data *api.ProtocolDaoClearSignallingAddressResponse, opts *bind.TransactOpts) (types.ResponseStatus, error) {

// 	// Return if there if no signalling address is set
// 	if c.signallingAddress == (common.Address{}) {
// 		data.CanClear = false
// 		return types.ResponseStatus_Error, nil
// 	} else {
// 		data.CanClear = true
// 	}

// 	if data.CanClear {
// 		txInfo, err := c.registry.ClearSigner(opts)
// 		if err != nil {
// 			return types.ResponseStatus_Error, fmt.Errorf("Error getting the TX info for ClearSigner: %w", err)
// 		}
// 		data.TxInfo = txInfo
// 	}
// 	return types.ResponseStatus_Success, nil
// }
