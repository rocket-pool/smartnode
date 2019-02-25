package node

import (
    "log"
    "time"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/minipool"
)


// Config
const LOAD_VALIDATORS_INTERVAL string = "15s"
var loadValidatorsInterval, _ = time.ParseDuration(LOAD_VALIDATORS_INTERVAL)


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
func NewValidatorManager(am *accounts.AccountManager, cm *rocketpool.ContractManager) *ValidatorManager {
    return &ValidatorManager{
        am: am,
        cm: cm,
    }
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
        vm.loadTimer = time.NewTicker(loadValidatorsInterval)
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
        return
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
                if status.Status != minipool.STAKING { break }
                statuses = append(statuses, status)
            case err := <-errorChannel:
                log.Println(err)
                return
        }
    }

    // Set active validators
    vm.Validators = statuses

}

