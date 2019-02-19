package deposit

import (
    "errors"
    "fmt"
    "math/big"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Get the current deposit RPL requirement
func getRplRequired(c *cli.Context, durationId string) error {

    // Command setup
    if message, err := setup(c, []string{"rocketMinipoolSettings", "rocketNodeAPI"}, []string{"rocketNodeContract"}); message != "" {
        fmt.Println(message)
        return nil
    } else if err != nil {
        return err
    }

    // Get minipool launch ether amount
    launchEtherAmountWei := new(*big.Int)
    if err := cm.Contracts["rocketMinipoolSettings"].Call(nil, launchEtherAmountWei, "getMinipoolLaunchAmount"); err != nil {
        return errors.New("Error retrieving minipool launch amount: " + err.Error())
    }

    // Get deposit ether amount
    depositEtherAmountWei := new(big.Int)
    depositEtherAmountWei.Quo(*launchEtherAmountWei, big.NewInt(2))

    // Get RPL required
    var depositRplAmountWei = new(*big.Int)
    var rplRatioWei = new(*big.Int)
    out := &[]interface{}{depositRplAmountWei, rplRatioWei}
    if err := cm.Contracts["rocketNodeAPI"].Call(nil, out, "getRPLRequired", depositEtherAmountWei, durationId); err != nil {
        return errors.New("Error retrieving required RPL amount: " + err.Error())
    }

    // Log & return
    fmt.Println(fmt.Sprintf(
        "%.2f RPL required to cover a deposit amount of %.2f ETH for %s @ %.2f RPL / ETH",
        eth.WeiToEth(*depositRplAmountWei),
        eth.WeiToEth(depositEtherAmountWei),
        durationId,
        eth.WeiToEth(*rplRatioWei)))
    return nil

}

