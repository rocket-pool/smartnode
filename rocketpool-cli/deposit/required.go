package deposit

import (
    "errors"
    "fmt"
    "math/big"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Get the current deposit RPL requirement
func getRplRequired(c *cli.Context, durationId string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        LoadContracts: []string{"rocketMinipoolSettings", "rocketNodeAPI"},
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get minipool launch ether amount
    launchEtherAmountWei := new(*big.Int)
    if err := p.CM.Contracts["rocketMinipoolSettings"].Call(nil, launchEtherAmountWei, "getMinipoolLaunchAmount"); err != nil {
        return errors.New("Error retrieving minipool launch amount: " + err.Error())
    }

    // Get deposit ether amount
    depositEtherAmountWei := new(big.Int)
    depositEtherAmountWei.Quo(*launchEtherAmountWei, big.NewInt(2))

    // Get RPL required
    depositRplAmountWei := new(*big.Int)
    rplRatioWei := new(*big.Int)
    out := &[]interface{}{depositRplAmountWei, rplRatioWei}
    if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, out, "getRPLRequired", depositEtherAmountWei, durationId); err != nil {
        return errors.New("Error retrieving required RPL amount: " + err.Error())
    }

    // Log & return
    fmt.Fprintln(p.Output, fmt.Sprintf(
        "%.2f RPL required to cover a deposit amount of %.2f ETH for %s @ %.2f RPL / ETH",
        eth.WeiToEth(*depositRplAmountWei),
        eth.WeiToEth(depositEtherAmountWei),
        durationId,
        eth.WeiToEth(*rplRatioWei)))
    return nil

}

