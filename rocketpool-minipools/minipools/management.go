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

    // Data channels
    stakingMinipoolAddressesChannel := make(chan []*common.Address)
    errorChannel := make(chan error)

    // Get staking minipool addresses
    go (func() {

        // Get minipool addresses
        minipoolAddresses, err := node.GetMinipoolAddresses(p.p.AM.GetNodeAccount().Address, p.p.CM)
        if err != nil {
            errorChannel <- err
            return
        }
        minipoolCount := len(minipoolAddresses)

        // Get minipool statuses
        statusChannels := make([]chan uint8, minipoolCount)
        statusErrorChannel := make(chan error)
        for mi := 0; mi < minipoolCount; mi++ {
            statusChannels[mi] = make(chan uint8)
            go (func(mi int) {
                if status, err := minipool.GetStatusCode(p.p.CM, minipoolAddresses[mi]); err != nil {
                    statusErrorChannel <- err
                } else {
                    statusChannels[mi] <- status
                }
            })(mi)
        }

        // Receive minipool statuses & filter staking minipools
        stakingMinipoolAddresses := []*common.Address{}
        for mi := 0; mi < minipoolCount; mi++ {
            select {
                case status := <-statusChannels[mi]:
                    if status == minipool.STAKING { stakingMinipoolAddresses = append(stakingMinipoolAddresses, minipoolAddresses[mi]) }
                case err := <-statusErrorChannel:
                    errorChannel <- err
                    return
            }
        }

        // Send staking minipool addresses
        stakingMinipoolAddressesChannel <- stakingMinipoolAddresses

    })()

    // Receive minipool data
    var stakingMinipoolAddresses []*common.Address
    for received := 0; received < 1; {
        select {
            case stakingMinipoolAddresses = <-stakingMinipoolAddressesChannel:
                received++
            case err := <-errorChannel:
                log.Println(err)
                return
        }
    }

}

