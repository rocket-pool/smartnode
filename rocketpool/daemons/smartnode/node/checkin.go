package node

import (
    "bytes"
    "errors"
    "fmt"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/database"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
)


// Config
const CHECKIN_INTERVAL string = "5s"


// Shared vars
var checkinInterval, _ = time.ParseDuration(CHECKIN_INTERVAL)
var db = new(database.Database)
var nodeContract = new(bind.BoundContract)


// Start node checkin process
func StartCheckinProcess(c *cli.Context, errors chan error, fatalErrors chan error) {

    // Setup
    if err := setup(c); err != nil {
        fatalErrors <- err
        return
    }

    // Get last checkin time
    lastCheckinTime := new(int64)
    if err := db.Open(); err != nil {
        errors <- err
    } else {
        if err = db.Get("node.checkin.latest", lastCheckinTime); err != nil {
            *lastCheckinTime = 0
        }
        if err = db.Close(); err != nil {
            errors <- err
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
        checkin(db, checkinTimer, errors)
    }

}


// Perform node checkin
func checkin(db *database.Database, checkinTimer *time.Timer, errors chan error) {

    // Log
    fmt.Println("Checking in...")

    // Set last checkin time
    if err := db.Open(); err != nil {
        errors <- err
    } else {
        if err = db.Put("node.checkin.latest", time.Now().Unix()); err != nil {
            errors <- err
        }
        if err = db.Close(); err != nil {
            errors <- err
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
    am := accounts.NewAccountManager(c.GlobalString("keychain"))

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
    cm, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress"))
    if err != nil {
        return err
    }

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

