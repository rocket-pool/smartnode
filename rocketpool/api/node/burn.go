package node

import (
    "math/big"

    "github.com/urfave/cli"

    //"github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func runNodeBurn(c *cli.Context, amountWei *big.Int, token string) {
    response, err := nodeBurn(c, amountWei, token)
    if err != nil {
        api.PrintResponse(&types.NodeBurnResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func nodeBurn(c *cli.Context, amountWei *big.Int, token string) (*types.NodeBurnResponse, error) {
    return nil, nil
}

