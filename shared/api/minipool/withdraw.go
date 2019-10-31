package minipool

import (
    "bytes"
    "context"
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// RocketMinipool NodeWithdrawal event
type NodeWithdrawal struct {
    To common.Address
    EtherAmount *big.Int
    RethAmount *big.Int
    RplAmount *big.Int
    Created *big.Int
}


// Withdraw minipool response types
type CanWithdrawMinipoolsResponse struct {

    // Status
    Success bool                `json:"success"`

    // Failure reasons
    WithdrawalsDisabled bool    `json:"withdrawalsDisabled"`

}
type CanWithdrawMinipoolResponse struct {

    // Status
    Success bool                `json:"success"`

    // Failure reasons
    MinipoolDidNotExist bool    `json:"minipoolDidNotExist"`
    WithdrawalsDisabled bool    `json:"withdrawalsDisabled"`
    InvalidNodeOwner bool       `json:"invalidNodeOwner"`
    InvalidStatus bool          `json:"invalidStatus"`
    NodeDepositDidNotExist bool `json:"nodeDepositDidNotExist"`

    // Failure info
    NodeOwner common.Address    `json:"nodeOwner"`
    Status uint8                `json:"status"`

}
type WithdrawMinipoolResponse struct {

    // Status
    Success bool                `json:"success"`

    // Withdrawal info
    EtherWithdrawnWei *big.Int  `json:"etherWithdrawnWei"`
    RethWithdrawnWei *big.Int   `json:"rethWithdrawnWei"`
    RplWithdrawnWei *big.Int    `json:"rplWithdrawnWei"`

}


// Get withdrawable minipools
func GetWithdrawableMinipools(p *services.Provider) ([]*minipool.NodeStatus, error) {

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()

    // Get minipool addresses
    minipoolAddresses, err := node.GetMinipoolAddresses(nodeAccount.Address, p.CM)
    if err != nil { return nil, err }
    minipoolCount := len(minipoolAddresses)

    // Get minipool node statuses
    nodeStatusChannel := make([]chan *minipool.NodeStatus, minipoolCount)
    nodeStatusErrorChannel := make(chan error)
    for mi := 0; mi < minipoolCount; mi++ {
        nodeStatusChannel[mi] = make(chan *minipool.NodeStatus)
        go (func(mi int) {
            if nodeStatus, err := minipool.GetNodeStatus(p.CM, minipoolAddresses[mi]); err != nil {
                nodeStatusErrorChannel <- err
            } else {
                nodeStatusChannel[mi] <- nodeStatus
            }
        })(mi)
    }

    // Receive minipool node statuses & filter withdrawable minipools
    withdrawableMinipools := []*minipool.NodeStatus{}
    for mi := 0; mi < minipoolCount; mi++ {
        select {
            case nodeStatus := <-nodeStatusChannel[mi]:
                if (nodeStatus.Status == minipool.INITIALIZED || nodeStatus.Status == minipool.WITHDRAWN || nodeStatus.Status == minipool.TIMED_OUT) && nodeStatus.DepositExists {
                    withdrawableMinipools = append(withdrawableMinipools, nodeStatus)
                }
            case err := <-nodeStatusErrorChannel:
                return nil, err
        }
    }

    // Return
    return withdrawableMinipools, nil

}


// Check node deposits can be withdrawn from minipools
func CanWithdrawMinipools(p *services.Provider) (*CanWithdrawMinipoolsResponse, error) {

    // Response
    response := &CanWithdrawMinipoolsResponse{}

    // Check withdrawals are allowed
    withdrawalsAllowed := new(bool)
    if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, withdrawalsAllowed, "getWithdrawalAllowed"); err != nil {
        return false, errors.New("Error checking node withdrawals enabled status: " + err.Error())
    } else {
        response.WithdrawalsDisabled = !*withdrawalsAllowed
    }

    // Update & return response
    response.Success = !response.WithdrawalsDisabled
    return response, nil

}


