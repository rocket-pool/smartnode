package deposit

import (
    "errors"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Cancel deposit response types
type CanCancelDepositResponse struct {

    // Status
    Success bool                    `json:"success"`

    // Failure reasons
    ReservationDidNotExist bool     `json:"reservationDidNotExist"`

}
type CancelDepositResponse struct {
    Success bool                    `json:"success"`
}


// Check deposit reservation can be cancelled
func CanCancelDeposit(p *services.Provider) (*CanCancelDepositResponse, error) {

    // Response
    response := &CanCancelDepositResponse{}

    // Check node has current deposit reservation
    hasReservation := new(bool)
    if err := p.NodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
        return nil, errors.New("Error retrieving deposit reservation status: " + err.Error())
    } else {
        response.ReservationDidNotExist = !*hasReservation
    }

    // Update & return response
    response.Success = !response.ReservationDidNotExist
    return response, nil

}


// Cancel deposit reservation
func CancelDeposit(p *services.Provider) (*CancelDepositResponse, error) {

    // Cancel deposit reservation
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "depositReserveCancel"); err != nil {
            return nil, errors.New("Error canceling deposit reservation: " + err.Error())
        }
    }

    // Return response
    return &CancelDepositResponse{
        Success: true,
    }, nil

}

