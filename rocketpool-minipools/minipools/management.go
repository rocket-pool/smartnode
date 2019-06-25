package minipools

import (
    "log"
    "time"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
)


// Config
const CHECK_MINIPOOLS_INTERVAL string = "15s"
var checkMinipoolsInterval, _ = time.ParseDuration(CHECK_MINIPOOLS_INTERVAL)


// Management process
type ManagementProcess struct {
    p *services.Provider
}


/**
 * Start minipools management process
 */
func StartManagementProcess(p *services.Provider) {

    // Initialise process
    process := &ManagementProcess{
        p: p,
    }

    // Start
    process.start()

}


/**
 * Start process
 */
func (p *ManagementProcess) start() {

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
func (p *ManagementProcess) checkMinipools() {

    // Get minipool addresses
    minipoolAddresses, err := node.GetMinipoolAddresses(p.p.AM.GetNodeAccount().Address, p.p.CM)
    if err != nil {
        log.Println(err)
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

    // Receive minipool statuses & filter staking minipools
    activeMinipoolAddresses := []*common.Address{}
    for mi := 0; mi < minipoolCount; mi++ {
        select {
            case status := <-statusChannels[mi]:
                if status == minipool.STAKING { activeMinipoolAddresses = append(activeMinipoolAddresses, minipoolAddresses[mi]) }
            case err := <-errorChannel:
                log.Println(err)
                return
        }
    }

}

