package node

import (
    "errors"
    "log"
    "time"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/minipool"
)


// Config
const LOAD_VALIDATORS_INTERVAL string = "15s"


// Validator manager
type ValidatorManager struct {
    Validators []*minipool.Status
    am *accounts.AccountManager
    cm *rocketpool.ContractManager
    loadTimer *time.Ticker
}


/**
 * Create validator manager
 */
func NewValidatorManager(c *cli.Context, client *ethclient.Client) (*ValidatorManager, error) {

    // Create instance
    vm := &ValidatorManager{}

    // Setup
    if err := vm.setup(c, client); err != nil {
        return nil, err
    }

    // Return instance
    return vm, nil

}


/**
 * Start loading active validators periodically
 */
func (vm *ValidatorManager) StartLoad() {

    // Cancel if already loading
    if vm.loadTimer != nil {
        return
    }

    // Initialise load timer
    go (func() {
        vm.load()
        loadInterval, _ := time.ParseDuration(LOAD_VALIDATORS_INTERVAL)
        vm.loadTimer = time.NewTicker(loadInterval)
        for _ = range vm.loadTimer.C {
            vm.load()
        }
    })()

}


/**
 * Load active validators
 */
func (vm *ValidatorManager) load() {

    // Get minipool addresses
    minipoolAddresses, err := GetMinipoolAddresses(vm.am.GetNodeAccount().Address, vm.cm)
    if err != nil {
        log.Println(err)
    }
    minipoolCount := len(minipoolAddresses)

    // Get minipool statuses
    statusChannels := make([]chan *minipool.Status, minipoolCount)
    errorChannel := make(chan error)
    for mi := 0; mi < minipoolCount; mi++ {
        statusChannels[mi] = make(chan *minipool.Status)
        go (func(mi int) {
            if status, err := minipool.GetStatus(vm.cm, minipoolAddresses[mi]); err != nil {
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
                log.Println(err)
        }
    }

    // Set active validators
    vm.Validators = statuses

}


/**
 * Setup validator manager
 */
func (vm *ValidatorManager) setup(c *cli.Context, client *ethclient.Client) error {

    // Initialise account manager
    vm.am = accounts.NewAccountManager(c.GlobalString("keychain"))

    // Check node account
    if !vm.am.NodeAccountExists() {
        return errors.New("Node account does not exist, please initialize with `rocketpool node init`")
    }

    // Initialise Rocket Pool contract manager
    if cm, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress")); err != nil {
        return err
    } else {
        vm.cm = cm
    }

    // Loading channels
    successChannel := make(chan bool)
    errorChannel := make(chan error)

    // Load Rocket Pool contracts
    go (func() {
        if err := vm.cm.LoadContracts([]string{"utilAddressSetStorage"}); err != nil {
            errorChannel <- err
        } else {
            successChannel <- true
        }
    })()
    go (func() {
        if err := vm.cm.LoadABIs([]string{"rocketMinipool"}); err != nil {
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

