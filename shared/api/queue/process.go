package queue

import (
    "errors"
    "math/big"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Process queue response type
type CanProcessQueueResponse struct {

    // Status
    Success bool                `json:"success"`

    // Failure reasons
    InvalidStakingDuration bool `json:"invalidStakingDuration"`
    InsufficientBalance bool    `json:"insufficientBalance"`
    NoAvailableNodes bool       `json:"noAvailableNodes"`

}
type ProcessQueueResponse struct {
    Success bool                `json:"success"`
}


// Check queue can be processes
func CanProcessQueue(p *services.Provider, durationId string) (*CanProcessQueueResponse, error) {

    // Response
    response := &CanProcessQueueResponse{}

    // Data channels
    invalidStakingDurationChannel := make(chan bool)
    noAvailableNodesChannel := make(chan bool)
    queueBalanceChannel := make(chan *big.Int)
    chunkSizeChannel := make(chan *big.Int)
    errorChannel := make(chan error)

    // Check staking duration is valid
    go (func() {
        stakingDurationValid := new(bool)
        if err := p.CM.Contracts["rocketMinipoolSettings"].Call(nil, stakingDurationValid, "getMinipoolStakingDurationExists", durationId); err != nil {
            errorChannel <- errors.New("Error checking staking duration validity: " + err.Error())
        } else {
            invalidStakingDurationChannel <- !*stakingDurationValid
        }
    })()

    // Check for available nodes
    go (func() {
        availableNodeCount := new(*big.Int)
        if err := p.CM.Contracts["rocketNode"].Call(nil, availableNodeCount, "getAvailableNodeCount", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving available node count: " + err.Error())
        } else {
            noAvailableNodesChannel <- ((*availableNodeCount).Cmp(big.NewInt(0)) == 0)
        }
    })()

    // Get queue balance
    go (func() {
        queueBalance := new(*big.Int)
        if err := p.CM.Contracts["rocketDepositQueue"].Call(nil, queueBalance, "getBalance", durationId); err != nil {
            errorChannel <- errors.New("Error retrieving deposit queue balance: " + err.Error())
        } else {
            queueBalanceChannel <- *queueBalance
        }
    })()

    // Get chunk size
    go (func() {
        chunkSize := new(*big.Int)
        if err := p.CM.Contracts["rocketDepositSettings"].Call(nil, chunkSize, "getDepositChunkSize"); err != nil {
            errorChannel <- errors.New("Error retrieving deposit chunk size: " + err.Error())
        } else {
            chunkSizeChannel <- *chunkSize
        }
    })()

    // Receive data
    var queueBalance *big.Int
    var chunkSize *big.Int
    for received := 0; received < 4; {
        select {
            case response.InvalidStakingDuration = <-invalidStakingDurationChannel:
                received++
            case response.NoAvailableNodes = <-noAvailableNodesChannel:
                received++
            case queueBalance = <-queueBalanceChannel:
                received++
            case chunkSize = <-chunkSizeChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Set response details
    response.InsufficientBalance = (queueBalance.Cmp(chunkSize) < 0)

    // Update & return response
    response.Success = !(response.InvalidStakingDuration || response.NoAvailableNodes || response.InsufficientBalance)
    return response, nil

}


// Process queue
func ProcessQueue(p *services.Provider, durationId string) (*ProcessQueueResponse, error) {

    // Process deposit queue
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses["rocketDepositQueue"], p.CM.Abis["rocketDepositQueue"], "assignChunks", durationId); err != nil {
            return nil, errors.New("Error processing deposit queue: " + err.Error())
        }
    }

    // Return response
    return &ProcessQueueResponse{
        Success: true,
    }, nil

}

