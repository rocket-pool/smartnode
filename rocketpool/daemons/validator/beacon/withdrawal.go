package beacon

import (
    "context"
    "encoding/hex"
    "errors"
    "fmt"
    "log"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
)


// Config
const CHECK_MINIPOOLS_INTERVAL string = "15s"


// Shared vars
var stakingMinipools []*minipool.Status

// Process vars
var am = new(accounts.AccountManager)
var client = new(ethclient.Client)
var cm = new(rocketpool.ContractManager)


// Start beacon withdrawal process
func StartWithdrawalProcess(c *cli.Context, fatalErrorChannel chan error) {

    // Setup
    if err := setup(c); err != nil {
        fatalErrorChannel <- err
        return
    }

    // Get staking minipool statuses
    // :TODO: reload minipools periodically
    if err := getStakingMinipools(); err != nil {
        fatalErrorChannel <- err
        return
    }

    // Check staking minipools for withdrawal on interval
    checkMinipoolsInterval, _ := time.ParseDuration(CHECK_MINIPOOLS_INTERVAL)
    checkMinipoolsTicker := time.NewTicker(checkMinipoolsInterval)
    for _ = range checkMinipoolsTicker.C {
        checkStakingMinipools()
    }

}


// Get staking minipool statuses
func getStakingMinipools() error {

    // Get minipool addresses
    minipoolAddresses, err := node.GetMinipoolAddresses(am.GetNodeAccount().Address, cm)
    if err != nil {
        return err
    }
    minipoolCount := len(minipoolAddresses)

    // Get minipool statuses
    statusChannels := make([]chan *minipool.Status, minipoolCount)
    errorChannel := make(chan error)
    for mi := 0; mi < minipoolCount; mi++ {
        statusChannels[mi] = make(chan *minipool.Status)
        go (func(mi int) {
            if status, err := minipool.GetStatus(cm, minipoolAddresses[mi]); err != nil {
                errorChannel <- err
            } else {
                statusChannels[mi] <- status
            }
        })(mi)
    }

    // Receive staking minipool statuses
    statuses := make([]*minipool.Status, 0)
    for mi := 0; mi < minipoolCount; mi++ {
        select {
            case status := <-statusChannels[mi]:
                if status.Status != 2 { break } // Staking
                statuses = append(statuses, status)
            case err := <-errorChannel:
                return err
        }
    }

    // Set staking minipools & return
    stakingMinipools = statuses
    return nil

}


// Check staking minipools for withdrawal
func checkStakingMinipools() {

    // Get latest block header
    header, err := client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        log.Println(errors.New("Error retrieving latest block header: " + err.Error()))
        return
    }

    // Log
    log.Println(fmt.Sprintf("Checking staking minipools for withdrawal at block %s...", header.Number.String()))

    // Check minipools
    for _, minipool := range stakingMinipools {
        var exitBlock big.Int
        exitBlock.Add(minipool.StatusBlock, minipool.StakingDuration)
        if header.Number.Cmp(&exitBlock) > -1 {
            log.Println(fmt.Sprintf("Validator %s ready to withdraw, since block %s...", hex.EncodeToString(minipool.ValidatorPubkey), exitBlock.String()))
        } else {
            log.Println(fmt.Sprintf("Validator %s not ready to withdraw until block %s...", hex.EncodeToString(minipool.ValidatorPubkey), exitBlock.String()))
        }
    }

}


// Set up beacon withdrawal process
func setup(c *cli.Context) error {

    // Initialise account manager
    *am = *accounts.NewAccountManager(c.GlobalString("keychain"))

    // Check node account
    if !am.NodeAccountExists() {
        return errors.New("Node account does not exist, please initialize with `rocketpool node init`")
    }

    // Connect to ethereum node
    if clientV, err := ethclient.Dial(c.GlobalString("provider")); err != nil {
        return errors.New("Error connecting to ethereum node: " + err.Error())
    } else {
        *client = *clientV
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
        if err := cm.LoadContracts([]string{"utilAddressSetStorage"}); err != nil {
            errorChannel <- err
        } else {
            successChannel <- true
        }
    })()
    go (func() {
        if err := cm.LoadABIs([]string{"rocketMinipool"}); err != nil {
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

    // Return
    return nil

}

