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
	server.RegisterSingleStageRoute[*protocolDaoClearSignallingAddressContext, api.ProtocolDAOClearSignallingAddressResponse](
		router, "set-signalling-address", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoClearSignallingAddressContext struct {
	handler *ProtocolDaoHandler
}

func (c *protocolDaoClearSignallingAddressContext) Initialize() (types.ResponseStatus, error) {
	return types.ResponseStatus_Success, nil
}

func (c *protocolDaoClearSignallingAddressContext) GetState(mc *batch.MultiCaller) {
}

func (c *protocolDaoClearSignallingAddressContext) PrepareData(data *api.ProtocolDAOClearSignallingAddressResponse, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	return types.ResponseStatus_Success, nil
}
