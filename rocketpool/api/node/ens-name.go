package node

import (
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
	ens "github.com/wealdtech/go-ens/v3"
)

func resolveEnsName(c *cli.Context, name string) (*api.ResolveEnsNameResponse, error) {
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	address, err := ens.Resolve(rp.Client, name)
	if err != nil {
		return nil, err
	}
	response := api.ResolveEnsNameResponse{}
	response.Address = address
	return &response, nil
}
