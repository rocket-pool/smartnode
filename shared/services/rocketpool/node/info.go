package node

import (
    "bytes"
    "context"
    "errors"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Node balance data
type Balances struct {
    EtherWei *big.Int
    RethWei *big.Int
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
// Requires rocketETHToken & rocketPoolToken contracts to be loaded with contract manager
func GetAccountBalances(nodeAccountAddress common.Address, client *ethclient.Client, cm *rocketpool.ContractManager) (*Balances, error) {

    // Check contracts are loaded
    if _, ok := cm.Contracts["rocketETHToken"]; !ok { return nil, errors.New("RocketETHToken contract is not loaded") }
    if _, ok := cm.Contracts["rocketPoolToken"]; !ok { return nil, errors.New("RocketPoolToken contract is not loaded") }

    // Account balances
    balances := &Balances{}

    // Balance data channels
    etherBalanceChannel := make(chan *big.Int)
    rethBalanceChannel := make(chan *big.Int)
    rplBalanceChannel := make(chan *big.Int)
    errorChannel := make(chan error)

    // Get node account ether balance
    go (func() {
        if etherBalanceWei, err := client.BalanceAt(context.Background(), nodeAccountAddress, nil); err != nil {
            errorChannel <- errors.New("Error retrieving node account ether balance: " + err.Error())
        } else {
            etherBalanceChannel <- etherBalanceWei
        }
    })()

    // Get node account rETH balance
    go (func() {
        rethBalanceWei := new(*big.Int)
        if err := cm.Contracts["rocketETHToken"].Call(nil, rethBalanceWei, "balanceOf", nodeAccountAddress); err != nil {
            errorChannel <- errors.New("Error retrieving node account rETH balance: " + err.Error())
        } else {
            rethBalanceChannel <- *rethBalanceWei
        }
    })()

    // Get node account RPL balance
    go (func() {
        rplBalanceWei := new(*big.Int)
        if err := cm.Contracts["rocketPoolToken"].Call(nil, rplBalanceWei, "balanceOf", nodeAccountAddress); err != nil {
            errorChannel <- errors.New("Error retrieving node account RPL balance: " + err.Error())
        } else {
            rplBalanceChannel <- *rplBalanceWei
        }
    })()

    // Receive balances
    for received := 0; received < 3; {
        select {
            case balances.EtherWei = <-etherBalanceChannel:
                received++
            case balances.RethWei = <-rethBalanceChannel:
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


// Get a node contract's balances
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
        if err := nodeContract.Call(nil, etherBalanceWei, "getBalanceETH"); err != nil {
            errorChannel <- errors.New("Error retrieving node ETH balance: " + err.Error())
        } else {
            etherBalanceChannel <- *etherBalanceWei
        }
    })()

    // Get node RPL balance
    go (func() {
        rplBalanceWei := new(*big.Int)
        if err := nodeContract.Call(nil, rplBalanceWei, "getBalanceRPL"); err != nil {
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
        if err := nodeContract.Call(nil, etherRequiredWei, "getDepositReserveEtherRequired"); err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation ETH requirement: " + err.Error())
        } else {
            etherRequiredChannel <- *etherRequiredWei
        }
    })()

    // Get deposit reservation RPL required
    go (func() {
        rplRequiredWei := new(*big.Int)
        if err := nodeContract.Call(nil, rplRequiredWei, "getDepositReserveRPLRequired"); err != nil {
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

    // Check rocketNodeSettings contract is loaded
    if _, ok := cm.Contracts["rocketNodeSettings"]; !ok { return nil, errors.New("RocketNodeSettings contract is not loaded") }

    // Reservation details
    details := &ReservationDetails{}

    // Check if node has current deposit reservation
    hasReservation := new(bool)
    if err := nodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
        return nil, errors.New("Error retrieving deposit reservation status: " + err.Error())
    }
    if details.Exists = *hasReservation; !details.Exists {
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
        if err := nodeContract.Call(nil, durationID, "getDepositReserveDurationID"); err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation staking duration ID: " + err.Error())
        } else {
            durationIDChannel <- *durationID
        }
    })()

    // Get required balances
    go (func() {
        if requiredBalances, err := GetRequiredBalances(nodeContract); err != nil {
            errorChannel <- err
        } else {
            requiredBalancesChannel <- requiredBalances
        }
    })()

    // Get deposit reservation reserved time
    go (func() {
        reservedTime := new(*big.Int)
        if err := nodeContract.Call(nil, reservedTime, "getDepositReservedTime"); err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation reserved time: " + err.Error())
        } else {
            reservedTimeChannel <- *reservedTime
        }
    })()

    // Get reservation duration
    go (func() {
        reservationTime := new(*big.Int)
        if err := cm.Contracts["rocketNodeSettings"].Call(nil, reservationTime, "getDepositReservationTime"); err != nil {
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


// Get a list of a node's minipool addresses
// Requires utilAddressSetStorage contract to be loaded with contract manager
func GetMinipoolAddresses(nodeAccountAddress common.Address, cm *rocketpool.ContractManager) ([]*common.Address, error) {

    // Check utilAddressSetStorage contract is loaded
    if _, ok := cm.Contracts["utilAddressSetStorage"]; !ok { return nil, errors.New("UtilAddressSetStorage contract is not loaded") }

    // Get node minipool list key
    minipoolListKey := eth.KeccakBytes(bytes.Join([][]byte{[]byte("minipools"), []byte("list.node"), nodeAccountAddress.Bytes()}, []byte{}))

    // Get node minipool count
    minipoolCountV := new(*big.Int)
    if err := cm.Contracts["utilAddressSetStorage"].Call(nil, minipoolCountV, "getCount", minipoolListKey); err != nil {
        return nil, errors.New("Error retrieving node minipool count: " + err.Error())
    }
    minipoolCount := (*minipoolCountV).Int64()

    // Get minipool addresses
    addressChannels := make([]chan *common.Address, minipoolCount)
    errorChannel := make(chan error)
    for mi := int64(0); mi < minipoolCount; mi++ {
        addressChannels[mi] = make(chan *common.Address)
        go (func(mi int64) {
            minipoolAddress := new(common.Address)
            if err := cm.Contracts["utilAddressSetStorage"].Call(nil, minipoolAddress, "getItem", minipoolListKey, big.NewInt(mi)); err != nil {
                errorChannel <- errors.New("Error retrieving node minipool address: " + err.Error())
            } else {
                addressChannels[mi] <- minipoolAddress
            }
        })(mi)
    }

    // Receive minipool addresses
    minipoolAddresses := make([]*common.Address, minipoolCount)
    for mi := int64(0); mi < minipoolCount; mi++ {
        select {
            case address := <-addressChannels[mi]:
                minipoolAddresses[mi] = address
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return minipoolAddresses, nil

}

