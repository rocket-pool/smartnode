package node

import (
    "context"
    "errors"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
)


// Node balance data
type Balances struct {
    EtherWei *big.Int
    RplWei *big.Int
}


// Reservation detail data
type ReservationDetails struct {
    Exists bool
    StakingDurationID string
    EtherRequiredWei *big.Int
    RplRequiredWei *big.Int
    ExpiryTime time.Time
}


// Get a node account's balances
// Requires rocketPoolToken contract to be loaded with contract manager
func GetAccountBalances(nodeAccountAddress common.Address, client *ethclient.Client, cm *rocketpool.ContractManager) (*Balances, error) {

    // Account balances
    balances := &Balances{}

    // Balance data channels
    etherBalanceChannel := make(chan *big.Int)
    rplBalanceChannel := make(chan *big.Int)
    errorChannel := make(chan error)

    // Get node account ether balance
    go (func() {
        etherBalanceWei, err := client.BalanceAt(context.Background(), nodeAccountAddress, nil)
        if err != nil {
            errorChannel <- errors.New("Error retrieving node account ether balance: " + err.Error())
        } else {
            etherBalanceChannel <- etherBalanceWei
        }
    })()

    // Get node account RPL balance
    go (func() {
        rplBalanceWei := new(*big.Int)
        err := cm.Contracts["rocketPoolToken"].Call(nil, rplBalanceWei, "balanceOf", nodeAccountAddress)
        if err != nil {
            errorChannel <- errors.New("Error retrieving node account RPL balance: " + err.Error())
        } else {
            rplBalanceChannel <- *rplBalanceWei
        }
    })()

    // Receive balances
    for received := 0; received < 2; {
        select {
            case balances.EtherWei = <-etherBalanceChannel:
                received++
            case balances.RplWei = <-rplBalanceChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return balances, nil

}


// Get a node's balances
func GetBalances(nodeContract *bind.BoundContract) (*Balances, error) {

    // Node balances
    balances := &Balances{}

    // Balance data channels
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
    for received := 0; received < 2; {
        select {
            case balances.EtherWei = <-etherBalanceChannel:
                received++
            case balances.RplWei = <-rplBalanceChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return balances, nil

}


// Get a node's deposit reservation balance requirements
func GetRequiredBalances(nodeContract *bind.BoundContract) (*Balances, error) {

    // Balance requirement
    required := &Balances{}

    // Requirement data channels
    etherRequiredChannel := make(chan *big.Int)
    rplRequiredChannel := make(chan *big.Int)
    errorChannel := make(chan error)

    // Get deposit reservation ETH required
    go (func() {
        etherRequiredWei := new(*big.Int)
        err := nodeContract.Call(nil, etherRequiredWei, "getDepositReserveEtherRequired")
        if err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation ETH requirement: " + err.Error())
        } else {
            etherRequiredChannel <- *etherRequiredWei
        }
    })()

    // Get deposit reservation RPL required
    go (func() {
        rplRequiredWei := new(*big.Int)
        err := nodeContract.Call(nil, rplRequiredWei, "getDepositReserveRPLRequired")
        if err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation RPL requirement: " + err.Error())
        } else {
            rplRequiredChannel <- *rplRequiredWei
        }
    })()

    // Receive requirements
    for received := 0; received < 2; {
        select {
            case required.EtherWei = <-etherRequiredChannel:
                received++
            case required.RplWei = <-rplRequiredChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return required, nil

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
    requiredBalancesChannel := make(chan *Balances)
    reservedTimeChannel := make(chan *big.Int)
    reservationTimeChannel := make(chan *big.Int)
    errorChannel := make(chan error)

    // Get deposit reservation duration ID
    go (func() {
        durationID := new(string)
        err := nodeContract.Call(nil, durationID, "getDepositReserveDurationID")
        if err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation staking duration ID: " + err.Error())
        } else {
            durationIDChannel <- *durationID
        }
    })()

    // Get required balances
    go (func() {
        requiredBalances, err := GetRequiredBalances(nodeContract)
        if err != nil {
            errorChannel <- err
        } else {
            requiredBalancesChannel <- requiredBalances
        }
    })()

    // Get deposit reservation reserved time
    go (func() {
        reservedTime := new(*big.Int)
        err := nodeContract.Call(nil, reservedTime, "getDepositReservedTime")
        if err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation reserved time: " + err.Error())
        } else {
            reservedTimeChannel <- *reservedTime
        }
    })()

    // Get reservation duration
    go (func() {
        reservationTime := new(*big.Int)
        err := cm.Contracts["rocketNodeSettings"].Call(nil, reservationTime, "getDepositReservationTime")
        if err != nil {
            errorChannel <- errors.New("Error retrieving node deposit reservation time setting: " + err.Error())
        } else {
            reservationTimeChannel <- *reservationTime
        }
    })()

    // Receive reservation data
    var reservedTime *big.Int
    var reservationTime *big.Int
    for received := 0; received < 4; {
        select {
            case details.StakingDurationID = <-durationIDChannel:
                received++
            case requiredBalances := <-requiredBalancesChannel:
                received++
                details.EtherRequiredWei = requiredBalances.EtherWei
                details.RplRequiredWei = requiredBalances.RplWei
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

