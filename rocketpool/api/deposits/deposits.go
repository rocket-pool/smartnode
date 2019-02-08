package deposits

import (
    "bytes"
    "errors"
    "fmt"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Get a node's current deposit status
func getDepositStatus(c *cli.Context) error {

    // Initialise account manager
    am := accounts.NewAccountManager(c.GlobalString("keychain"))

    // Get node account
    if !am.NodeAccountExists() {
        fmt.Println("Node account does not exist, please initialize with `rocketpool node init`")
        return nil
    }
    nodeAccount := am.GetNodeAccount()

    // Connect to ethereum node
    client, err := ethclient.Dial(c.GlobalString("provider"))
    if err != nil {
        return errors.New("Error connecting to ethereum node: " + err.Error())
    }

    // Initialise Rocket Pool contract manager
    rp, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress"))
    if err != nil {
        return err
    }

    // Load Rocket Pool node contracts
    err = rp.LoadContracts([]string{"rocketNodeAPI", "rocketNodeSettings"})
    if err != nil {
        return err
    }
    err = rp.LoadABIs([]string{"rocketNodeContract"})
    if err != nil {
        return err
    }

    // Check node is registered & get node contract address
    nodeContractAddress := new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address)
    if err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    }
    if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        fmt.Println("Node is not registered with Rocket Pool, please register with `rocketpool node register`")
        return nil
    }

    // Initialise node contract
    nodeContract, err := rp.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil {
        return errors.New("Error initialising node contract: " + err.Error())
    }

    // Get node ETH balance
    etherBalanceWei := new(*big.Int)
    err = nodeContract.Call(nil, etherBalanceWei, "getBalanceETH")
    if err != nil {
        return errors.New("Error retrieving node ETH balance: " + err.Error())
    }
    etherBalance := eth.WeiToEth(*etherBalanceWei)

    // Get node RPL balance
    rplBalanceWei := new(*big.Int)
    err = nodeContract.Call(nil, rplBalanceWei, "getBalanceRPL")
    if err != nil {
        return errors.New("Error retrieving node RPL balance: " + err.Error())
    }
    rplBalance := eth.WeiToEth(*rplBalanceWei)

    // Get node deposit reservation details
    var stakingDurationID string
    var etherRequired big.Int
    var rplRequired big.Int
    var expiryTime time.Time
    hasReservation := new(bool)
    err = nodeContract.Call(nil, hasReservation, "getHasDepositReservation")
    if err != nil {
        return errors.New("Error retrieving deposit reservation status: " + err.Error())
    }
    if *hasReservation {

        // Get deposit reservation duration ID
        durationID := new(string)
        err = nodeContract.Call(nil, durationID, "getDepositReserveDurationID")
        if err != nil {
            return errors.New("Error retrieving deposit reservation staking duration ID: " + err.Error())
        }
        stakingDurationID = *durationID

        // Get deposit reservation ETH required
        etherRequiredWei := new(*big.Int)
        err = nodeContract.Call(nil, etherRequiredWei, "getDepositReserveEtherRequired")
        if err != nil {
            return errors.New("Error retrieving deposit reservation ETH requirement: " + err.Error())
        }
        etherRequired = eth.WeiToEth(*etherRequiredWei)

        // Get deposit reservation RPL required
        rplRequiredWei := new(*big.Int)
        err = nodeContract.Call(nil, rplRequiredWei, "getDepositReserveRPLRequired")
        if err != nil {
            return errors.New("Error retrieving deposit reservation RPL requirement: " + err.Error())
        }
        rplRequired = eth.WeiToEth(*rplRequiredWei)

        // Get deposit reservation reserved time
        reservedTime := new(*big.Int)
        err = nodeContract.Call(nil, reservedTime, "getDepositReservedTime")
        if err != nil {
            return errors.New("Error retrieving deposit reservation reserved time: " + err.Error())
        }

        // Get reservation duration
        reservationTime := new(*big.Int)
        err = rp.Contracts["rocketNodeSettings"].Call(nil, reservationTime, "getDepositReservationTime")
        if err != nil {
            return errors.New("Error retrieving node deposit reservation time setting: " + err.Error())
        }

        // Get deposit reservation expiry time
        var expiryTimestamp big.Int
        expiryTimestamp.Add(*reservedTime, *reservationTime)
        expiryTime = time.Unix(expiryTimestamp.Int64(), 0)

    }

    // Log status & return
    fmt.Println(fmt.Sprintf("Node has a balance of %s ETH and %s RPL", etherBalance.String(), rplBalance.String()))
    if *hasReservation {
        fmt.Println(fmt.Sprintf(
            "Node has a deposit reservation requiring %s ETH and %s RPL, with a staking duration of %s and expiring at %s",
            etherRequired.String(),
            rplRequired.String(),
            stakingDurationID,
            expiryTime.Format("2006-01-02, 15:04 -0700 MST")))
    } else {
        fmt.Println("Node does not have a deposit reservation")
    }
    return nil

}

