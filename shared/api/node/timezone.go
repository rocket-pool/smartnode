package node

import (
    "errors"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Node timezone response type
type NodeTimezoneResponse struct {

    // Status
    Success bool        `json:"success"`

    // Timezone info
    Timezone string     `json:"timezone"`

}


// Set node timezone
func SetNodeTimezone(p *services.Provider, timezone string) (*NodeTimezoneResponse, error) {

    // Response
    response := &NodeTimezoneResponse{}

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()

    // Set node timezone
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses["rocketNodeAPI"], p.CM.Abis["rocketNodeAPI"], "setTimezoneLocation", timezone); err != nil {
            return nil, errors.New("Error setting node timezone: " + err.Error())
        } else {
            response.Success = true
        }
    }

    // Get node timezone
    nodeTimezone := new(string)
    if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeTimezone, "getTimezoneLocation", nodeAccount.Address); err != nil {
        return nil, errors.New("Error retrieving node timezone: " + err.Error())
    } else {
        response.Timezone = *nodeTimezone
    }

    // Return response
    return response, nil

}

