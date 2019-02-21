package node

import (
    "bytes"
    "errors"
    "fmt"
    "log"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/shirou/gopsutil/cpu"
    "github.com/shirou/gopsutil/load"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/database"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
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
var db = new(database.Database)
var am = new(accounts.AccountManager)
var cm = new(rocketpool.ContractManager)
var nodeContract = new(bind.BoundContract)


// Start node checkin process
func StartCheckinProcess(c *cli.Context, fatalErrorChannel chan error) {

    // Setup
    if err := setup(c); err != nil {
        fatalErrorChannel <- err
        return
    }

    // Get last checkin time
    lastCheckinTime := new(int64)
    if err := db.GetAtomic("node.checkin.latest", lastCheckinTime); err != nil {
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
    checkinTimer = time.NewTimer(nextCheckinDuration)
    for _ = range checkinTimer.C {
        checkin()
    }

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
    if txor, err := am.GetNodeAccountTransactor(); err != nil {
        log.Println(err)
    } else {
        txor.GasLimit = 400000 // Gas estimates on this method are incorrect
        if _, err := nodeContract.Transact(txor, "checkin", eth.EthToWei(serverLoad), big.NewInt(nodeFeeVote)); err != nil {
            log.Println(errors.New("Error checking in with Rocket Pool: " + err.Error()))
        } else {
            log.Println(fmt.Sprintf("Checked in successfully with an average load of %.2f%% and a node fee vote of '%s'", serverLoad * 100, getNodeFeeVoteType(nodeFeeVote)))
        }
    }

    // Set last checkin time
    if err := db.PutAtomic("node.checkin.latest", time.Now().Unix()); err != nil {
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
    if err := db.GetAtomic("user.fee.target", targetUserFeePerc); err != nil {
        return nodeFeeVote, err
    } else if *targetUserFeePerc == -1 {
        return nodeFeeVote, nil
    }

    // Load latest contracts
    if err := cm.LoadContracts([]string{"rocketNodeSettings"}); err != nil {
        return nodeFeeVote, err
    }

    // Get current user fee
    userFee := new(*big.Int)
    if err := cm.Contracts["rocketNodeSettings"].Call(nil, userFee, "getFeePerc"); err != nil {
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


// Set up node checkin process
func setup(c *cli.Context) error {

    // Initialise database
    *db = *database.NewDatabase(c.GlobalString("database"))

    // Initialise account manager
    *am = *accounts.NewAccountManager(c.GlobalString("keychain"))

    // Check node account
    if !am.NodeAccountExists() {
        return errors.New("Node account does not exist, please initialize with `rocketpool node init`")
    }

    // Connect to ethereum node
    client, err := ethclient.Dial(c.GlobalString("provider"))
    if err != nil {
        return errors.New("Error connecting to ethereum node: " + err.Error())
    }

    // Initialise Rocket Pool contract manager
    if cmV, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress")); err != nil {
        return err
    } else {
        *cm = *cmV
    }

    // Loading channels
    successChannel := make(chan bool)
    errorChannel := make(chan error)

    // Load Rocket Pool contracts
    go (func() {
        if err := cm.LoadContracts([]string{"rocketNodeAPI"}); err != nil {
            errorChannel <- err
        } else {
            successChannel <- true
        }
    })()
    go (func() {
        if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil {
            errorChannel <- err
        } else {
            successChannel <- true
        }
    })()

    // Await loading
    for received := 0; received < 2; {
        select {
            case <-successChannel:
                received++
            case err := <-errorChannel:
                return err
        }
    }

    // Check node is registered & get node contract address
    nodeContractAddress := new(common.Address)
    if err := cm.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", am.GetNodeAccount().Address); err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    } else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        return errors.New("Node is not registered with Rocket Pool, please register with `rocketpool node register`")
    }

    // Initialise node contract
    if nodeContractV, err := cm.NewContract(nodeContractAddress, "rocketNodeContract"); err != nil {
        return errors.New("Error initialising node contract: " + err.Error())
    } else {
        *nodeContract = *nodeContractV
    }

    // Return
    return nil

}

