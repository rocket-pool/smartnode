package node

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"
	ens "github.com/wealdtech/go-ens/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func resolveEnsName(c *cli.Command, name string) (*api.ResolveEnsNameResponse, error) {
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

func reverseResolveEnsName(c *cli.Command, address common.Address) (*api.ResolveEnsNameResponse, error) {
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

func formatResolvedAddress(c *cli.Command, address common.Address) string {
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

func resolveEnsNameHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		resp, err := resolveEnsName(c, name)
		response.WriteResponse(w, resp, err)
	}
}

func reverseResolveEnsNameHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addr := common.HexToAddress(r.URL.Query().Get("address"))
		resp, err := reverseResolveEnsName(c, addr)
		response.WriteResponse(w, resp, err)
	}
}
