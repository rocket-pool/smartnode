package deposit

import (
    "errors"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Reserve deposit response type
type CanReserveDepositResponse struct {

    // Status
    Success bool                    `json:"success"`

    // Failure reasons
    HadExistingReservation bool     `json:"hadExistingReservation"`
    DepositsDisabled bool           `json:"depositsDisabled"`
    StakingDurationDisabled bool    `json:"stakingDurationDisabled"`

}
type ReserveDepositResponse struct {
    Success bool                    `json:"success"`
}


// Check node deposit can be reserved
func CanReserveDeposit(p *services.Provider, durationId string) (*CanReserveDepositResponse, error) {

    // Response
    response := &CanReserveDepositResponse{}

    // Status channels
    hasExistingReservationChannel := make(chan bool)
    depositsDisabledChannel := make(chan bool)
    stakingDurationDisabledChannel := make(chan bool)
    errorChannel := make(chan error)

    // Check node does not have current deposit reservation
    go (func() {
        hasReservation := new(bool)
        if err := p.NodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation status: " + err.Error())
        } else {
            hasExistingReservationChannel <- *hasReservation
        }
    })()

    // Check node deposits are enabled
    go (func() {
        depositsAllowed := new(bool)
        if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, depositsAllowed, "getDepositAllowed"); err != nil {
            errorChannel <- errors.New("Error checking node deposits enabled status: " + err.Error())
        } else {
            depositsDisabledChannel <- !*depositsAllowed
        }
    })()

    // Check staking duration is enabled
    go (func() {
        stakingDurationEnabled := new(bool)
        if err := p.CM.Contracts["rocketMinipoolSettings"].Call(nil, stakingDurationEnabled, "getMinipoolStakingDurationEnabled", durationId); err != nil {
            errorChannel <- errors.New("Error checking staking duration enabled status: " + err.Error())
        } else {
            stakingDurationDisabledChannel <- !*stakingDurationEnabled
        }
    })()

    // Receive status
    for received := 0; received < 3; {
        select {
            case response.HadExistingReservation = <-hasExistingReservationChannel:
                received++
            case response.DepositsDisabled = <-depositsDisabledChannel:
                received++
            case response.StakingDurationDisabled = <-stakingDurationDisabledChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Update & return response
    response.Success = !(response.HadExistingReservation || response.DepositsDisabled || response.StakingDurationDisabled)
    return response, nil

}


// Reserve node deposit
func ReserveDeposit(p *services.Provider, durationId string) (*ReserveDepositResponse, error) {

    // Create deposit reservation
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "depositReserve", durationId, validatorPubkey, signature, depositDataRoot); err != nil {
            return nil, errors.New("Error making deposit reservation: " + err.Error())
        }
    }

    // Return response
    return &ReserveDepositResponse{
        Success: true,
    }, nil

}

