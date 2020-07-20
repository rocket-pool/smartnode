package minipool

import (
    "encoding/hex"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func getStatus(c *cli.Context) (*api.MinipoolStatusResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.MinipoolStatusResponse{}

    // Get minipool details
    nodeAccount, _ := am.GetNodeAccount()
    details, err := getNodeMinipoolDetails(rp, nodeAccount.Address)
    if err != nil {
        return nil, err
    }

    // Update response
    response.Minipools = make([]api.MinipoolDetails, len(details))
    for mi, minipoolDetails := range details {
        response.Minipools[mi] = api.MinipoolDetails{
            Address:                minipoolDetails.Address,
            ValidatorPubkey:        hex.EncodeToString(minipoolDetails.ValidatorPubkey),
            Status:                 minipoolDetails.Status.Status,
            StatusBlock:            minipoolDetails.Status.StatusBlock,
            StatusTime:             minipoolDetails.Status.StatusTime,
            DepositType:            minipoolDetails.DepositType,
            NodeFee:                minipoolDetails.Node.Fee,
            NodeDepositBalance:     minipoolDetails.Node.DepositBalance,
            NodeRefundBalance:      minipoolDetails.Node.RefundBalance,
            NethBalance:            minipoolDetails.NethBalance,
            UserDepositBalance:     minipoolDetails.User.DepositBalance,
            UserDepositAssigned:    minipoolDetails.User.DepositAssigned,
            StakingStartBalance:    minipoolDetails.Staking.StartBalance,
            StakingEndBalance:      minipoolDetails.Staking.EndBalance,
            StakingStartBlock:      minipoolDetails.Staking.StartBlock,
            StakingUserStartBlock:  minipoolDetails.Staking.UserStartBlock,
            StakingEndBlock:        minipoolDetails.Staking.EndBlock,
        }
    }

    // Return response
    return &response, nil

}
