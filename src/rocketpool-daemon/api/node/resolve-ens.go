package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	ens "github.com/wealdtech/go-ens/v3"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeResolveEnsContextFactory struct {
	handler *NodeHandler
}

func (f *nodeResolveEnsContextFactory) Create(args url.Values) (*nodeResolveEnsContext, error) {
	c := &nodeResolveEnsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
		server.GetStringFromVars("name", args, &c.name),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeResolveEnsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeResolveEnsContext, api.NodeResolveEnsData](
		router, "resolve-ens", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeResolveEnsContext struct {
	handler *NodeHandler
	address common.Address
	name    string
}

func (c *nodeResolveEnsContext) PrepareData(data *api.NodeResolveEnsData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()

	// Requirements
	err := sp.RequireEthClientSynced(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_ClientsNotSynced, err
	}

	emptyAddress := common.Address{}
	if c.address != emptyAddress {
		data.Address = c.address
		name, err := ens.ReverseResolve(rp.Client, c.address)
		if err != nil {
			data.FormattedName = data.Address.Hex()
		} else {
			data.EnsName = name
			data.FormattedName = fmt.Sprintf("%s (%s)", name, data.Address.Hex())
		}
	} else if c.name != "" {
		data.EnsName = c.name
		address, err := ens.Resolve(rp.Client, c.name)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error resolving ENS address for [%s]: %w", c.name, err)
		}
		data.Address = address
		data.FormattedName = fmt.Sprintf("%s (%s)", c.name, data.Address.Hex())
	} else {
		return types.ResponseStatus_InvalidArguments, fmt.Errorf("either address or name must not be empty")
	}

	return types.ResponseStatus_Success, nil
}
