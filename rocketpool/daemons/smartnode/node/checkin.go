package node

import (
    "errors"
    "fmt"
    "log"
    "math/big"
    "time"

    "github.com/shirou/gopsutil/cpu"
    "github.com/shirou/gopsutil/load"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Config
const CHECKIN_INTERVAL string = "15s"
const NODE_FEE_VOTE_NO_CHANGE int64 = 0
const NODE_FEE_VOTE_INCREASE int64 = 1
const NODE_FEE_VOTE_DECREASE int64 = 2


// Shared vars
var checkinInterval, _ = time.ParseDuration(CHECKIN_INTERVAL)
var checkinTimer *time.Timer
var p *services.Provider


// Start node checkin process
func StartCheckinProcess(provider *services.Provider) {

    // Set service provider
    p = provider

    // Get last checkin time
    lastCheckinTime := new(int64)
    if err := p.DB.GetAtomic("node.checkin.latest", lastCheckinTime); err != nil {
        log.Println(err)
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
        log.Println("Time until next checkin:", nextCheckinDuration.String())
    }

    // Initialise checkin timer
    go (func() {
        checkinTimer = time.NewTimer(nextCheckinDuration)
        for _ = range checkinTimer.C {
            checkin()
        }
    })()

}


// Perform node checkin
func checkin() {

    // Log
    log.Println("Checking in...")

    // Get average server load
    serverLoad, err := getServerLoad()
    if err != nil {
        log.Println(err)
    }

    // Get node fee vote
    nodeFeeVote, err := getNodeFeeVote()
    if err != nil {
        log.Println(err)
    }

    // Checkin
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        log.Println(err)
    } else {
        txor.GasLimit = 400000 // Gas estimates on this method are incorrect
        if _, err := p.NodeContract.Transact(txor, "checkin", eth.EthToWei(serverLoad), big.NewInt(nodeFeeVote)); err != nil {
            log.Println(errors.New("Error checking in with Rocket Pool: " + err.Error()))
        } else {
            log.Println(fmt.Sprintf("Checked in successfully with an average load of %.2f%% and a node fee vote of '%s'", serverLoad * 100, getNodeFeeVoteType(nodeFeeVote)))
        }
    }

    // Set last checkin time
    if err := p.DB.PutAtomic("node.checkin.latest", time.Now().Unix()); err != nil {
        log.Println(err)
    }

    // Log time until next checkin
    log.Println("Time until next checkin:", checkinInterval.String())

    // Reset timer for next checkin
    checkinTimer.Reset(checkinInterval)

}


// Get the average server load
func getServerLoad() (float64, error) {

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


// Get the node fee vote based on current and target user fee
func getNodeFeeVote() (int64, error) {

    // Node fee vote
    nodeFeeVote := NODE_FEE_VOTE_NO_CHANGE

    // Get target user fee
    targetUserFeePerc := new(float64)
    *targetUserFeePerc = -1
    if err := p.DB.GetAtomic("user.fee.target", targetUserFeePerc); err != nil {
        return nodeFeeVote, err
    } else if *targetUserFeePerc == -1 {
        return nodeFeeVote, nil
    }

    // Load latest contracts
    if err := p.CM.LoadContracts([]string{"rocketNodeSettings"}); err != nil {
        return nodeFeeVote, err
    }

    // Get current user fee
    userFee := new(*big.Int)
    if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, userFee, "getFeePerc"); err != nil {
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

