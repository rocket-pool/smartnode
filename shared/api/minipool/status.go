package minipool

import (
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
)


// Minipool status response type
type MinipoolStatusResponse struct {
    Minipools []*minipool.Details   `json:"minipools"`
}


// Get minipool statuses
func GetMinipoolStatus(p *services.Provider) (*MinipoolStatusResponse, error) {

    // Response
    response := &MinipoolStatusResponse{}

    // Get minipool addresses
    nodeAccount, _ := p.AM.GetNodeAccount()
    minipoolAddresses, err := node.GetMinipoolAddresses(nodeAccount.Address, p.CM)
    if err != nil {
        return nil, err
    }
    minipoolCount := len(minipoolAddresses)

    // Get minipool details
    detailsChannels := make([]chan *minipool.Details, minipoolCount)
    errorChannel := make(chan error)
    for mi := 0; mi < minipoolCount; mi++ {
        detailsChannels[mi] = make(chan *minipool.Details)
        go (func(mi int) {
            if details, err := minipool.GetDetails(p.CM, minipoolAddresses[mi]); err != nil {
                errorChannel <- err
            } else {
                detailsChannels[mi] <- details
            }
        })(mi)
    }

    // Receive minipool details
    response.Minipools = make([]*minipool.Details, minipoolCount)
    for mi := 0; mi < minipoolCount; mi++ {
        select {
            case details := <-detailsChannels[mi]:
                response.Minipools[mi] = details
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return response
    return response, nil

}

