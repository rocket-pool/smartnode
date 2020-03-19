package minipool

import (
    "context"
    "errors"
    "fmt"
    "math/big"
    "time"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Config
const CHECK_MINIPOOL_INTERVAL string = "1m"
var checkMinipoolInterval, _ = time.ParseDuration(CHECK_MINIPOOL_INTERVAL)


// Withdrawal process
type WithdrawalProcess struct {
    p *services.Provider
    minipool *Minipool
    minipoolExiting bool
    stop chan struct{}
    done chan struct{}
}


/**
 * Start withdrawal process
 */
func StartWithdrawalProcess(p *services.Provider, minipool *Minipool, done chan struct{}) {

    // Initialise process
    process := &WithdrawalProcess{
        p: p,
        minipool: minipool,
        minipoolExiting: false,
        stop: make(chan struct{}),
        done: done,
    }

    // Start
    process.start()

}


/**
 * Start process
 */
func (p *WithdrawalProcess) start() {

    // Check minipool for withdrawal on interval while checking
    go (func() {
        checkMinipoolsTimer := time.NewTicker(checkMinipoolInterval)
        checking := true
        for checking {
            select {
                case <-checkMinipoolsTimer.C:
                    p.checkWithdrawal()
                case <-p.stop:
                    checkMinipoolsTimer.Stop()
                    checking = false
            }
        }
    })()

    // Block thread until done
    select {
        case <-p.stop:
            p.p.Log.Println(fmt.Sprintf("Ending minipool %s withdrawal process...", p.minipool.Address.Hex()))
            p.done <- struct{}{}
    }

}


/**
 * Check minipool for withdrawal
 */
func (p *WithdrawalProcess) checkWithdrawal() {

    // Wait for node to sync
    eth.WaitSync(p.p.Client, true, false)

    // Check minipool contract still exists
    if code, err := p.p.Client.CodeAt(context.Background(), *(p.minipool.Address), nil); err != nil {
        p.p.Log.Println(errors.New("Error retrieving contract code at minipool address: " + err.Error()))
        return
    } else if len(code) == 0 {
        p.p.Log.Println(fmt.Sprintf("Minipool %s no longer exists...", p.minipool.Address.Hex()))
        close(p.stop)
        return
    }

    // Get latest block header
    header, err := p.p.Client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        p.p.Log.Println(errors.New("Error retrieving latest block header: " + err.Error()))
        return
    }

    // Get minipool status
    status, err := minipool.GetStatus(p.p.CM, p.minipool.Address)
    if err != nil {
        p.p.Log.Println(errors.New("Error retrieving minipool status: " + err.Error()))
        return
    }

    // Log
    p.p.Log.Println(fmt.Sprintf("Checking minipool %s for withdrawal at block %s...", p.minipool.Address.Hex(), header.Number.String()))

    // Check minipool status
    if status.Status > minipool.STAKING {
        p.p.Log.Println(fmt.Sprintf("Minipool %s has already progressed beyond staking...", p.minipool.Address.Hex()))
        close(p.stop)
        return
    } else if status.Status < minipool.STAKING {
        p.p.Log.Println(fmt.Sprintf("Minipool %s is not staking yet...", p.minipool.Address.Hex()))
        return
    }

    // Get minipool exit block
    var exitBlock big.Int
    exitBlock.Add(status.StatusBlock, status.StakingDuration)

    // Check exit block
    if header.Number.Cmp(&exitBlock) == -1 {
        p.p.Log.Println(fmt.Sprintf("Minipool %s is not ready to withdraw until block %s...", p.minipool.Address.Hex(), exitBlock.String()))
        return
    }

    // Check if already marked for exit
    if p.minipoolExiting { return }

    // Mark minipool for exit and log
    p.minipoolExiting = true
    p.p.Log.Println(fmt.Sprintf("Minipool %s is ready to withdraw, since block %s...", p.minipool.Address.Hex(), exitBlock.String()))

    // Exit validator
    // TODO:
    //   - request validator status & check if active
    //   - send voluntary exit message
    //   - poll validator status until exit initiated

}

