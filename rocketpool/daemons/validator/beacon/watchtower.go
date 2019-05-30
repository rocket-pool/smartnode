package beacon

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/fatih/color"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
    beaconchain "github.com/rocket-pool/smartnode-cli/rocketpool/services/beacon-chain"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)

// Config
const WATCHTOWER_LOG_COLOR = color.FgRed
const CHECK_TRUSTED_INTERVAL string = "15s"
const GET_ACTIVE_MINIPOOLS_INTERVAL string = "15s"

var checkTrustedInterval, _ = time.ParseDuration(CHECK_TRUSTED_INTERVAL)
var getActiveMinipoolsInterval, _ = time.ParseDuration(GET_ACTIVE_MINIPOOLS_INTERVAL)

// Watchtower process
type WatchtowerProcess struct {
    c                      func(a ...interface{}) string
    p                      *services.Provider
    updatingMinipools      bool
    getActiveMinipoolsStop chan bool
    beaconMessageChannel   chan interface{}
    activeMinipools        map[string]common.Address
}

/**
 * Start beacon watchtower process
 */
func StartWatchtowerProcess(p *services.Provider) {

    // Initialise process
    process := &WatchtowerProcess{
        c:                      color.New(WATCHTOWER_LOG_COLOR).SprintFunc(),
        p:                      p,
        updatingMinipools:      false,
        getActiveMinipoolsStop: make(chan bool),
        beaconMessageChannel:   make(chan interface{}),
        activeMinipools:        make(map[string]common.Address),
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

    // Handle beacon chain events
    go (func() {
        for eventData := range p.beaconMessageChannel {
            event := (eventData).(struct {
                Client  *beaconchain.Client
                Message []byte
            })
            p.onBeaconClientMessage(event.Message)
        }
    })()

}

/**
 * Check if node is trusted
 */
func (p *WatchtowerProcess) checkTrusted() {

    // Get trusted status
    trusted := new(bool)
    if err := p.p.CM.Contracts["rocketNodeAPI"].Call(nil, trusted, "getTrusted", p.p.AM.GetNodeAccount().Address); err != nil {
        log.Println(p.c(errors.New("Error retrieving node trusted status: " + err.Error())))
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
    if p.updatingMinipools {
        return
    }
    p.updatingMinipools = true

    // Log
    log.Println(p.c("Node is trusted, starting watchtower process..."))

    // Get active minipools on interval
    go (func() {
        p.getActiveMinipools()
        getActiveMinipoolsTimer := time.NewTicker(getActiveMinipoolsInterval)
        for {
            select {
            case <-getActiveMinipoolsTimer.C:
                p.getActiveMinipools()
            case <-p.getActiveMinipoolsStop:
                getActiveMinipoolsTimer.Stop()
                return
            }
        }
    })()

    // Subscribe to beacon chain events
    p.p.Publisher.AddSubscriber("beacon.client.message", p.beaconMessageChannel)

}

/**
 * Stop minipool updates
 */
func (p *WatchtowerProcess) stopUpdateMinipools() {

    // Cancel if not updating minipools
    if !p.updatingMinipools {
        return
    }
    p.updatingMinipools = false

    // Log
    log.Println(p.c("Node is untrusted, stopping watchtower process..."))

    // Stop getting active minipools
    p.getActiveMinipoolsStop <- true

    // Unsubscribe from beacon chain events
    p.p.Publisher.RemoveSubscriber("beacon.client.message", p.beaconMessageChannel)

}

/**
 * Get active minipools by validator pubkey
 */
func (p *WatchtowerProcess) getActiveMinipools() {

    // Get active minipools
    if activeMinipools, err := minipool.GetActiveMinipoolsByValidatorPubkey(p.p.CM); err != nil {
        log.Println(p.c(errors.New("Error getting active minipools: " + err.Error())))
        return
    } else {
        p.activeMinipools = *activeMinipools
    }

    // Request validator statuses for active minipools
    for pubkey, minipoolAddress := range p.activeMinipools {
        go (func(pubkey string, minipoolAddress common.Address) {

            // Check minipool status
            if status, err := getMinipoolStatus(p.p, &minipoolAddress); err != nil {
                log.Println(p.c(err))
                return
            } else if *status < minipool.STAKING || *status > minipool.LOGGED_OUT {
                return
            }

            // Request validator status
            if payload, err := json.Marshal(beaconchain.ClientMessage{
                Message: "get_validator_status",
                Pubkey:  pubkey,
            }); err != nil {
                log.Println(p.c(errors.New("Error encoding get validator status payload: " + err.Error())))
            } else if err := p.p.Beacon.Send(payload); err != nil {
                log.Println(p.c(errors.New("Error sending get validator status message: " + err.Error())))
            }

        })(pubkey, minipoolAddress)
    }

}

/**
 * Handle beacon chain client messages
 */
func (p *WatchtowerProcess) onBeaconClientMessage(messageData []byte) {

    // Parse message
    message := new(beaconchain.ServerMessage)
    if err := json.Unmarshal(messageData, message); err != nil {
        log.Println(p.c(errors.New("Error decoding beacon message: " + err.Error())))
        return
    }

    // Handle exit and withdrawal validator status messages only
    if message.Message != "validator_status" {
        return
    }
    if !(message.Status.Code == "exited" || message.Status.Code == "withdrawable") {
        return
    }

    // Get associated minipool
    minipoolAddress, ok := p.activeMinipools[message.Pubkey]
    if !ok {
        return
    }

    // Get minipool status
    status, err := getMinipoolStatus(p.p, &minipoolAddress)
    if err != nil {
        log.Println(p.c(err))
        return
    }

    // Log minipool out if staking
    if *status == minipool.STAKING {

        // Log
        log.Println(p.c(fmt.Sprintf("Minipool %s is ready for logout...", minipoolAddress.Hex())))

        // Log out
        if txor, err := p.p.AM.GetNodeAccountTransactor(); err != nil {
            log.Println(p.c(err))
        } else {
            txor.GasLimit = 300000 // Gas estimates on this method are incorrect
            if _, err := p.p.CM.Contracts["rocketNodeWatchtower"].Transact(txor, "logoutMinipool", minipoolAddress); err != nil {
                log.Println(p.c(errors.New("Error logging out minipool: " + err.Error())))
            } else {
                log.Println(p.c(fmt.Sprintf("Minipool %s was successfully logged out", minipoolAddress.Hex())))
            }
        }

        return
    }

    // Withdraw minipool if logged out and withdrawable
    if *status == minipool.LOGGED_OUT && message.Status.Code == "withdrawable" {

        // Log
        log.Println(p.c(fmt.Sprintf("Minipool %s is ready for withdrawal...", minipoolAddress.Hex())))

        // Get balance to withdraw
        balanceWei := eth.GweiToWei(float64(message.Balance))

        // Withdraw
        if txor, err := p.p.AM.GetNodeAccountTransactor(); err != nil {
            log.Println(p.c(err))
        } else {
            txor.GasLimit = 300000 // Gas estimates on this method are incorrect
            if _, err := p.p.CM.Contracts["rocketNodeWatchtower"].Transact(txor, "withdrawMinipool", minipoolAddress, balanceWei); err != nil {
                log.Println(p.c(errors.New("Error withdrawing minipool: " + err.Error())))
            } else {
                log.Println(p.c(fmt.Sprintf("Minipool %s was successfully withdrawn with a balance of %.2f ETH", minipoolAddress.Hex(), eth.WeiToEth(balanceWei))))
            }
        }

        return
    }

}

// Get a minipool's status
func getMinipoolStatus(p *services.Provider, address *common.Address) (*uint8, error) {

    // Initialise minipool contract
    minipoolContract, err := p.CM.NewContract(address, "rocketMinipool")
    if err != nil {
        return nil, errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Get minipool's current status
    status := new(uint8)
    if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
        return nil, errors.New("Error retrieving minipool status: " + err.Error())
    }

    // Return
    return status, nil

}
