package node

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
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
	response := api.ResolveEnsNameResponse{
		Address: address,
		EnsName: name,
	}
	return &response, nil
}

func reverseResolveEnsName(c *cli.Context, address common.Address) (*api.ResolveEnsNameResponse, error) {
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	name, err := ens.ReverseResolve(rp.Client, address)
	if err != nil {
		return nil, err
	}
	response := api.ResolveEnsNameResponse{
		Address: address,
		EnsName: name,
	}
	return &response, nil
}

func formatResolvedAddress(c *cli.Context, address common.Address) string {
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return address.Hex()
	}

	name, err := ens.ReverseResolve(rp.Client, address)
	if err != nil {
		return address.Hex()
	}
	return fmt.Sprintf("%s (%s)", name, address.Hex())
}
