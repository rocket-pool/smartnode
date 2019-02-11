package node

import (
    "errors"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
)


// Reservation detail data
type ReservationDetails struct {
    Exists bool
    StakingDurationID string
    EtherRequiredWei *big.Int
    RplRequiredWei *big.Int
    ExpiryTime time.Time
}


// Get a node's balances
func GetBalances(nodeContract *bind.BoundContract) (*big.Int, *big.Int, error) {

    // Balance channels
    etherBalanceChannel := make(chan *big.Int)
    rplBalanceChannel := make(chan *big.Int)
    errorChannel := make(chan error)

    // Get node ETH balance
    go (func() {
        etherBalanceWei := new(*big.Int)
        err := nodeContract.Call(nil, etherBalanceWei, "getBalanceETH")
        if err != nil {
            errorChannel <- errors.New("Error retrieving node ETH balance: " + err.Error())
        } else {
            etherBalanceChannel <- *etherBalanceWei
        }
    })()

    // Get node RPL balance
    go (func() {
        rplBalanceWei := new(*big.Int)
        err := nodeContract.Call(nil, rplBalanceWei, "getBalanceRPL")
        if err != nil {
            errorChannel <- errors.New("Error retrieving node RPL balance: " + err.Error())
        } else {
            rplBalanceChannel <- *rplBalanceWei
        }
    })()

    // Receive balances
    var etherBalanceWei *big.Int
    var rplBalanceWei *big.Int
    for received := 0; received < 2; {
        select {
            case etherBalanceWei = <-etherBalanceChannel:
                received++
            case rplBalanceWei = <-rplBalanceChannel:
                received++
            case err := <-errorChannel:
                return nil, nil, err
        }
    }

    // Return
    return etherBalanceWei, rplBalanceWei, nil

}


// Get a node's deposit reservation details
// Requires rocketNodeSettings contract to be loaded with contract manager
func GetReservationDetails(nodeContract *bind.BoundContract, cm *rocketpool.ContractManager) (*ReservationDetails, error) {

    // Reservation details
    details := &ReservationDetails{}

    // Check if node has current deposit reservation
    hasReservation := new(bool)
    err := nodeContract.Call(nil, hasReservation, "getHasDepositReservation")
    if err != nil {
        return nil, errors.New("Error retrieving deposit reservation status: " + err.Error())
    }
    details.Exists = *hasReservation
    if !details.Exists {
        return details, nil
    }

    // Reservation data channels
    durationIDChannel := make(chan string)
    etherRequiredChannel := make(chan *big.Int)
    rplRequiredChannel := make(chan *big.Int)
    reservedTimeChannel := make(chan *big.Int)
    reservationTimeChannel := make(chan *big.Int)
    errorChannel := make(chan error)

    // Get deposit reservation duration ID
    go (func() {
        durationID := new(string)
        err = nodeContract.Call(nil, durationID, "getDepositReserveDurationID")
        if err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation staking duration ID: " + err.Error())
        } else {
            durationIDChannel <- *durationID
        }
    })()

    // Get deposit reservation ETH required
    go (func() {
        etherRequiredWei := new(*big.Int)
        err = nodeContract.Call(nil, etherRequiredWei, "getDepositReserveEtherRequired")
        if err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation ETH requirement: " + err.Error())
        } else {
            etherRequiredChannel <- *etherRequiredWei
        }
    })()

    // Get deposit reservation RPL required
    go (func() {
        rplRequiredWei := new(*big.Int)
        err = nodeContract.Call(nil, rplRequiredWei, "getDepositReserveRPLRequired")
        if err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation RPL requirement: " + err.Error())
        } else {
            rplRequiredChannel <- *rplRequiredWei
        }
    })()

    // Get deposit reservation reserved time
    go (func() {
        reservedTime := new(*big.Int)
        err = nodeContract.Call(nil, reservedTime, "getDepositReservedTime")
        if err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation reserved time: " + err.Error())
        } else {
            reservedTimeChannel <- *reservedTime
        }
    })()

    // Get reservation duration
    go (func() {
        reservationTime := new(*big.Int)
        err = cm.Contracts["rocketNodeSettings"].Call(nil, reservationTime, "getDepositReservationTime")
        if err != nil {
            errorChannel <- errors.New("Error retrieving node deposit reservation time setting: " + err.Error())
        } else {
            reservationTimeChannel <- *reservationTime
        }
    })()

    // Receive reservation data
    var reservedTime *big.Int
    var reservationTime *big.Int
    for received := 0; received < 5; {
        select {
            case details.StakingDurationID = <-durationIDChannel:
                received++
            case details.EtherRequiredWei = <-etherRequiredChannel:
                received++
            case details.RplRequiredWei = <-rplRequiredChannel:
                received++
            case reservedTime = <-reservedTimeChannel:
                received++
            case reservationTime = <-reservationTimeChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Get deposit reservation expiry time
    var expiryTimestamp big.Int
    expiryTimestamp.Add(reservedTime, reservationTime)
    details.ExpiryTime = time.Unix(expiryTimestamp.Int64(), 0)

    // Return
    return details, nil

}

