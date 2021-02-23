package tndao

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    //"github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canProposeInvite(c *cli.Context, memberAddress common.Address) (*api.CanProposeTNDAOInviteResponse, error) {
    return nil, nil
}


func proposeInvite(c *cli.Context, memberAddress common.Address, memberId, memberEmail string) (*api.ProposeTNDAOInviteResponse, error) {
    return nil, nil
}

