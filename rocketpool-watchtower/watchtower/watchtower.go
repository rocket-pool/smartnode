package watchtower

import (
    "errors"
    "fmt"
    "math/big"
    "sync"
    "time"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Config
const CHECK_TRUSTED_INTERVAL string = "1m"
const DEFAULT_CHECK_MINIPOOLS_INTERVAL string = "1m"
var checkTrustedInterval, _ = time.ParseDuration(CHECK_TRUSTED_INTERVAL)


// Watchtower process
type WatchtowerProcess struct {
    p                    *services.Provider
    updatingMinipools    bool
    stopCheckMinipools   chan struct{}
    activeMinipools      map[string]common.Address
    txLock               sync.Mutex
}


/**
 * Start watchtower process
 */
func StartWatchtowerProcess(p *services.Provider) {

    // Initialise process
    process := &WatchtowerProcess{
        p:                    p,
        updatingMinipools:    false,
        stopCheckMinipools:   make(chan struct{}),
        activeMinipools:      make(map[string]common.Address),
    }

    // Start
    process.start()

}


/**
 * Start process
 */
func (p *WatchtowerProcess) start() {

    // Check if node is trusted on interval
    go (func() {
        p.checkTrusted()
        checkTrustedTimer := time.NewTicker(checkTrustedInterval)
        for _ = range checkTrustedTimer.C {
            p.checkTrusted()
        }
    })()

}


/**
 * Check if node is trusted
 */
func (p *WatchtowerProcess) checkTrusted() {

    // Wait for node to sync
    eth.WaitSync(p.p.Client, true, false)

    // Get trusted status
    nodeAccount, _ := p.p.AM.GetNodeAccount()
    trusted := new(bool)
    if err := p.p.CM.Contracts["rocketNodeAPI"].Call(nil, trusted, "getTrusted", nodeAccount.Address); err != nil {
        p.p.Log.Println(errors.New("Error retrieving node trusted status: " + err.Error()))
        return
    }

    // Start/stop minipool updates
    if *trusted {
        p.startUpdateMinipools()
    } else {
        p.stopUpdateMinipools()
    }

}


/**
 * Start minipool updates
 */
func (p *WatchtowerProcess) startUpdateMinipools() {

    // Cancel if already updating minipools
    if p.updatingMinipools { return }
    p.updatingMinipools = true

    // Log
    p.p.Log.Println("Node is trusted, starting watchtower process...")

    // Check minipools
    p.checkMinipools()

}


/**
 * Stop minipool updates
 */
func (p *WatchtowerProcess) stopUpdateMinipools() {

    // Cancel if not updating minipools
    if !p.updatingMinipools { return }
    p.updatingMinipools = false

    // Log
    p.p.Log.Println("Node is untrusted, stopping watchtower process...")

    // Stop checking minipools
    p.stopCheckMinipools <- struct{}{}

}


/**
 * Check active minipools
 */
func (p *WatchtowerProcess) checkMinipools() {

    // Log
    p.p.Log.Println("Checking active minipools...")

    // Wait for node to sync
    eth.WaitSync(p.p.Client, true, false)

    // Wait for beacon to sync
    // TODO: implement

    // Get active minipools
    if activeMinipools, err := minipool.GetActiveMinipoolsByValidatorPubkey(p.p.CM); err != nil {
        p.p.Log.Println(errors.New("Error getting active minipools: " + err.Error()))
        return
    } else {
        p.activeMinipools = *activeMinipools
    }

    // Get current beacon head
    head, err := p.p.Beacon.GetBeaconHead()
    if err != nil {
        p.p.Log.Println(errors.New("Error retrieving current beacon head: " + err.Error()))
        return
    }

    // Request validator statuses for active minipools
    for pubkey, minipoolAddress := range p.activeMinipools {
        go p.checkMinipool(minipoolAddress, pubkey, head)
    }

    // Schedule next minipool check
    p.scheduleCheckMinipools()

}


/**
 * Check minipool
 */
func (p *WatchtowerProcess) checkMinipool(minipoolAddress common.Address, pubkey string, head *beacon.BeaconHeadResponse) {

    // Log
    p.p.Log.Println(fmt.Sprintf("Checking minipool %s status...", minipoolAddress.Hex()))

    // Get minipool status
    status, err := getMinipoolStatus(p.p, &minipoolAddress)
    if err != nil {
        p.p.Log.Println(err)
        return
    }

    // Get validator status
    validator, err := p.p.Beacon.GetValidatorStatus(pubkey)
    if err != nil {
        p.p.Log.Println(errors.New("Error retrieving validator status: " + err.Error()))
        return
    } else if validator.Validator.ActivationEpoch == 0 {
        p.p.Log.Println(fmt.Sprintf("Minipool %s validator does not yet exist on beacon chain...", minipoolAddress.Hex()))
        return
    }

    // Log minipool out if validator has exited
    if status == minipool.STAKING && head.Epoch >= validator.Validator.ExitEpoch {
        if txor, err := p.p.AM.GetNodeAccountTransactor(); err != nil {
            p.p.Log.Println(err)
        } else {
            if _, err := eth.ExecuteContractTransaction(p.p.Client, txor, p.p.CM.Addresses["rocketNodeWatchtower"], p.p.CM.Abis["rocketNodeWatchtower"], "logoutMinipool", minipoolAddress); err != nil {
                p.p.Log.Println(errors.New("Error logging out minipool: " + err.Error()))
            } else {
                p.p.Log.Println(fmt.Sprintf("Minipool %s was successfully logged out", minipoolAddress.Hex()))
            }
        }
        return
    }

    // Withdraw minipool if validator is withdrawable
    if status == minipool.LOGGED_OUT && head.Epoch >= validator.Validator.WithdrawableEpoch {
        if txor, err := p.p.AM.GetNodeAccountTransactor(); err != nil {
            p.p.Log.Println(err)
        } else {
            validatorBalanceWei := eth.GweiToWei(float64(validator.Balance))
            if _, err := eth.ExecuteContractTransaction(p.p.Client, txor, p.p.CM.Addresses["rocketNodeWatchtower"], p.p.CM.Abis["rocketNodeWatchtower"], "withdrawMinipool", minipoolAddress, validatorBalanceWei); err != nil {
                p.p.Log.Println(errors.New("Error withdrawing minipool: " + err.Error()))
            } else {
                p.p.Log.Println(fmt.Sprintf("Minipool %s was successfully withdrawn with a balance of %.2f ETH", minipoolAddress.Hex(), eth.WeiToEth(validatorBalanceWei)))
            }
        }
        return
    }

}


/**
 * Schedule next minipool check
 */
func (p *WatchtowerProcess) scheduleCheckMinipools() {

    // Wait for node to sync
    eth.WaitSync(p.p.Client, true, false)

    // Get current check interval
    var checkInterval time.Duration
    checkIntervalSeconds := new(*big.Int)
    if err := p.p.CM.Contracts["rocketMinipoolSettings"].Call(nil, checkIntervalSeconds, "getMinipoolCheckInterval"); err == nil {
        checkInterval, _ = time.ParseDuration((*checkIntervalSeconds).String() + "s")
    }
    if checkInterval.Seconds() == 0 {
        checkInterval, _ = time.ParseDuration(DEFAULT_CHECK_MINIPOOLS_INTERVAL)
    }

    // Log check interval
    p.p.Log.Println("Time until next minipool check:", checkInterval.String())

    // Initialise check timer
    go (func() {
        checkMinipoolsTimer := time.NewTimer(checkInterval)
        select {
            case <-checkMinipoolsTimer.C:
                p.checkMinipools()
            case <-p.stopCheckMinipools:
                checkMinipoolsTimer.Stop()
        }
    })()

}


// Get a minipool's status
func getMinipoolStatus(p *services.Provider, address *common.Address) (uint8, error) {

    // Initialise minipool contract
    minipoolContract, err := p.CM.NewContract(address, "rocketMinipool")
    if err != nil {
        return 0, errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Get minipool's current status
    status := new(uint8)
    if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
        return 0, errors.New("Error retrieving minipool status: " + err.Error())
    }

    // Return
    return *status, nil

}

