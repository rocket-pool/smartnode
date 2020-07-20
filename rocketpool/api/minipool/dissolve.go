package minipool

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    //"github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canDissolveMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanDissolveMinipoolResponse, error) {
    return nil, nil
}


func dissolveMinipool(c *cli.Context, minipoolAddress common.Address) (*api.DissolveMinipoolResponse, error) {
    return nil, nil
}

