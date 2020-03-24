package minipool

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
    hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
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
        checkMinipoolTimer := time.NewTicker(checkMinipoolInterval)
        checking := true
        for checking {
            select {
                case <-checkMinipoolTimer.C:
                    p.checkWithdrawal()
                case <-p.stop:
                    checkMinipoolTimer.Stop()
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

    // Wait for beacon to sync
    // TODO: implement

    // Check minipool contract still exists
    if code, err := p.p.Client.CodeAt(context.Background(), *(p.minipool.Address), nil); err != nil {
        p.p.Log.Println(errors.New("Error retrieving contract code at minipool address: " + err.Error()))
        return
    } else if len(code) == 0 {
        p.p.Log.Println(fmt.Sprintf("Minipool %s no longer exists...", p.minipool.Address.Hex()))
        close(p.stop)
        return
    }

    // Get minipool status
    status, err := minipool.GetStatus(p.p.CM, p.minipool.Address)
    if err != nil {
        p.p.Log.Println(errors.New("Error retrieving minipool status: " + err.Error()))
        return
    }

    // Check minipool status
    if status.Status > minipool.STAKING {
        p.p.Log.Println(fmt.Sprintf("Minipool %s has already progressed beyond staking...", p.minipool.Address.Hex()))
        close(p.stop)
        return
    } else if status.Status < minipool.STAKING {
        p.p.Log.Println(fmt.Sprintf("Minipool %s is not staking yet...", p.minipool.Address.Hex()))
        return
    }

    // Get current beacon head
    head, err := p.p.Beacon.GetBeaconHead()
    if err != nil {
        p.p.Log.Println(errors.New("Error retrieving current beacon head: " + err.Error()))
        return
    }

    // Log
    p.p.Log.Println(fmt.Sprintf("Checking minipool %s for withdrawal at epoch %d...", p.minipool.Address.Hex(), head.Epoch))

    // Get & check validator status; get minipool exit epoch
    validator, err := p.p.Beacon.GetValidatorStatus(hexutil.AddPrefix(p.minipool.Pubkey))
    if err != nil {
        p.p.Log.Println(errors.New("Error retrieving validator status: " + err.Error()))
        return
    } else if !validator.Exists {
        p.p.Log.Println(fmt.Sprintf("Minipool %s validator does not yet exist on beacon chain...", p.minipool.Address.Hex()))
        return
    }
    exitEpoch := validator.Validator.ActivationEpoch + status.StakingDuration.Uint64()

    // Check exit epoch
    if head.Epoch < exitEpoch {
        p.p.Log.Println(fmt.Sprintf("Minipool %s is not ready to withdraw until epoch %d...", p.minipool.Address.Hex(), exitEpoch))
        return
    }

    // Check if already marked for exit
    if p.minipoolExiting { return }

    // Mark minipool for exit and log
    p.minipoolExiting = true
    p.p.Log.Println(fmt.Sprintf("Minipool %s is ready to withdraw, since epoch %d...", p.minipool.Address.Hex(), exitEpoch))

    // Withdraw minipool
    p.withdraw()

}


/**
 * Withdraw minipool
 */
func (p *WithdrawalProcess) withdraw() {

    // Log
    p.p.Log.Println("Minipool withdrawal process not yet implemented...")

}

