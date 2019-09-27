package deposit

import (
    "errors"
    "fmt"
    "math/big"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// RPL requirement
type RplRequirement struct {
    DurationId string
    EtherAmount *big.Int
    RplAmount *big.Int
    RplRatio *big.Int
    NetworkUtilisationPercent *big.Int
}


// Get the current deposit RPL requirement
func getRplRequired(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        LoadContracts: []string{"rocketMinipoolSettings", "rocketNodeAPI", "rocketPool"},
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

    // Staking durations to get RPL requirements for
    durations := []string{"3m", "6m", "12m"}
    durationCount := len(durations)

    // Get RPL requirements
    requirementChannels := make([]chan *RplRequirement, durationCount)
    errorChannel := make(chan error)
    for di := 0; di < durationCount; di++ {
        requirementChannels[di] = make(chan *RplRequirement)
        go (func(di int) {
            if requirement, err := getRplRequiredDuration(p, depositEtherAmountWei, durations[di]); err != nil {
                errorChannel <- err
            } else {
                requirementChannels[di] <- requirement
            }
        })(di)
    }

    // Receive RPL requirements
    requirements := make([]*RplRequirement, durationCount)
    for di := 0; di < durationCount; di++ {
        select {
            case requirement := <-requirementChannels[di]:
                requirements[di] = requirement
            case err := <-errorChannel:
                return err
        }
    }

    // Log & return
    for _, requirement := range requirements {
        fmt.Fprintln(p.Output, fmt.Sprintf(
            "Depositing %.2f ETH for %s requires %.2f RPL @ %.2f RPL / ETH. Current network utilisation for %s is %.2f%%.",
            eth.WeiToEth(requirement.EtherAmount),
            requirement.DurationId,
            eth.WeiToEth(requirement.RplAmount),
            eth.WeiToEth(requirement.RplRatio),
            requirement.DurationId,
            eth.WeiToEth(requirement.NetworkUtilisationPercent)))
    }
    return nil

}


// Get the current deposit RPL requirement for a duration
func getRplRequiredDuration(p *services.Provider, depositEtherAmountWei *big.Int, durationId string) (*RplRequirement, error) {

    // Data channels
    depositRplAmountWeiChannel := make(chan *big.Int)
    rplRatioWeiChannel := make(chan *big.Int)
    networkUtilisationPercentChannel := make(chan *big.Int)
    errorChannel := make(chan error)

    // Get RPL amount & ratio
    go (func() {
        depositRplAmountWei := new(*big.Int)
        rplRatioWei := new(*big.Int)
        out := &[]interface{}{depositRplAmountWei, rplRatioWei}
        if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, out, "getRPLRequired", depositEtherAmountWei, durationId); err != nil {
            errorChannel <- errors.New("Error retrieving required RPL amount: " + err.Error())
        } else {
            depositRplAmountWeiChannel <- *depositRplAmountWei
            rplRatioWeiChannel <- *rplRatioWei
        }
    })()

    // Get network utilisation
    go (func() {
        networkUtilisation := new(*big.Int)
        if err := p.CM.Contracts["rocketPool"].Call(nil, networkUtilisation, "getNetworkUtilisation", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving network utilisation: " + err.Error())
        } else {
            networkUtilisationPercent := new(big.Int)
            networkUtilisationPercent.Mul(*networkUtilisation, big.NewInt(100))
            networkUtilisationPercentChannel <- networkUtilisationPercent
        }
    })()

    // Receive data
    var depositRplAmountWei *big.Int
    var rplRatioWei *big.Int
    var networkUtilisationPercent *big.Int
    for received := 0; received < 3; {
        select {
            case depositRplAmountWei = <-depositRplAmountWeiChannel:
                received++
            case rplRatioWei = <-rplRatioWeiChannel:
                received++
            case networkUtilisationPercent = <-networkUtilisationPercentChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return &RplRequirement{
        DurationId: durationId,
        EtherAmount: depositEtherAmountWei,
        RplAmount: depositRplAmountWei,
        RplRatio: rplRatioWei,
        NetworkUtilisationPercent: networkUtilisationPercent,
    }, nil

}

