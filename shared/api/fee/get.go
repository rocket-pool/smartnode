package fee

import (
    "errors"
    "math/big"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Get user fee response type
type GetUserFeeResponse struct {
    CurrentUserFeePerc float64  `json:"currentUserFeePerc"`
    TargetUserFeePerc float64   `json:"targetUserFeePerc"`
}


// Get user fee
func GetUserFee(p *services.Provider) (*GetUserFeeResponse, error) {

    // Open database
    if err := p.DB.Open(); err != nil {
        return nil, err
    }
    defer p.DB.Close()

    // Get current user fee
    userFee := new(*big.Int)
    if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, userFee, "getFeePerc"); err != nil {
        return nil, errors.New("Error retrieving node user fee percentage setting: " + err.Error())
    }

    // Get target user fee
    targetUserFeePerc := new(float64)
    *targetUserFeePerc = -1
    if err := p.DB.Get("user.fee.target", targetUserFeePerc); err != nil {
        return nil, errors.New("Error retrieving target node user fee percentage: " + err.Error())
    }

    // Return response
    return &GetUserFeeResponse{
        CurrentUserFeePerc: eth.WeiToEth(*userFee) * 100,
        TargetUserFeePerc: *targetUserFeePerc,
    }, nil

}