// Check node deposit can be withdrawn from minipool
func CanWithdrawMinipool(p *services.Provider, minipoolAddress common.Address) (*CanWithdrawMinipoolResponse, error) {

    // Response
    response := &CanWithdrawMinipoolResponse{}

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()

    // Check contract code at minipool address
    if code, err := p.Client.CodeAt(context.Background(), minipoolAddress, nil); err != nil {
        return nil, errors.New("Error retrieving contract code at minipool address: " + err.Error())
    } else {
        response.MinipoolDidNotExist = (len(code) == 0)
    }

    // Check minipool exists
    if response.MinipoolDidNotExist {
        return response, nil
    }

    // Initialise minipool contract
    minipoolContract, err := p.CM.NewContract(&minipoolAddress, "rocketMinipool")
    if err != nil {
        return nil, errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Status channels
    withdrawalsDisabledChannel := make(chan bool)
    nodeOwnerChannel := make(chan common.Address)
    statusChannel := make(chan uint8)
    depositNotExistsChannel := make(chan bool)
    errorChannel := make(chan error)

    // Check withdrawals are allowed
    go (func() {
        withdrawalsAllowed := new(bool)
        if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, withdrawalsAllowed, "getWithdrawalAllowed"); err != nil {
            errorChannel <- errors.New("Error checking node withdrawals enabled status: " + err.Error())
        } else {
            withdrawalsDisabledChannel <- !*withdrawalsAllowed
        }
    })()

    // Get minipool node owner
    go (func() {
        nodeOwner := new(common.Address)
        if err := minipoolContract.Call(nil, nodeOwner, "getNodeOwner"); err != nil {
           errorChannel <- errors.New("Error retrieving minipool node owner: " + err.Error())
        } else {
            nodeOwnerChannel <- *nodeOwner
        }
    })()

    // Get minipool status
    go (func() {
        status := new(uint8)
        if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status: " + err.Error())
        } else {
            statusChannel <- *status
        }
    })()

    // Get node deposit status
    go (func() {
        nodeDepositExists := new(bool)
        if err := minipoolContract.Call(nil, nodeDepositExists, "getNodeDepositExists"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool node deposit status: " + err.Error())
        } else {
            depositNotExistsChannel <- !*nodeDepositExists
        }
    })()

    // Receive status
    for received := 0; received < 4; {
        select {
            case response.WithdrawalsDisabled = <-withdrawalsDisabledChannel:
                received++
            case response.NodeOwner = <-nodeOwnerChannel:
                received++
            case response.Status = <-statusChannel:
                received++
            case response.NodeDepositDidNotExist = <-depositNotExistsChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Update status
    response.InvalidNodeOwner = !bytes.Equal(response.NodeOwner.Bytes(), nodeAccount.Address.Bytes())
    response.InvalidStatus = !(response.Status == minipool.INITIALIZED || response.Status == minipool.WITHDRAWN || response.Status == minipool.TIMED_OUT)

    // Update & return response
    response.Success = !(response.MinipoolDidNotExist || response.WithdrawalsDisabled || response.InvalidNodeOwner || response.InvalidStatus || response.NodeDepositDidNotExist)
    return response, nil

}


// Withdraw node deposit from minipool
func WithdrawMinipool(p *services.Provider, minipoolAddress common.Address) (*WithdrawMinipoolResponse, error) {

    // Get account transactor
    txor, err := p.AM.GetNodeAccountTransactor()
    if err != nil { return nil, err }

    // Withdraw from minipool
    txReceipt, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "withdrawMinipoolDeposit", minipoolAddress)
    if err != nil {
        return nil, errors.New("Error withdrawing deposit: " + err.Error())
    }

    // Get withdrawal event
    nodeWithdrawalEvents, err := eth.GetTransactionEvents(p.Client, txReceipt, &minipoolAddress, p.CM.Abis["rocketMinipoolDelegateNode"], "NodeWithdrawal", NodeWithdrawal{})
    if err != nil {
        return nil, errors.New("Error retrieving node deposit withdrawal event: " + err.Error())
    } else if len(nodeWithdrawalEvents) == 0 {
        return nil, errors.New("Could not retrieve node deposit withdrawal event")
    }
    nodeWithdrawalEvent := (nodeWithdrawalEvents[0]).(*NodeWithdrawal)

    // Return response
    return &WithdrawMinipoolResponse{
        Success: true,
        EtherWithdrawnWei: nodeWithdrawalEvent.EtherAmount,
        RethWithdrawnWei: nodeWithdrawalEvent.RethAmount,
        RplWithdrawnWei: nodeWithdrawalEvent.RplAmount,
    }, nil

}

