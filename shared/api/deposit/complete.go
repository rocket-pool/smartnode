package deposit

import (
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// RocketPool PoolCreated event
type PoolCreated struct {
    Address common.Address
    DurationID [32]byte
    Created *big.Int
}


// Deposit completion response type
type DepositCompleteResponse struct {

    // Status
    Success bool                        `json:"success"`

    // Completion info
    TxValueWei *big.Int                 `json:"txValueWei"`
    MinipoolAddress common.Address      `json:"minipoolAddress"`

    // Failure info
    ReservationDidNotExist bool         `json:"reservationDidNotExist"`
    DepositsDisabled bool               `json:"depositsDisabled"`
    MinipoolCreationDisabled bool       `json:"minipoolCreationDisabled"`
    InsufficientNodeEtherBalance bool   `json:"insufficientNodeEtherBalance"`
    InsufficientNodeRplBalance bool     `json:"insufficientNodeRplBalance"`

}


// Check deposit reservation can be completed
func CanCompleteDeposit(p *services.Provider) (*DepositCompleteResponse, error) {

    // Response
    response := &DepositCompleteResponse{}

    // Status channels
    reservationNotExistsChannel := make(chan bool)
    depositsDisabledChannel := make(chan bool)
    minipoolCreationDisabledChannel := make(chan bool)
    errorChannel := make(chan error)

    // Check node has current deposit reservation
    go (func() {
        hasReservation := new(bool)
        if err := p.NodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation status: " + err.Error())
        } else {
            reservationNotExistsChannel <- !*hasReservation
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

    // Check minipool creation is enabled
    go (func() {
        minipoolCreationAllowed := new(bool)
        if err := p.CM.Contracts["rocketMinipoolSettings"].Call(nil, minipoolCreationAllowed, "getMinipoolCanBeCreated"); err != nil {
            errorChannel <- errors.New("Error checking minipool creation enabled status: " + err.Error())
        } else {
            minipoolCreationDisabledChannel <- !*minipoolCreationAllowed
        }
    })()

    // Receive status
    for received := 0; received < 3; {
        select {
            case response.ReservationDidNotExist = <-reservationNotExistsChannel:
                received++
            case response.DepositsDisabled = <-depositsDisabledChannel:
                received++
            case response.MinipoolCreationDisabled = <-minipoolCreationDisabledChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Check status
    if response.ReservationDidNotExist || response.DepositsDisabled || response.MinipoolCreationDisabled {
        return response, nil
    }

    // Get deposit reservation validator pubkey
    validatorPubkey := new([]byte)
    if err := p.NodeContract.Call(nil, validatorPubkey, "getDepositReserveValidatorPubkey"); err != nil {
        return nil, errors.New("Error retrieving deposit reservation validator pubkey: " + err.Error())
    }

    // Check for local validator key
    if _, err := p.KM.GetValidatorKey(*validatorPubkey); err != nil {
        return nil, errors.New("Local validator key matching deposit reservation validator pubkey not found")
    }

    // Data channels
    accountBalancesChannel := make(chan *node.Balances)
    nodeBalancesChannel := make(chan *node.Balances)
    requiredBalancesChannel := make(chan *node.Balances)
    depositDurationIDChannel := make(chan string)

    // Get node account balances
    go (func() {
        nodeAccount, _ := p.AM.GetNodeAccount()
        if accountBalances, err := node.GetAccountBalances(nodeAccount.Address, p.Client, p.CM); err != nil {
            errorChannel <- err
        } else {
            accountBalancesChannel <- accountBalances
        }
    })()

    // Get node balances
    go (func() {
        if nodeBalances, err := node.GetBalances(p.NodeContract); err != nil {
            errorChannel <- err
        } else {
            nodeBalancesChannel <- nodeBalances
        }
    })()

    // Get node balance requirements
    go (func() {
        if requiredBalances, err := node.GetRequiredBalances(p.NodeContract); err != nil {
            errorChannel <- err
        } else {
            requiredBalancesChannel <- requiredBalances
        }
    })()

    // Get deposit duration ID
    go (func() {
        durationID := new(string)
        if err := p.NodeContract.Call(nil, durationID, "getDepositReserveDurationID"); err != nil {
            errorChannel <- errors.New("Error retrieving deposit duration ID: " + err.Error())
        } else {
            depositDurationIDChannel <- *durationID
        }
    })()

    // Receive data
    var accountBalances *node.Balances
    var nodeBalances *node.Balances
    var requiredBalances *node.Balances
    var depositDurationID string
    for received := 0; received < 4; {
        select {
            case accountBalances = <-accountBalancesChannel:
                received++
            case nodeBalances = <-nodeBalancesChannel:
                received++
            case requiredBalances = <-requiredBalancesChannel:
                received++
            case depositDurationID = <-depositDurationIDChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Check node ether balance and get required deposit transaction value
    response.TxValueWei := new(big.Int)
    if nodeBalances.EtherWei.Cmp(requiredBalances.EtherWei) < 0 {
        response.TxValueWei.Sub(requiredBalances.EtherWei, nodeBalances.EtherWei)
        if accountBalances.EtherWei.Cmp(response.TxValueWei) < 0 {
            response.InsufficientNodeEtherBalance = true
        }
    }

    // Check node RPL balance
    if nodeBalances.RplWei.Cmp(requiredBalances.RplWei) < 0 {
        response.InsufficientNodeRplBalance = true
    }

    // Return response
    return response, nil

}


// Complete reserved deposit
func CompleteDeposit(p *services.Provider, txValueWei *big.Int) (*DepositCompleteResponse, error) {

    // Get account transactor
    txor, err := p.AM.GetNodeAccountTransactor()
    if err != nil { return nil, err }
    txor.Value = txValueWei

    // Complete deposit
    txReceipt, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "deposit")
    if err != nil {
        return nil, errors.New("Error completing deposit: " + err.Error())
    }

    // Get minipool created event
    minipoolCreatedEvents, err := eth.GetTransactionEvents(p.Client, txReceipt, p.CM.Addresses["rocketPool"], p.CM.Abis["rocketPool"], "PoolCreated", PoolCreated{})
    if err != nil {
        return nil, errors.New("Error retrieving deposit transaction minipool created event: " + err.Error())
    } else if len(minipoolCreatedEvents) == 0 {
        return nil, errors.New("Could not retrieve deposit transaction minipool created event")
    }
    minipoolCreatedEvent := (minipoolCreatedEvents[0]).(*PoolCreated)

    // Process deposit queue for duration
    if txor, err := p.AM.GetNodeAccountTransactor(); err == nil {
        _, _ = eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses["rocketDepositQueue"], p.CM.Abis["rocketDepositQueue"], "assignChunks", depositDurationID)
    }

    // Return response
    return &DepositCompleteResponse{
        Success: true,
        MinipoolAddress: minipoolCreatedEvent.Address,
    }, nil

}

