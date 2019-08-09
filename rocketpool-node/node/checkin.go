package node

import (
    "errors"
    "fmt"
    "math/big"
    "time"

    "github.com/shirou/gopsutil/cpu"
    "github.com/shirou/gopsutil/load"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Config
const DEFAULT_CHECKIN_INTERVAL string = "24h"
const NODE_FEE_VOTE_NO_CHANGE int64 = 0
const NODE_FEE_VOTE_INCREASE int64 = 1
const NODE_FEE_VOTE_DECREASE int64 = 2


// Checkin process
type CheckinProcess struct {
    p *services.Provider
}


/**
 * Start node checkin process
 */
func StartCheckinProcess(p *services.Provider) {

    // Initialise process
    process := &CheckinProcess{
        p: p,
    }

    // Schedule next checkin
    process.scheduleCheckin()

}


/**
 * Schedule next checkin
 */
func (p *CheckinProcess) scheduleCheckin() {

    // Wait for node to sync
    eth.WaitSync(p.p.Client, true, false)

    // Get last checkin time
    lastCheckinTime := new(int64)
    if err := p.p.DB.GetAtomic("node.checkin.latest", lastCheckinTime); err != nil {
        p.p.Log.Println(err)
    }

    // Get current checkin interval
    var checkinInterval time.Duration
    checkinIntervalSeconds := new(*big.Int)
    if err := p.p.CM.Contracts["rocketNodeSettings"].Call(nil, checkinIntervalSeconds, "getCheckinInterval"); err == nil {
        checkinInterval, _ = time.ParseDuration((*checkinIntervalSeconds).String() + "s")
    }
    if checkinInterval.Seconds() == 0 {
        checkinInterval, _ = time.ParseDuration(DEFAULT_CHECKIN_INTERVAL)
    }

    // Get next checkin time
    var nextCheckinTime time.Time
    if *lastCheckinTime == 0 {
        nextCheckinTime = time.Now()
    } else {
        nextCheckinTime = time.Unix(*lastCheckinTime, 0).Add(checkinInterval)
    }

    // Get time until next checkin & log
    nextCheckinDuration := time.Until(nextCheckinTime)
    if nextCheckinDuration.Seconds() > 0 {
        p.p.Log.Println("Time until next checkin:", nextCheckinDuration.String())
    }

    // Initialise checkin timer
    go (func() {
        checkinTimer := time.NewTimer(nextCheckinDuration)
        select {
            case <-checkinTimer.C:
                p.checkin()
        }
    })()

}


/**
 * Perform node checkin
 */
func (p *CheckinProcess) checkin() {

    // Log
    p.p.Log.Println("Checking in...")

    // Wait for node to sync
    eth.WaitSync(p.p.Client, true, false)

    // Get average server load
    serverLoad, err := p.getServerLoad()
    if err != nil {
        p.p.Log.Println(err)
    }

    // Get node fee vote
    nodeFeeVote, err := p.getNodeFeeVote()
    if err != nil {
        p.p.Log.Println(err)
    }

    // Checkin
    if txor, err := p.p.AM.GetNodeAccountTransactor(); err != nil {
        p.p.Log.Println(err)
    } else {
        if _, err := eth.ExecuteContractTransaction(p.p.Client, txor, p.p.NodeContractAddress, p.p.CM.Abis["rocketNodeContract"], "checkin", eth.EthToWei(serverLoad), big.NewInt(nodeFeeVote)); err != nil {
            p.p.Log.Println(errors.New("Error checking in with Rocket Pool: " + err.Error()))
        } else {
            p.p.Log.Println(fmt.Sprintf("Checked in successfully with an average load of %.2f%% and a node fee vote of '%s'", serverLoad * 100, getNodeFeeVoteType(nodeFeeVote)))
        }
    }

    // Set last checkin time
    if err := p.p.DB.PutAtomic("node.checkin.latest", time.Now().Unix()); err != nil {
        p.p.Log.Println(err)
    }

    // Schedule next checkin
    p.scheduleCheckin()

}


/**
 * Get the average server load
 */
func (p *CheckinProcess) getServerLoad() (float64, error) {

    // Server load
    var serverLoad float64

    // Get average load
    load, err := load.Avg()
    if err != nil {
        return serverLoad, errors.New("Error retrieving system CPU load: " + err.Error())
    }

    // Get CPU info
    cpus, err := cpu.Info()
    if err != nil {
        return serverLoad, errors.New("Error retrieving system CPU information: " + err.Error())
    }

    // Get number of CPU cores
    var cores int32 = 0
    for _, cpu := range cpus {
        cores += cpu.Cores
    }

    // Calculate server load
    serverLoad = load.Load15 / float64(cores)
    if serverLoad > 1 { serverLoad = 1 }

    // Return
    return serverLoad, nil

}


/**
 * Get the node fee vote based on current and target user fee
 */
func (p *CheckinProcess) getNodeFeeVote() (int64, error) {

    // Node fee vote
    nodeFeeVote := NODE_FEE_VOTE_NO_CHANGE

    // Get target user fee
    targetUserFeePerc := new(float64)
    *targetUserFeePerc = -1
    if err := p.p.DB.GetAtomic("user.fee.target", targetUserFeePerc); err != nil {
        return nodeFeeVote, err
    } else if *targetUserFeePerc == -1 {
        return nodeFeeVote, nil
    }

    // Load latest contracts
    if err := p.p.CM.LoadContracts([]string{"rocketNodeSettings"}); err != nil {
        return nodeFeeVote, err
    }

    // Get current user fee
    userFee := new(*big.Int)
    if err := p.p.CM.Contracts["rocketNodeSettings"].Call(nil, userFee, "getFeePerc"); err != nil {
        return nodeFeeVote, errors.New("Error retrieving node user fee percentage setting: " + err.Error())
    }
    userFeePerc := eth.WeiToEth(*userFee) * 100

    // Set node fee vote
    if userFeePerc < *targetUserFeePerc {
        nodeFeeVote = NODE_FEE_VOTE_INCREASE
    } else if userFeePerc > *targetUserFeePerc {
        nodeFeeVote = NODE_FEE_VOTE_DECREASE
    }

    // Return
    return nodeFeeVote, nil

}


// Get the node fee vote by value
func getNodeFeeVoteType(value int64) string {
    switch value {
        case NODE_FEE_VOTE_INCREASE: return "increase"
        case NODE_FEE_VOTE_DECREASE: return "decrease"
        default: return "no change"
    }
}

