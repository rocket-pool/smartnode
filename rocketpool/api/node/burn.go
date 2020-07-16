package node

import (
    "math/big"

    "github.com/urfave/cli"

    //"github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func runCanNodeBurn(c *cli.Context, amountWei *big.Int, token string) {
    response, err := canNodeBurn(c, amountWei, token)
    if err != nil {
        api.PrintResponse(&types.CanNodeBurnResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func runNodeBurn(c *cli.Context, amountWei *big.Int, token string) {
    response, err := nodeBurn(c, amountWei, token)
    if err != nil {
        api.PrintResponse(&types.NodeBurnResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func canNodeBurn(c *cli.Context, amountWei *big.Int, token string) (*types.CanNodeBurnResponse, error) {
    return nil, nil
}


func nodeBurn(c *cli.Context, amountWei *big.Int, token string) (*types.NodeBurnResponse, error) {
    return nil, nil
}

