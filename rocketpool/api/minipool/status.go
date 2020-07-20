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
            Address: minipoolDetails.Address.Hex(),
            ValidatorPubkey: hex.EncodeToString(minipoolDetails.ValidatorPubkey),
            Status: minipoolDetails.Status.Status,
            StatusBlock: int(minipoolDetails.Status.StatusBlock),
            StatusTime: minipoolDetails.Status.StatusTime,
            DepositType: minipoolDetails.DepositType,
            NodeFee: minipoolDetails.Node.Fee,
            NodeDepositBalance: minipoolDetails.Node.DepositBalance.String(),
            NodeRefundBalance: minipoolDetails.Node.RefundBalance.String(),
            NethBalance: minipoolDetails.NethBalance.String(),
            UserDepositBalance: minipoolDetails.User.DepositBalance.String(),
            UserDepositAssigned: minipoolDetails.User.DepositAssigned,
            StakingStartBalance: minipoolDetails.Staking.StartBalance.String(),
            StakingEndBalance: minipoolDetails.Staking.EndBalance.String(),
            StakingStartBlock: int(minipoolDetails.Staking.StartBlock),
            StakingUserStartBlock: int(minipoolDetails.Staking.UserStartBlock),
            StakingEndBlock: int(minipoolDetails.Staking.EndBlock),
        }
    }

    // Return response
    return &response, nil

}
