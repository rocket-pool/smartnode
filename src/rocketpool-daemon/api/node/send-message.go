package node

import (
	"errors"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
)

// ===============
// === Factory ===
// ===============

type nodeSendMessageContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSendMessageContextFactory) Create(args url.Values) (*nodeSendMessageContext, error) {
	c := &nodeSendMessageContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
		server.ValidateArg("message", args, input.ValidateByteArray, &c.message),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSendMessageContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeSendMessageContext, types.TxInfoData](
		router, "send-message", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSendMessageContext struct {
	handler *NodeHandler
	address common.Address
	message []byte
}

func (c *nodeSendMessageContext) PrepareData(data *types.TxInfoData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	txMgr := sp.GetTransactionManager()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return err
	}

	info := txMgr.CreateTransactionInfoRaw(c.address, c.message, opts)
	data.TxInfo = info
	return nil
}
