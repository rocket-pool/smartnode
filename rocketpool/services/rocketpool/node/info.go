package node

import (
    "errors"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Reservation detail data
type ReservationDetails struct {
    Exists bool
    StakingDurationID string
    EtherRequired big.Int
    RplRequired big.Int
    ExpiryTime time.Time
}


// Get a node's balances
func GetBalances(nodeContract *bind.BoundContract) (*big.Int, *big.Int, error) {

    // Get node ETH balance
    etherBalanceWei := new(*big.Int)
    err := nodeContract.Call(nil, etherBalanceWei, "getBalanceETH")
    if err != nil {
        return nil, nil, errors.New("Error retrieving node ETH balance: " + err.Error())
    }
    etherBalance := eth.WeiToEth(*etherBalanceWei)

    // Get node RPL balance
    rplBalanceWei := new(*big.Int)
    err = nodeContract.Call(nil, rplBalanceWei, "getBalanceRPL")
    if err != nil {
        return nil, nil, errors.New("Error retrieving node RPL balance: " + err.Error())
    }
    rplBalance := eth.WeiToEth(*rplBalanceWei)

    // Return
    return &etherBalance, &rplBalance, nil

}


// Get a node's deposit reservation details
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

    // Get deposit reservation duration ID
    durationID := new(string)
    err = nodeContract.Call(nil, durationID, "getDepositReserveDurationID")
    if err != nil {
        return nil, errors.New("Error retrieving deposit reservation staking duration ID: " + err.Error())
    }
    details.StakingDurationID = *durationID

    // Get deposit reservation ETH required
    etherRequiredWei := new(*big.Int)
    err = nodeContract.Call(nil, etherRequiredWei, "getDepositReserveEtherRequired")
    if err != nil {
        return nil, errors.New("Error retrieving deposit reservation ETH requirement: " + err.Error())
    }
    details.EtherRequired = eth.WeiToEth(*etherRequiredWei)

    // Get deposit reservation RPL required
    rplRequiredWei := new(*big.Int)
    err = nodeContract.Call(nil, rplRequiredWei, "getDepositReserveRPLRequired")
    if err != nil {
        return nil, errors.New("Error retrieving deposit reservation RPL requirement: " + err.Error())
    }
    details.RplRequired = eth.WeiToEth(*rplRequiredWei)

    // Get deposit reservation reserved time
    reservedTime := new(*big.Int)
    err = nodeContract.Call(nil, reservedTime, "getDepositReservedTime")
    if err != nil {
        return nil, errors.New("Error retrieving deposit reservation reserved time: " + err.Error())
    }

    // Get reservation duration
    reservationTime := new(*big.Int)
    err = cm.Contracts["rocketNodeSettings"].Call(nil, reservationTime, "getDepositReservationTime")
    if err != nil {
        return nil, errors.New("Error retrieving node deposit reservation time setting: " + err.Error())
    }

    // Get deposit reservation expiry time
    var expiryTimestamp big.Int
    expiryTimestamp.Add(*reservedTime, *reservationTime)
    details.ExpiryTime = time.Unix(expiryTimestamp.Int64(), 0)

    // Return
    return details, nil

}

