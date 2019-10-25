package node

import (
    "errors"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Node initialization response type
type NodeInitResponse struct {
    Success bool                        `json:"success"`
    PasswordSet bool                    `json:"passwordSet"`
    AccountCreated bool                 `json:"accountCreated"`
    AccountAddress common.Address       `json:"accountAddress"`
}


// Initialize node
func InitNode(p *services.Provider, password string) (*NodeInitResponse, error) {

    // Response
    response := &NodeInitResponse{}

    // Create password if it isn't set
    if !p.PM.PasswordExists() {
        if err := p.PM.SetPassword(password); err != nil {
            return nil, errors.New("Error setting node password: " + err.Error())
        } else {
            response.Success = true
            response.PasswordSet = true
        }
    }

    // Create node account if it doesn't exist
    if p.AM.NodeAccountExists() {
        nodeAccount, _ := p.AM.GetNodeAccount()
        response.AccountAddress = nodeAccount.Address
    } else {
        if account, err := p.AM.CreateNodeAccount(); err != nil {
            return nil, errors.New("Error creating node account: " + err.Error())
        } else {
            response.Success = true
            response.AccountCreated = true
            response.AccountAddress = account.Address
        }
    }

    // Return response
    return response, nil

}

