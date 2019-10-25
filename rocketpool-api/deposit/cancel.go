package deposit

import (
    "errors"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Cancel the reserved node deposit
func cancelDeposit(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Response
    response := api.DepositCancelResponse{}

    // Check node has current deposit reservation
    hasReservation := new(bool)
    if err := p.NodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
        return errors.New("Error retrieving deposit reservation status: " + err.Error())
    } else {
        response.HadExistingReservation = *hasReservation
    }

    // Check reservation status
    if !response.HadExistingReservation {
        api.PrintResponse(p.Output, response)
        return nil
    }

    // Cancel deposit reservation
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "depositReserveCancel"); err != nil {
            return errors.New("Error canceling deposit reservation: " + err.Error())
        } else {
            response.Success = true
        }
    }

    // Print response & return
    api.PrintResponse(p.Output, response)
    return nil

}

