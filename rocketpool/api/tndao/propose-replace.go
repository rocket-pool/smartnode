package tndao

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    //"github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canProposeReplace(c *cli.Context, memberAddress common.Address) (*api.CanProposeTNDAOReplaceResponse, error) {
    return nil, nil
}


func proposeReplace(c *cli.Context, memberAddress common.Address, memberId, memberEmail string) (*api.ProposeTNDAOReplaceResponse, error) {
    return nil, nil
}

