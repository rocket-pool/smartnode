package node

import (
    "bytes"
    "errors"
    "fmt"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/database"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Config
const CHECKIN_INTERVAL string = "5s"
const NODE_FEE_VOTE_NO_CHANGE int64 = 0
const NODE_FEE_VOTE_INCREASE int64 = 1
const NODE_FEE_VOTE_DECREASE int64 = 2


// Shared vars
var checkinInterval, _ = time.ParseDuration(CHECKIN_INTERVAL)
var db = new(database.Database)
var am = new(accounts.AccountManager)
var cm = new(rocketpool.ContractManager)
var nodeContract = new(bind.BoundContract)


// Start node checkin process
func StartCheckinProcess(c *cli.Context, errorChannel chan error, fatalErrorChannel chan error) {

    // Setup
    if err := setup(c); err != nil {
        fatalErrorChannel <- err
        return
    }

    // Get last checkin time
    lastCheckinTime := new(int64)
    if err := db.Open(); err != nil {
        errorChannel <- err
    } else {
        _ = db.Get("node.checkin.latest", lastCheckinTime)
        if err = db.Close(); err != nil {
            errorChannel <- err
        }
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
        fmt.Println("Time until next checkin:", nextCheckinDuration.String())
    }

    // Initialise checkin timer
    checkinTimer := time.NewTimer(nextCheckinDuration)
    for _ = range checkinTimer.C {
        checkin(db, checkinTimer, errorChannel)
    }

}


// Perform node checkin
func checkin(db *database.Database, checkinTimer *time.Timer, errorChannel chan error) {

    // Log
    fmt.Println("Checking in...")

    // Get target user fee
    targetUserFeePerc := new(float64)
    *targetUserFeePerc = -1
    if err := db.Open(); err != nil {
        errorChannel <- err
    } else {
        _ = db.Get("user.fee.target", targetUserFeePerc)
        if err = db.Close(); err != nil {
            errorChannel <- err
        }
    }

    // Get node fee vote
    nodeFeeVote := NODE_FEE_VOTE_NO_CHANGE
    if *targetUserFeePerc != -1 {

        // Load latest contracts
        if err := cm.LoadContracts([]string{"rocketNodeSettings"}); err != nil {
            errorChannel <- err
        } else {

            // Get current user fee
            userFee := new(*big.Int)
            if err = cm.Contracts["rocketNodeSettings"].Call(nil, userFee, "getFeePerc"); err != nil {
                errorChannel <- errors.New("Error retrieving node user fee percentage setting: " + err.Error())
            } else {

                // Set node fee vote
                userFeePerc := eth.WeiToEth(*userFee) * 100
                if userFeePerc < *targetUserFeePerc {
                    nodeFeeVote = NODE_FEE_VOTE_INCREASE
                } else if userFeePerc > *targetUserFeePerc {
                    nodeFeeVote = NODE_FEE_VOTE_DECREASE
                }

            }

        }

    }

    // Checkin
    if nodeAccountTransactor, err := am.GetNodeAccountTransactor(); err != nil {
        errorChannel <- err
    } else {
        nodeAccountTransactor.GasLimit = 200000 // Gas estimates on this method are incorrect
        if _, err = nodeContract.Transact(nodeAccountTransactor, "checkin", big.NewInt(0), big.NewInt(nodeFeeVote)); err != nil {
            errorChannel <- errors.New("Error checking in with Rocket Pool: " + err.Error())
        } else {
            fmt.Println(fmt.Sprintf("Checked in successfully with average load of %.2f and node fee vote of %d", 0.00, nodeFeeVote))
        }
    }

    // Set last checkin time
    if err := db.Open(); err != nil {
        errorChannel <- err
    } else {
        if err = db.Put("node.checkin.latest", time.Now().Unix()); err != nil {
            errorChannel <- err
        }
        if err = db.Close(); err != nil {
            errorChannel <- err
        }
    }

    // Log time until next checkin
    fmt.Println("Time until next checkin:", checkinInterval.String())

    // Reset timer for next checkin
    checkinTimer.Reset(checkinInterval)

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
    cmV, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress"))
    if err != nil {
        return err
    }
    *cm = *cmV

    // Loading channels
    successChannel := make(chan bool)
    errorChannel := make(chan error)

    // Load Rocket Pool contracts
    go (func() {
        err := cm.LoadContracts([]string{"rocketNodeAPI"})
        if err != nil {
            errorChannel <- err
        } else {
            successChannel <- true
        }
    })()
    go (func() {
        err := cm.LoadABIs([]string{"rocketNodeContract"})
        if err != nil {
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
    err = cm.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", am.GetNodeAccount().Address)
    if err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    }
    if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        return errors.New("Node is not registered with Rocket Pool, please register with `rocketpool node register`")
    }

    // Initialise node contract
    nodeContractV, err := cm.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil {
        return errors.New("Error initialising node contract: " + err.Error())
    }
    *nodeContract = *nodeContractV

    // Return
    return nil

}

