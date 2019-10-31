package node

import (
    "errors"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Initialize node password response type
type InitNodePasswordResponse struct {

    // Status
    Success bool                    `json:"success"`

    // Failure reasons
    HadExistingPassword bool        `json:"hadExistingPassword"`

}


// Initialize node account response type
type InitNodeAccountResponse struct {

    // Status
    Success bool                    `json:"success"`

    // Initialization info
    AccountAddress common.Address   `json:"accountAddress"`

    // Failure reasons
    HadExistingAccount bool         `json:"hadExistingAccount"`

}


// Check node password can be initialized
func CanInitNodePassword(p *services.Provider) *InitNodePasswordResponse {
    return &InitNodePasswordResponse{
        HadExistingPassword: p.PM.PasswordExists(),
    }
}


// Initialize node password
func InitNodePassword(p *services.Provider, password string) (*InitNodePasswordResponse, error) {

    // Set password
    if err := p.PM.SetPassword(password); err != nil {
        return nil, errors.New("Error setting node password: " + err.Error())
    }

    // Return response
    return &InitNodePasswordResponse{
        Success: true,
    }, nil

}


// Check node account can be initialized
func CanInitNodeAccount(p *services.Provider) *InitNodeAccountResponse {

    // Response
    response := &InitNodeAccountResponse{}

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
func InitNodeAccount(p *services.Provider) (*InitNodeAccountResponse, error) {

    // Create node account
    account, err := p.AM.CreateNodeAccount()
    if err != nil {
        return nil, errors.New("Error creating node account: " + err.Error())
    }

    // Return response
    return &InitNodeAccountResponse{
        Success: true,
        AccountAddress: account.Address,
    }, nil

}

