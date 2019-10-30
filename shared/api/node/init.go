package node

import (
    "errors"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Node password initialization response type
type NodePasswordInitResponse struct {

    // Status
    Success bool                    `json:"success"`

    // Failure info
    HadExistingPassword bool        `json:"hadExistingPassword"`

}


// Node account initialization response type
type NodeAccountInitResponse struct {

    // Status
    Success bool                    `json:"success"`

    // Initialization info
    AccountAddress common.Address   `json:"accountAddress"`

    // Failure info
    HadExistingAccount bool         `json:"hadExistingAccount"`

}


// Check node password can be initialized
func CanInitNodePassword(p *services.Provider) *NodePasswordInitResponse {
    return &NodePasswordInitResponse{
        HadExistingPassword: p.PM.PasswordExists(),
    }
}


// Initialize node password
func InitNodePassword(p *services.Provider, password string) (*NodePasswordInitResponse, error) {

    // Set password
    if err := p.PM.SetPassword(password); err != nil {
        return nil, errors.New("Error setting node password: " + err.Error())
    }

    // Return response
    return &NodePasswordInitResponse{
        Success: true,
    }, nil

}


// Check node account can be initialized
func CanInitNodeAccount(p *services.Provider) *NodeAccountInitResponse {

    // Response
    response := &NodeAccountInitResponse{}

    // Check if node account already exists
    if p.AM.NodeAccountExists() {
        nodeAccount, _ := p.AM.GetNodeAccount()
        response.AccountAddress = nodeAccount.Address
        response.HadExistingAccount = true
    }

    // Return response
    return response

}


// Initialize node account
func InitNodeAccount(p *services.Provider) (*NodeAccountInitResponse, error) {

    // Create node account
    account, err := p.AM.CreateNodeAccount()
    if err != nil {
        return nil, errors.New("Error creating node account: " + err.Error())
    }

    // Return response
    return &NodeAccountInitResponse{
        Success: true,
        AccountAddress: account.Address,
    }, nil

}

