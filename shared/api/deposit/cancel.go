package deposit

import (
    "errors"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Deposit cancellation response type
type DepositCancelResponse struct {

    // Status
    Success bool                    `json:"success"`

    // Failure info
    ReservationDidNotExist bool     `json:"reservationDidNotExist"`

}


// Cancel deposit reservation
func CancelDeposit(p *services.Provider) (*DepositCancelResponse, error) {

    // Response
    response := &DepositCancelResponse{}

    // Check node has current deposit reservation
    hasReservation := new(bool)
    if err := p.NodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
        return nil, errors.New("Error retrieving deposit reservation status: " + err.Error())
    } else {
        response.ReservationDidNotExist = !*hasReservation
    }

    // Check reservation status
    if response.ReservationDidNotExist {
        return response, nil
    }

    // Cancel deposit reservation
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "depositReserveCancel"); err != nil {
            return nil, errors.New("Error canceling deposit reservation: " + err.Error())
        } else {
            response.Success = true
        }
    }

    // Return response
    return response, nil

}

