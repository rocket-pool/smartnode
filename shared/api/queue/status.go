package queue

import (
    "errors"
    "math/big"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/settings"
)


// Get queue status response type
type GetQueueStatusResponse struct {
    Queues []*QueueStatus   `json:"queues"`
}
type QueueStatus struct {
    DurationId string       `json:"durationId"`
    Balance *big.Int        `json:"balance"`
    Chunks uint64           `json:"chunks"`
}


// Get queue statuses
func GetQueueStatus(p *services.Provider) (*GetQueueStatusResponse, error) {

    // Response
    response := &GetQueueStatusResponse{}

    // Get chunk size
    chunkSize := new(*big.Int)
    if err := p.CM.Contracts["rocketDepositSettings"].Call(nil, chunkSize, "getDepositChunkSize"); err != nil {
        return nil, errors.New("Error retrieving deposit chunk size: " + err.Error())
    }

    // Get minipool staking durations
    stakingDurations, err := settings.GetMinipoolStakingDurations(p.CM)
    if err != nil {
        return nil, errors.New("Error retrieving minipool staking durations: " + err.Error())
    }
    stakingDurationCount := len(stakingDurations)

    // Get queue balances
    balanceChannels := make([]chan *big.Int, stakingDurationCount)
    errorChannel := make(chan error)
    for di := 0; di < stakingDurationCount; di++ {
        balanceChannels[di] = make(chan *big.Int)
        go (func(di int) {
            balance := new(*big.Int)
            if err := p.CM.Contracts["rocketDepositQueue"].Call(nil, balance, "getBalance", stakingDurations[di].Id); err != nil {
                errorChannel <- errors.New("Error retrieving deposit queue balance: " + err.Error())
            } else {
                balanceChannels[di] <- *balance
            }
        })(di)
    }

    // Receive queue balances
    response.Queues = make([]*QueueStatus, stakingDurationCount)
    for di := 0; di < stakingDurationCount; di++ {
        select {
            case balance := <-balanceChannels[di]:
                chunks := big.NewInt(0)
                chunks.Div(balance, *chunkSize)
                response.Queues[di] = &QueueStatus{
                    DurationId: stakingDurations[di].Id,
                    Balance: balance,
                    Chunks: chunks.Uint64(),
                }
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return response
    return response, nil

}

