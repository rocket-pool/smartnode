package settings

import (
    "errors"
    "math/big"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


// Minipool staking duration details
type StakingDuration struct {
    Id string
    Epochs uint64
    Enabled bool
}


// Get all minipool staking durations
// Requires rocketMinipoolSettings contract to be loaded with contract manager
func GetMinipoolStakingDurations(cm *rocketpool.ContractManager) ([]*StakingDuration, error) {

    // Check contracts are loaded
    if _, ok := cm.Abis["rocketMinipoolSettings"]; !ok { return nil, errors.New("RocketMinipoolSettings contract is not loaded") }

    // Get staking duration count
    stakingDurationCountV := new(*big.Int)
    if err := cm.Contracts["rocketMinipoolSettings"].Call(nil, stakingDurationCountV, "getMinipoolStakingDurationCount"); err != nil {
        return nil, errors.New("Error retrieving staking duration count: " + err.Error())
    }
    stakingDurationCount := (*stakingDurationCountV).Int64()

    // Data channels
    idChannels := make([]chan string, stakingDurationCount)
    stakingDurationChannels := make([]chan *StakingDuration, stakingDurationCount)
    errorChannel := make(chan error)

    // Get staking duration IDs
    for di := int64(0); di < stakingDurationCount; di++ {
        idChannels[di] = make(chan string)
        go (func(di int64) {
            stakingDurationId := new(string)
            if err := cm.Contracts["rocketMinipoolSettings"].Call(nil, stakingDurationId, "getMinipoolStakingDurationAt", big.NewInt(di)); err != nil {
                errorChannel <- errors.New("Error retrieving staking duration ID: " + err.Error())
            } else {
                idChannels[di] <- *stakingDurationId
            }
        })(di)
    }

    // Receive staking duration IDs
    stakingDurationIds := make([]string, stakingDurationCount)
    for di := int64(0); di < stakingDurationCount; di++ {
        select {
            case stakingDurationId := <-idChannels[di]:
                stakingDurationIds[di] = stakingDurationId
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Get staking durations
    for di := int64(0); di < stakingDurationCount; di++ {
        stakingDurationChannels[di] = make(chan *StakingDuration)
        go (func(di int64) {
            if stakingDuration, err := GetMinipoolStakingDuration(cm, stakingDurationIds[di]); err != nil {
                errorChannel <- err
            } else {
                stakingDurationChannels[di] <- stakingDuration
            }
        })(di)
    }

    // Receive staking durations
    stakingDurations := make([]*StakingDuration, stakingDurationCount);
    for di := int64(0); di < stakingDurationCount; di++ {
        select {
            case stakingDuration := <-stakingDurationChannels[di]:
                stakingDurations[di] = stakingDuration
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return staking durations
    return stakingDurations, nil

}


// Get enabled minipool staking durations
// Requires rocketMinipoolSettings contract to be loaded with contract manager
func GetEnabledMinipoolStakingDurations(cm *rocketpool.ContractManager) ([]*StakingDuration, error) {

    // Get all staking durations
    stakingDurations, err := GetMinipoolStakingDurations(cm)
    if err != nil { return nil, err }

    // Filter staking durations and return
    filteredStakingDurations := []*StakingDuration{}
    for _, stakingDuration := range stakingDurations {
        if stakingDuration.Enabled {
            filteredStakingDurations = append(filteredStakingDurations, stakingDuration)
        }
    }
    return filteredStakingDurations, nil

}


// Get a minipool staking duration
// Requires rocketMinipoolSettings contract to be loaded with contract manager
func GetMinipoolStakingDuration(cm *rocketpool.ContractManager, id string) (*StakingDuration, error) {

    // Check contracts are loaded
    if _, ok := cm.Abis["rocketMinipoolSettings"]; !ok { return nil, errors.New("RocketMinipoolSettings contract is not loaded") }

    // Staking duration details
    details := &StakingDuration{
        Id: id,
    }

    // Data channels
    epochsChannel := make(chan uint64)
    enabledChannel := make(chan bool)
    errorChannel := make(chan error)

    // Get staking duration epochs
    go (func() {
        stakingDurationEpochs := new(*big.Int)
        if err := cm.Contracts["rocketMinipoolSettings"].Call(nil, stakingDurationEpochs, "getMinipoolStakingDurationEpochs", id); err != nil {
            errorChannel <- errors.New("Error retrieving staking duration epochs: " + err.Error())
        } else {
            epochsChannel <- (*stakingDurationEpochs).Uint64()
        }
    })()

    // Get staking duration enabled status
    go (func() {
        stakingDurationEnabled := new(bool)
        if err := cm.Contracts["rocketMinipoolSettings"].Call(nil, stakingDurationEnabled, "getMinipoolStakingDurationEnabled", id); err != nil {
            errorChannel <- errors.New("Error retrieving staking duration enabled status: " + err.Error())
        } else {
            enabledChannel <- *stakingDurationEnabled
        }
    })()

    // Receive staking duration details
    for received := 0; received < 2; {
        select {
            case details.Epochs = <-epochsChannel:
                received++
            case details.Enabled = <-enabledChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return details, nil

}

