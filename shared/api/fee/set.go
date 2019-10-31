package fee

import (
    "errors"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Set target user fee response type
type SetTargetUserFeeResponse struct {
    Success bool    `json:"success"`
}


// Set target user fee
func SetTargetUserFee(p *services.Provider, feePercent float64) (*SetTargetUserFeeResponse, error) {

    // Open database
    if err := p.DB.Open(); err != nil {
        return nil, err
    }
    defer p.DB.Close()

    // Set target user fee percent
    if err := p.DB.Put("user.fee.target", feePercent); err != nil {
        return nil, errors.New("Error setting target user fee percentage: " + err.Error())
    }

    // Return response
    return &SetTargetUserFeeResponse{
        Success: true,
    }, nil

}

