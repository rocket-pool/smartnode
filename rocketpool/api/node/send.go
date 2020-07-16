package node

import (
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    //"github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func runCanNodeSend(c *cli.Context, amountWei *big.Int, token string) {
    response, err := canNodeSend(c, amountWei, token)
    if err != nil {
        api.PrintResponse(&types.CanNodeSendResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func runNodeSend(c *cli.Context, amountWei *big.Int, token string, to common.Address) {
    response, err := nodeSend(c, amountWei, token, to)
    if err != nil {
        api.PrintResponse(&types.NodeSendResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func canNodeSend(c *cli.Context, amountWei *big.Int, token string) (*types.CanNodeSendResponse, error) {
    return nil, nil
}


func nodeSend(c *cli.Context, amountWei *big.Int, token string, to common.Address) (*types.NodeSendResponse, error) {
    return nil, nil
}

