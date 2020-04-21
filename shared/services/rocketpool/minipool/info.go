package minipool

import (
    "encoding/hex"
    "errors"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


// Minipoool status types
const (
    INITIALIZED = 0
    DEPOSIT_ASSIGNED = 1
    PRELAUNCH = 2
    STAKING = 3
    LOGGED_OUT = 4
    WITHDRAWN = 5
    CLOSED = 6
    TIMED_OUT = 7
)


// Minipool detail data
type Details struct {
    Address *common.Address         `json:"address"`
    Status uint8                    `json:"status"`
    StatusType string               `json:"statusType"`
    StatusTime time.Time            `json:"statusTime"`
    StatusBlock *big.Int            `json:"statusBlock"`
    StakingDurationId string        `json:"stakingDurationId"`
    StakingDuration *big.Int        `json:"stakingDuration"`
    NodeDepositExists bool          `json:"nodeDepositExists"`
    NodeEtherBalanceWei *big.Int    `json:"nodeEtherBalanceWei"`
    NodeRplBalanceWei *big.Int      `json:"nodeRplBalanceWei"`
    UserDepositCount *big.Int       `json:"userDepositCount"`
    UserDepositCapacityWei *big.Int `json:"userDepositCapacityWei"`
    UserDepositTotalWei *big.Int    `json:"userDepositTotalWei"`
}


// Minipool status data
type Status struct {
    Status uint8
    StatusBlock *big.Int
    StakingDuration *big.Int
    ValidatorPubkey []byte
}


// Minipool node status data
type NodeStatus struct {
    Address *common.Address
    Status uint8
    StatusType string
    StatusTime time.Time
    StakingDurationId string
    DepositExists bool
}


// Get a minipool's details
// Requires rocketMinipool ABI and rocketPoolToken contract to be loaded with contract manager
func GetDetails(cm *rocketpool.ContractManager, minipoolAddress *common.Address) (*Details, error) {

    // Check contracts & ABIs are loaded
    if _, ok := cm.Abis["rocketMinipool"]; !ok { return nil, errors.New("RocketMinipool ABI is not loaded") }
    if _, ok := cm.Contracts["rocketPoolToken"]; !ok { return nil, errors.New("RocketPoolToken contract is not loaded") }

    // Minipool details
    details := &Details{
        Address: minipoolAddress,
    }

    // Initialise minipool contract
    minipoolContract, err := cm.NewContract(minipoolAddress, "rocketMinipool")
    if err != nil {
        return nil, errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Data channels
    statusChannel := make(chan uint8)
    statusTimeChannel := make(chan time.Time)
    statusBlockChannel := make(chan *big.Int)
    stakingDurationIdChannel := make(chan string)
    stakingDurationChannel := make(chan *big.Int)
    nodeDepositExistsChannel := make(chan bool)
    nodeEtherBalanceChannel := make(chan *big.Int)
    nodeRplBalanceChannel := make(chan *big.Int)
    userDepositCountChannel := make(chan *big.Int)
    userDepositCapacityChannel := make(chan *big.Int)
    userDepositTotalChannel := make(chan *big.Int)
    errorChannel := make(chan error)

    // Get status
    go (func() {
        status := new(uint8)
        if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status: " + err.Error())
        } else {
            statusChannel <- *status
        }
    })()

    // Get status time
    go (func() {
        statusChangedTime := new(*big.Int)
        if err := minipoolContract.Call(nil, statusChangedTime, "getStatusChangedTime"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status changed time: " + err.Error())
        } else {
            statusTimeChannel <- time.Unix((*statusChangedTime).Int64(), 0)
        }
    })()

    // Get status block
    go (func() {
        statusBlock := new(*big.Int)
        if err := minipoolContract.Call(nil, statusBlock, "getStatusChangedBlock"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status changed block: " + err.Error())
        } else {
            statusBlockChannel <- *statusBlock
        }
    })()

    // Get staking duration ID
    go (func() {
        stakingDurationId := new(string)
        if err := minipoolContract.Call(nil, stakingDurationId, "getStakingDurationID"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool staking duration ID: " + err.Error())
        } else {
            stakingDurationIdChannel <- *stakingDurationId
        }
    })()

    // Get staking duration
    go (func() {
        stakingDuration := new(*big.Int)
        if err := minipoolContract.Call(nil, stakingDuration, "getStakingDuration"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool staking duration: " + err.Error())
        } else {
            stakingDurationChannel <- *stakingDuration
        }
    })()

    // Get node deposit exists flag
    go (func() {
        nodeDepositExists := new(bool)
        if err := minipoolContract.Call(nil, nodeDepositExists, "getNodeDepositExists"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool node deposit status: " + err.Error())
        } else {
            nodeDepositExistsChannel <- *nodeDepositExists
        }
    })()

    // Get node ETH balance
    go (func() {
        nodeEtherBalanceWei := new(*big.Int)
        if err := minipoolContract.Call(nil, nodeEtherBalanceWei, "getNodeBalance"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool node ETH balance: " + err.Error())
        } else {
            nodeEtherBalanceChannel <- *nodeEtherBalanceWei
        }
    })()

    // Get node RPL balance
    go (func() {
        nodeRplBalanceWei := new(*big.Int)
        if err := cm.Contracts["rocketPoolToken"].Call(nil, nodeRplBalanceWei, "balanceOf", minipoolAddress); err != nil {
            errorChannel <- errors.New("Error retrieving minipool node RPL balance: " + err.Error())
        } else {
            nodeRplBalanceChannel <- *nodeRplBalanceWei
        }
    })()

    // Get deposit count
    go (func() {
        userDepositCount := new(*big.Int)
        if err := minipoolContract.Call(nil, userDepositCount, "getDepositCount"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool deposit count: " + err.Error())
        } else {
            userDepositCountChannel <- *userDepositCount
        }
    })()

    // Get user deposit capacity
    go (func() {
        userDepositCapacityWei := new(*big.Int)
        if err := minipoolContract.Call(nil, userDepositCapacityWei, "getUserDepositCapacity"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool user deposit capacity: " + err.Error())
        } else {
            userDepositCapacityChannel <- *userDepositCapacityWei
        }
    })()

    // Get user deposit total
    go (func() {
        userDepositTotalWei := new(*big.Int)
        if err := minipoolContract.Call(nil, userDepositTotalWei, "getUserDepositTotal"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool user deposit total: " + err.Error())
        } else {
            userDepositTotalChannel <- *userDepositTotalWei
        }
    })()

    // Receive minipool data
    for received := 0; received < 11; {
        select {
            case details.Status = <-statusChannel:
                details.StatusType = GetStatusType(details.Status)
                received++
            case details.StatusTime = <-statusTimeChannel:
                received++
            case details.StatusBlock = <-statusBlockChannel:
                received++
            case details.StakingDurationId = <-stakingDurationIdChannel:
                received++
            case details.StakingDuration = <-stakingDurationChannel:
                received++
            case details.NodeDepositExists = <-nodeDepositExistsChannel:
                received++
            case details.NodeEtherBalanceWei = <-nodeEtherBalanceChannel:
                received++
            case details.NodeRplBalanceWei = <-nodeRplBalanceChannel:
                received++
            case details.UserDepositCount = <-userDepositCountChannel:
                received++
            case details.UserDepositCapacityWei = <-userDepositCapacityChannel:
                received++
            case details.UserDepositTotalWei = <-userDepositTotalChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return details, nil

}


// Get a minipool's status details
// Requires rocketMinipool ABI to be loaded with contract manager
func GetStatus(cm *rocketpool.ContractManager, minipoolAddress *common.Address) (*Status, error) {

    // Check rocketMinipool ABI is loaded
    if _, ok := cm.Abis["rocketMinipool"]; !ok { return nil, errors.New("RocketMinipool ABI is not loaded") }

    // Minipool status
    status := &Status{}

    // Initialise minipool contract
    minipoolContract, err := cm.NewContract(minipoolAddress, "rocketMinipool")
    if err != nil {
        return nil, errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Data channels
    statusChannel := make(chan uint8)
    statusBlockChannel := make(chan *big.Int)
    stakingDurationChannel := make(chan *big.Int)
    validatorPubkeyChannel := make(chan []byte)
    errorChannel := make(chan error)

    // Get status
    go (func() {
        status := new(uint8)
        if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status: " + err.Error())
        } else {
            statusChannel <- *status
        }
    })()

    // Get status block
    go (func() {
        statusBlock := new(*big.Int)
        if err := minipoolContract.Call(nil, statusBlock, "getStatusChangedBlock"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status changed block: " + err.Error())
        } else {
            statusBlockChannel <- *statusBlock
        }
    })()

    // Get staking duration
    go (func() {
        stakingDuration := new(*big.Int)
        if err := minipoolContract.Call(nil, stakingDuration, "getStakingDuration"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool staking duration: " + err.Error())
        } else {
            stakingDurationChannel <- *stakingDuration
        }
    })()

    // Get validator pubkey
    go (func() {
        validatorPubkey := new([]byte)
        if err := minipoolContract.Call(nil, validatorPubkey, "getValidatorPubkey"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool validator pubkey: " + err.Error())
        } else {
            validatorPubkeyChannel <- *validatorPubkey
        }
    })()

    // Receive minipool data
    for received := 0; received < 4; {
        select {
            case status.Status = <-statusChannel:
                received++
            case status.StatusBlock = <-statusBlockChannel:
                received++
            case status.StakingDuration = <-stakingDurationChannel:
                received++
            case status.ValidatorPubkey = <-validatorPubkeyChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return status, nil

}


// Get a minipool's node status
// Requires rocketMinipool ABI to be loaded with contract manager
func GetNodeStatus(cm *rocketpool.ContractManager, minipoolAddress *common.Address) (*NodeStatus, error) {

    // Check rocketMinipool ABI is loaded
    if _, ok := cm.Abis["rocketMinipool"]; !ok { return nil, errors.New("RocketMinipool ABI is not loaded") }

    // Node status
    nodeStatus := &NodeStatus{
        Address: minipoolAddress,
    }

    // Initialise minipool contract
    minipoolContract, err := cm.NewContract(minipoolAddress, "rocketMinipool")
    if err != nil {
        return nil, errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Data channels
    statusChannel := make(chan uint8)
    statusTimeChannel := make(chan time.Time)
    stakingDurationIdChannel := make(chan string)
    depositExistsChannel := make(chan bool)
    errorChannel := make(chan error)

    // Get status
    go (func() {
        status := new(uint8)
        if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status: " + err.Error())
        } else {
            statusChannel <- *status
        }
    })()

    // Get status time
    go (func() {
        statusChangedTime := new(*big.Int)
        if err := minipoolContract.Call(nil, statusChangedTime, "getStatusChangedTime"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status changed time: " + err.Error())
        } else {
            statusTimeChannel <- time.Unix((*statusChangedTime).Int64(), 0)
        }
    })()

    // Get staking duration ID
    go (func() {
        stakingDurationId := new(string)
        if err := minipoolContract.Call(nil, stakingDurationId, "getStakingDurationID"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool staking duration ID: " + err.Error())
        } else {
            stakingDurationIdChannel <- *stakingDurationId
        }
    })()

    // Get node deposit status
    go (func() {
        nodeDepositExists := new(bool)
        if err := minipoolContract.Call(nil, nodeDepositExists, "getNodeDepositExists"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool node deposit status: " + err.Error())
        } else {
            depositExistsChannel <- *nodeDepositExists
        }
    })()

    // Receive minipool data
    for received := 0; received < 4; {
        select {
            case nodeStatus.Status = <-statusChannel:
                nodeStatus.StatusType = GetStatusType(nodeStatus.Status)
                received++
            case nodeStatus.StatusTime = <-statusTimeChannel:
                received++
            case nodeStatus.StakingDurationId = <-stakingDurationIdChannel:
                received++
            case nodeStatus.DepositExists = <-depositExistsChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return nodeStatus, nil

}


// Get a minipool's status code
// Requires rocketMinipool ABI to be loaded with contract manager
func GetStatusCode(cm *rocketpool.ContractManager, minipoolAddress *common.Address) (uint8, error) {

    // Check rocketMinipool ABI is loaded
    if _, ok := cm.Abis["rocketMinipool"]; !ok { return 0, errors.New("RocketMinipool ABI is not loaded") }

    // Initialise minipool contract
    minipoolContract, err := cm.NewContract(minipoolAddress, "rocketMinipool")
    if err != nil {
        return 0, errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Get minipool status
    status := new(uint8)
    if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
        return 0, errors.New("Error retrieving minipool status: " + err.Error())
    }

    // Return
    return *status, nil

}


// Get a map of all active minipools by validator pubkey
// Requires rocketPool contract and rocketMinipool ABI to be loaded with contract manager
func GetActiveMinipoolsByValidatorPubkey(cm *rocketpool.ContractManager) (*map[string]common.Address, error) {

    // Check contracts & ABIs are loaded
    if _, ok := cm.Contracts["rocketPool"]; !ok { return nil, errors.New("RocketPool contract is not loaded") }
    if _, ok := cm.Abis["rocketMinipool"]; !ok { return nil, errors.New("RocketMinipool ABI is not loaded") }

    // Get minipool count
    minipoolCountV := new(*big.Int)
    if err := cm.Contracts["rocketPool"].Call(nil, minipoolCountV, "getPoolsCount"); err != nil {
        return nil, errors.New("Error retrieving minipool count: " + err.Error())
    }
    minipoolCount := (*minipoolCountV).Int64()

    // Data channels
    addressChannels := make([]chan *common.Address, minipoolCount)
    statusChannels := make([]chan uint8, minipoolCount)
    validatorPubkeyChannels := make([]chan string, minipoolCount)
    errorChannel := make(chan error)

    // Get minipool addresses
    for mi := int64(0); mi < minipoolCount; mi++ {
        addressChannels[mi] = make(chan *common.Address)
        go (func(mi int64) {
            minipoolAddress := new(common.Address)
            if err := cm.Contracts["rocketPool"].Call(nil, minipoolAddress, "getPoolAt", big.NewInt(mi)); err != nil {
                errorChannel <- errors.New("Error retrieving minipool address: " + err.Error())
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

    // Get minipool statuses & validator pubkeys
    for mi := int64(0); mi < minipoolCount; mi++ {
        statusChannels[mi] = make(chan uint8)
        validatorPubkeyChannels[mi] = make(chan string)
        go (func(mi int64) {

            // Initialise minipool contract
            minipoolContract, err := cm.NewContract(minipoolAddresses[mi], "rocketMinipool")
            if err != nil {
                errorChannel <- errors.New("Error initialising minipool contract: " + err.Error())
                return
            }

            // Get status
            status := new(uint8)
            if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
                errorChannel <- errors.New("Error retrieving minipool status: " + err.Error())
            } else {
                statusChannels[mi] <- *status
            }

            // Get validator pubkey
            validatorPubkey := new([]byte)
            if err := minipoolContract.Call(nil, validatorPubkey, "getValidatorPubkey"); err != nil {
                errorChannel <- errors.New("Error retrieving minipool validator pubkey: " + err.Error())
            } else {
                validatorPubkeyChannels[mi] <- hex.EncodeToString(*validatorPubkey)
            }

        })(mi)
    }

    // Receive minipool statuses & validator pubkeys & build map
    activeMinipools := make(map[string]common.Address)
    for mi := int64(0); mi < minipoolCount; mi++ {

        // Receive data
        var status uint8
        var validatorPubkey string
        for received := 0; received < 2; {
            select {
                case status = <- statusChannels[mi]:
                    received++
                case validatorPubkey = <-validatorPubkeyChannels[mi]:
                    received++
                case err := <-errorChannel:
                    return nil, err
            }
        }

        // Filter by status & add to map
        if status >= STAKING && status <= LOGGED_OUT {
            activeMinipools[validatorPubkey] = *minipoolAddresses[mi];
        }

    }

    // Return
    return &activeMinipools, nil

}


// Get the status type by value
func GetStatusType(value uint8) string {
    switch value {
        case INITIALIZED: return "initialized"
        case DEPOSIT_ASSIGNED: return "depositassigned"
        case PRELAUNCH: return "prelaunch"
        case STAKING: return "staking"
        case LOGGED_OUT: return "loggedout"
        case WITHDRAWN: return "withdrawn"
        case CLOSED: return "closed"
        case TIMED_OUT: return "timedout"
        default: return "unknown"
    }
}

