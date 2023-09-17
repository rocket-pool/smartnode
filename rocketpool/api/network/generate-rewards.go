package network

import (
	"errors"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// ===============
// === Factory ===
// ===============

type networkGenerateRewardsContextFactory struct {
	handler *NetworkHandler
}

func (f *networkGenerateRewardsContextFactory) Create(vars map[string]string) (*networkGenerateRewardsContext, error) {
	c := &networkGenerateRewardsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("index", vars, cliutils.ValidateUint, &c.index),
	}
	return c, errors.Join(inputErrs...)
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
