package minipools

import (
    "sync"
    "time"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Config
const CHECK_MINIPOOLS_INTERVAL string = "7m" // ~ 1 epoch
var checkMinipoolsInterval, _ = time.ParseDuration(CHECK_MINIPOOLS_INTERVAL)


// Minipools process
type MinipoolsProcess struct {
    p       *services.Provider
    txLock  sync.Mutex
}


/**
 * Start minipools process
 */
func StartMinipoolsProcess(p *services.Provider) {

    // Initialise process
    process := &MinipoolsProcess{
        p: p,
    }

    // Start
    process.start()

}


/**
 * Start process
 */
func (p *MinipoolsProcess) start() {

    // Check minipools on interval
    go (func() {
        p.checkMinipools()
        checkMinipoolsTimer := time.NewTicker(checkMinipoolsInterval)
        for _ = range checkMinipoolsTimer.C {
            p.checkMinipools()
        }
    })()

}


/**
 * Check minipools
 */
func (p *MinipoolsProcess) checkMinipools() {

    // Wait for node to sync
    eth.WaitSync(p.p.Client, true, false)

    // Wait for beacon to sync
    // TODO: implement

    // Get minipool addresses
    nodeAccount, _ := p.p.AM.GetNodeAccount()
    minipoolAddresses, err := node.GetMinipoolAddresses(nodeAccount.Address, p.p.CM)
    if err != nil {
        p.p.Log.Println(err)
        return
    }
    minipoolCount := len(minipoolAddresses)

    // Get minipool statuses
    statusChannels := make([]chan uint8, minipoolCount)
    errorChannel := make(chan error)
    for mi := 0; mi < minipoolCount; mi++ {
        statusChannels[mi] = make(chan uint8)
        go (func(mi int) {
            if status, err := minipool.GetStatusCode(p.p.CM, minipoolAddresses[mi]); err != nil {
                errorChannel <- err
            } else {
                statusChannels[mi] <- status
            }
        })(mi)
    }

    // Receive minipool statuses and filter prelaunch & staking minipools
    prelaunchMinipoolAddresses := []*common.Address{}
    stakingMinipoolAddresses := []*common.Address{}
    for mi := 0; mi < minipoolCount; mi++ {
        select {
            case status := <-statusChannels[mi]:
                if status == minipool.PRELAUNCH { prelaunchMinipoolAddresses = append(prelaunchMinipoolAddresses, minipoolAddresses[mi]) }
                if status == minipool.STAKING { stakingMinipoolAddresses = append(stakingMinipoolAddresses, minipoolAddresses[mi]) }
            case err := <-errorChannel:
                p.p.Log.Println(err)
                return
        }
    }

    // Handle minipool staking & withdrawals
    go p.stakePrelaunchMinipools(prelaunchMinipoolAddresses)
    go p.checkStakingMinipools(stakingMinipoolAddresses)

}

