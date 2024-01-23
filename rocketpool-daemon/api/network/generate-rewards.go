package network

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
	server.RegisterQuerylessGet[*networkGenerateRewardsContext, api.SuccessData](
		router, "generate-rewards-tree", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkGenerateRewardsContext struct {
	handler *NetworkHandler

	index uint64
}

func (c *networkGenerateRewardsContext) PrepareData(data *api.SuccessData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()

	// Create the generation request
	requestPath := cfg.Smartnode.GetRegenerateRewardsTreeRequestPath(c.index, true)
	requestFile, err := os.Create(requestPath)
	if requestFile != nil {
		requestFile.Close()
	}
	if err != nil {
		return fmt.Errorf("error creating request marker: %w", err)
	}
	data.Success = true
	return nil
}
