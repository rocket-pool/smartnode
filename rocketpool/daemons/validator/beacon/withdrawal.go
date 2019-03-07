package beacon

import (
    "context"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "math/big"
    "time"

    "github.com/fatih/color"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
    beaconchain "github.com/rocket-pool/smartnode-cli/rocketpool/services/beacon-chain"
)


// Config
const WITHDRAWAL_LOG_COLOR = color.FgYellow
const CHECK_MINIPOOLS_INTERVAL string = "15s"
var checkMinipoolsInterval, _ = time.ParseDuration(CHECK_MINIPOOLS_INTERVAL)


// Withdrawal process
type WithdrawalProcess struct {
    c func(a ...interface{}) string
    p *services.Provider
    exitReadyValidators map[string]bool
}


/**
 * Start beacon withdrawal process
 */
func StartWithdrawalProcess(p *services.Provider) {

    // Initialise process
    withdrawalProcess := &WithdrawalProcess{
        c: color.New(WITHDRAWAL_LOG_COLOR).SprintFunc(),
        p: p,
        exitReadyValidators: make(map[string]bool),
    }

    // Start
    withdrawalProcess.start()

}


/**
 * Start process
 */
func (w *WithdrawalProcess) start() {

    // Check staking minipools for withdrawal on interval
    go (func() {
        checkMinipoolsTimer := time.NewTicker(checkMinipoolsInterval)
        for _ = range checkMinipoolsTimer.C {
            w.checkStakingMinipools()
        }
    })()

    // Subscribe to beacon chain events
    messageChannel := make(chan interface{})
    w.p.Publisher.AddSubscriber("beacon.client.message", messageChannel)

    // Handle beacon chain events
    go (func() {
        for eventData := range messageChannel {
            event := (eventData).(struct{Client *beaconchain.Client; Message []byte})
            w.onBeaconClientMessage(event.Message)
        }
    })()

}


/**
 * Check staking minipools for withdrawal
 */
func (w *WithdrawalProcess) checkStakingMinipools() {

    // Get latest block header
    header, err := w.p.Client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        log.Println(w.c(errors.New("Error retrieving latest block header: " + err.Error())))
        return
    }

    // Log
    log.Println(w.c(fmt.Sprintf("Checking staking minipools for withdrawal at block %s...", header.Number.String())))

    // Check minipools
    for _, minipool := range w.p.VM.Validators {

        // Get minipool validator exit block and pubkey
        var exitBlock big.Int
        exitBlock.Add(minipool.StatusBlock, minipool.StakingDuration)
        pubkeyHex := hex.EncodeToString(minipool.ValidatorPubkey)

        // Check exit block
        if header.Number.Cmp(&exitBlock) == -1 { continue }

        // Check if already marked for exit
        if w.exitReadyValidators[pubkeyHex] { continue }

        // Mark validator for exit and log
        w.exitReadyValidators[pubkeyHex] = true
        log.Println(w.c(fmt.Sprintf("Validator %s ready to withdraw, since block %s...", pubkeyHex, exitBlock.String())))

        // Request validator status
        if payload, err := json.Marshal(beaconchain.ClientMessage{
            Message: "get_validator_status",
            Pubkey: pubkeyHex,
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
func (w *WithdrawalProcess) onBeaconClientMessage(messageData []byte) {

    // Parse message
    message := new(beaconchain.ServerMessage)
    if err := json.Unmarshal(messageData, message); err != nil {
        log.Println(w.c(errors.New("Error decoding beacon message: " + err.Error())))
        return
    }

    // Handle message by type
    switch message.Message {

        // Validator status
        case "validator_status":

            // Check validator is ready to exit
            if !w.exitReadyValidators[message.Pubkey] { break }

            // Handle statuses
            switch message.Status.Code {

                // Not exited
                case "inactive": fallthrough
                case "active":
                    if message.Status.Initiated.Exit == 0 {

                        // Log status
                        log.Println(w.c(fmt.Sprintf("Validator %s has not exited yet, exiting...", message.Pubkey)))

                        // Send exit message
                        if payload, err := json.Marshal(beaconchain.ClientMessage{
                            Message: "exit",
                            Pubkey: message.Pubkey,
                        }); err != nil {
                            log.Println(w.c(errors.New("Error encoding exit payload: " + err.Error())))
                        } else if err := w.p.Beacon.Send(payload); err != nil {
                            log.Println(w.c(errors.New("Error sending exit message: " + err.Error())))
                        }

                    } else {
                        log.Println(w.c(fmt.Sprintf("Validator %s is already exiting...", message.Pubkey)))
                    }

                // Exited
                case "exited": fallthrough
                case "withdrawable": fallthrough
                case "withdrawn":
                    log.Println(w.c(fmt.Sprintf("Validator %s has exited successfully...", message.Pubkey)))

            }

        // Success response
        case "success":
            if message.Action == "initiate_exit" {
                log.Println(w.c("Validator initiated exit successfully..."))
            }

        // Error
        case "error":
            log.Println(w.c("A beacon server error occurred:", message.Error))

    }

}

