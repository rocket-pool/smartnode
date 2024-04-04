package network

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
)

// ===============
// === Factory ===
// ===============

type networkGenerateRewardsContextFactory struct {
	handler *NetworkHandler
}

func (f *networkGenerateRewardsContextFactory) Create(args url.Values) (*networkGenerateRewardsContext, error) {
	c := &networkGenerateRewardsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("index", args, input.ValidateUint, &c.index),
	}
	return c, errors.Join(inputErrs...)
}

func (f *networkGenerateRewardsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*networkGenerateRewardsContext, types.SuccessData](
		router, "generate-rewards-tree", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkGenerateRewardsContext struct {
	handler *NetworkHandler

	index uint64
}

func (c *networkGenerateRewardsContext) PrepareData(data *types.SuccessData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()

	// Create the generation request
	requestPath := cfg.GetRegenerateRewardsTreeRequestPath(c.index)
	requestFile, err := os.Create(requestPath)
	if requestFile != nil {
		requestFile.Close()
	}
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating request marker: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
