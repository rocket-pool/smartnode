package pdao

import (
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type protocolDaoSetSignallingAddressFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoSetSignallingAddressFactory) Create(args url.Values) (*protocolDaoSetSignallingAddressContext, error) {
	c := &protocolDaoSetSignallingAddressContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoSetSignallingAddressFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoSetSignallingAddressContext, api.ProtocolDaoSetSignallingAddressResponse](
		router, "set-signalling-address", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoSetSignallingAddressContext struct {
	handler *ProtocolDaoHandler
}

func (c *protocolDaoSetSignallingAddressContext) Initialize() (types.ResponseStatus, error) {
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoSetSignallingAddressContext) GetState(mc *batch.MultiCaller) {
}

func (c *protocolDaoSetSignallingAddressContext) PrepareData(data *api.ProtocolDaoSetSignallingAddressResponse, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	return types.ResponseStatus_Success, nil
}
