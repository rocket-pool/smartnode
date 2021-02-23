package tndao

import (
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    //"github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canProposeKick(c *cli.Context, memberAddress common.Address, fineAmountWei *big.Int) (*api.CanProposeTNDAOKickResponse, error) {
    return nil, nil
}


func proposeKick(c *cli.Context, memberAddress common.Address, fineAmountWei *big.Int) (*api.ProposeTNDAOKickResponse, error) {
    return nil, nil
}

