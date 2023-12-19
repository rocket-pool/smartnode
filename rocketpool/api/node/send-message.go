package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/core"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
	server.RegisterQuerylessGet[*nodeSendMessageContext, api.TxInfoData](
		router, "send-message", f, f.handler.serviceProvider,
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

func (c *nodeSendMessageContext) PrepareData(data *api.TxInfoData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	ec := sp.GetEthClient()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return err
	}

	info, err := core.NewTransactionInfoRaw(ec, c.address, c.message, opts)
	if err != nil {
		return fmt.Errorf("error getting transaction info: %w", err)
	}
	data.TxInfo = info
	return nil
}
