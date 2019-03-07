package beacon

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/fatih/color"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
    beaconchain "github.com/rocket-pool/smartnode-cli/rocketpool/services/beacon-chain"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/minipool"
)


// Config
const WATCHTOWER_LOG_COLOR = color.FgRed
const CHECK_TRUSTED_INTERVAL string = "15s"
const GET_ACTIVE_MINIPOOLS_INTERVAL string = "15s"
var checkTrustedInterval, _ = time.ParseDuration(CHECK_TRUSTED_INTERVAL)
var getActiveMinipoolsInterval, _ = time.ParseDuration(GET_ACTIVE_MINIPOOLS_INTERVAL)


// Watchtower process
type WatchtowerProcess struct {
    c func(a ...interface{}) string
    p *services.Provider
    updatingMinipools bool
    getActiveMinipoolsStop chan bool
    beaconMessageChannel chan interface{}
    activeMinipools map[string]common.Address
}


/**
 * Start beacon watchtower process
 */
func StartWatchtowerProcess(p *services.Provider) {

    // Initialise process
    watchtowerProcess := &WatchtowerProcess{
        c: color.New(WATCHTOWER_LOG_COLOR).SprintFunc(),
        p: p,
        updatingMinipools: false,
        getActiveMinipoolsStop: make(chan bool),
        beaconMessageChannel: make(chan interface{}),
        activeMinipools: make(map[string]common.Address),
    }

    // Start
    watchtowerProcess.start()

}


/**
 * Start process
 */
func (w *WatchtowerProcess) start() {

    // Check if node is trusted on interval
    go (func() {
        w.checkTrusted()
        checkTrustedTimer := time.NewTicker(checkTrustedInterval)
        for _ = range checkTrustedTimer.C {
            w.checkTrusted()
        }
    })()

    // Handle beacon chain events
    go (func() {
        for eventData := range w.beaconMessageChannel {
            event := (eventData).(struct{Client *beaconchain.Client; Message []byte})
            w.onBeaconClientMessage(event.Message)
        }
    })()

}


/**
 * Check if node is trusted
 */
func (w *WatchtowerProcess) checkTrusted() {

    // Get trusted status
    trusted := new(bool)
    if err := w.p.CM.Contracts["rocketAdmin"].Call(nil, trusted, "getNodeTrusted", w.p.AM.GetNodeAccount().Address); err != nil {
        log.Println(w.c(errors.New("Error retrieving node trusted status: " + err.Error())))
        return
    }

    // Start/stop minipool updates
    if *trusted {
        w.startUpdateMinipools()
    } else {
        w.stopUpdateMinipools()
    }

}


/**
 * Start minipool updates
 */
func (w *WatchtowerProcess) startUpdateMinipools() {

    // Cancel if already updating minipools
    if w.updatingMinipools { return }
    w.updatingMinipools = true

    // Get active minipools on interval
    go (func() {
        w.getActiveMinipools()
        getActiveMinipoolsTimer := time.NewTicker(getActiveMinipoolsInterval)
        for {
            select {
                case <-getActiveMinipoolsTimer.C:
                    w.getActiveMinipools()
                case <-w.getActiveMinipoolsStop:
                    getActiveMinipoolsTimer.Stop()
                    return
            }
        }
    })()

    // Subscribe to beacon chain events
    w.p.Publisher.AddSubscriber("beacon.client.message", w.beaconMessageChannel)

}


/**
 * Stop minipool updates
 */
func (w *WatchtowerProcess) stopUpdateMinipools() {

    // Cancel if not updating minipools
    if !w.updatingMinipools { return }
    w.updatingMinipools = false

    // Stop getting active minipools
    w.getActiveMinipoolsStop <- true

    // Unsubscribe from beacon chain events
    w.p.Publisher.RemoveSubscriber("beacon.client.message", w.beaconMessageChannel)

}


/**
 * Get active minipools by validator pubkey
 */
func (w *WatchtowerProcess) getActiveMinipools() {

    // Get active minipools
    if activeMinipools, err := minipool.GetActiveMinipoolsByValidatorPubkey(w.p.CM); err != nil {
        log.Println(w.c(errors.New("Error getting active minipools: " + err.Error())))
        return
    } else {
        w.activeMinipools = *activeMinipools
    }

    // Request validator statuses for active minipools
    for pubkey, _ := range w.activeMinipools {
        if payload, err := json.Marshal(beaconchain.ClientMessage{
            Message: "get_validator_status",
            Pubkey: pubkey,
        }); err != nil {
            log.Println(w.c(errors.New("Error encoding get validator status payload: " + err.Error())))
        } else if err := w.p.Beacon.Send(payload); err != nil {
            log.Println(w.c(errors.New("Error sending get validator status message: " + err.Error())))
        }
    }

}


/**
 * Handle beacon chain client messages
 */
func (w *WatchtowerProcess) onBeaconClientMessage(messageData []byte) {

    // Parse message
    message := new(beaconchain.ServerMessage)
    if err := json.Unmarshal(messageData, message); err != nil {
        log.Println(w.c(errors.New("Error decoding beacon message: " + err.Error())))
        return
    }

    // Handle exit and withdrawal validator status messages only
    if message.Message != "validator_status" { return }
    if !(message.Status.Code == "exited" || message.Status.Code == "withdrawable") { return }

    // Get associated minipool
    minipoolAddress, ok := w.activeMinipools[message.Pubkey]
    if !ok { return }

    // Initialise minipool contract
    minipoolContract, err := w.p.CM.NewContract(&minipoolAddress, "rocketMinipool")
    if err != nil {
        log.Println(w.c(errors.New("Error initialising minipool contract: " + err.Error())))
        return
    }

    // Get minipool's current status
    status := new(uint8)
    minipoolContract.Call(nil, status, "getStatus")
    if err != nil {
        log.Println(w.c(errors.New("Error retrieving minipool status: " + err.Error())))
        return
    }

    // Log minipool out if staking
    if *status == minipool.STAKING {
        if txor, err := w.p.AM.GetNodeAccountTransactor(); err != nil {
            log.Println(w.c(err))
        } else {
            txor.GasLimit = 100000 // Gas estimates on this method are incorrect
            if _, err := w.p.CM.Contracts["rocketNodeWatchtower"].Transact(txor, "logoutMinipool", minipoolAddress); err != nil {
                log.Println(w.c(errors.New("Error logging out minipool: " + err.Error())))
            } else {
                log.Println(w.c(fmt.Sprintf("Minipool %s was logged out", minipoolAddress.Hex())))
            }
        }
        return
    }

    // Withdraw minipool if logged out and withdrawable
    if *status == minipool.LOGGED_OUT && message.Status.Code == "withdrawable" {
        if txor, err := w.p.AM.GetNodeAccountTransactor(); err != nil {
            log.Println(w.c(err))
        } else {
            txor.GasLimit = 100000 // Gas estimates on this method are incorrect
            if _, err := w.p.CM.Contracts["rocketNodeWatchtower"].Transact(txor, "withdrawMinipool", minipoolAddress, big.NewInt(0)); err != nil {
                log.Println(w.c(errors.New("Error withdrawing minipool: " + err.Error())))
            } else {
                log.Println(w.c(fmt.Sprintf("Minipool %s was withdrawn with a balance of %.2f ETH", minipoolAddress.Hex(), 0.00)))
            }
        }
        return
    }

}

