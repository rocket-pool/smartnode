package deposit

import (
    "errors"
    "math/big"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Deposit required response type
type DepositRequiredResponse struct {
    Durations []*DurationRequirement    `json:"durations"`
}
type DurationRequirement struct {
    DurationId string                   `json:"durationId"`
    EtherAmountWei *big.Int             `json:"etherAmountWei"`
    RplAmountWei *big.Int               `json:"rplAmountWei"`
    RplRatioWei *big.Int                `json:"rplRatioWei"`
    NetworkUtilisation float64          `json:"networkUtilisation"`
}


// Get deposit RPL requirement
func GetRplRequired(p *services.Provider) (*DepositRequiredResponse, error) {

    // Response
    response := &DepositRequiredResponse{}

    // Get minipool launch ether amount
    launchEtherAmountWei := new(*big.Int)
    if err := p.CM.Contracts["rocketMinipoolSettings"].Call(nil, launchEtherAmountWei, "getMinipoolLaunchAmount"); err != nil {
        return nil, errors.New("Error retrieving minipool launch amount: " + err.Error())
    }

    // Get deposit ether amount
    depositEtherAmountWei := new(big.Int)
    depositEtherAmountWei.Quo(*launchEtherAmountWei, big.NewInt(2))

    // Staking durations to get RPL requirements for
    durations := []string{"3m", "6m", "12m"}
    durationCount := len(durations)

    // Get duration requirements
    requirementChannels := make([]chan *DurationRequirement, durationCount)
    errorChannel := make(chan error)
    for di := 0; di < durationCount; di++ {
        requirementChannels[di] = make(chan *DurationRequirement)
        go (func(di int) {
            if requirement, err := getRplRequiredDuration(p, depositEtherAmountWei, durations[di]); err != nil {
                errorChannel <- err
            } else {
                requirementChannels[di] <- requirement
            }
        })(di)
    }

    // Receive duration requirements
    response.Durations = make([]*DurationRequirement, durationCount)
    for di := 0; di < durationCount; di++ {
        select {
            case response.Durations[di] = <-requirementChannels[di]:
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return response
    return response, nil

}


// Get the current deposit RPL requirement for a duration
func getRplRequiredDuration(p *services.Provider, depositEtherAmountWei *big.Int, durationId string) (*DurationRequirement, error) {

    // Data channels
    depositRplAmountWeiChannel := make(chan *big.Int)
    rplRatioWeiChannel := make(chan *big.Int)
    networkUtilisationChannel := make(chan float64)
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
            networkUtilisationChannel <- eth.WeiToEth(networkUtilisation)
        }
    })()

    // Receive data
    var depositRplAmountWei *big.Int
    var rplRatioWei *big.Int
    var networkUtilisation float64
    for received := 0; received < 3; {
        select {
            case depositRplAmountWei = <-depositRplAmountWeiChannel:
                received++
            case rplRatioWei = <-rplRatioWeiChannel:
                received++
            case networkUtilisation = <-networkUtilisationChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return &DurationRequirement{
        DurationId: durationId,
        EtherAmountWei: depositEtherAmountWei,
        RplAmountWei: depositRplAmountWei,
        RplRatioWei: rplRatioWei,
        NetworkUtilisation: networkUtilisation,
    }, nil

}

