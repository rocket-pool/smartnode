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
    process := &WithdrawalProcess{
        c: color.New(WITHDRAWAL_LOG_COLOR).SprintFunc(),
        p: p,
        exitReadyValidators: make(map[string]bool),
    }

    // Start
    process.start()

}


/**
 * Start process
 */
func (p *WithdrawalProcess) start() {

    // Check staking minipools for withdrawal on interval
    go (func() {
        checkMinipoolsTimer := time.NewTicker(checkMinipoolsInterval)
        for _ = range checkMinipoolsTimer.C {
            p.checkStakingMinipools()
        }
    })()

    // Subscribe to beacon chain events
    messageChannel := make(chan interface{})
    p.p.Publisher.AddSubscriber("beacon.client.message", messageChannel)

    // Handle beacon chain events
    go (func() {
        for eventData := range messageChannel {
            event := (eventData).(struct{Client *beaconchain.Client; Message []byte})
            p.onBeaconClientMessage(event.Message)
        }
    })()

}


/**
 * Check staking minipools for withdrawal
 */
func (p *WithdrawalProcess) checkStakingMinipools() {

    // Get latest block header
    header, err := p.p.Client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        log.Println(p.c(errors.New("Error retrieving latest block header: " + err.Error())))
        return
    }

    // Log
    log.Println(p.c(fmt.Sprintf("Checking staking minipools for withdrawal at block %s...", header.Number.String())))

    // Check minipools
    for _, minipool := range p.p.VM.Validators {

        // Get minipool validator exit block and pubkey
        var exitBlock big.Int
        exitBlock.Add(minipool.StatusBlock, minipool.StakingDuration)
        pubkeyHex := hex.EncodeToString(minipool.ValidatorPubkey)

        // Check exit block
        if header.Number.Cmp(&exitBlock) == -1 {
            log.Println(p.c(fmt.Sprintf("Validator %s not ready to withdraw until block %s...", pubkeyHex, exitBlock.String())))
            continue
        }

        // Check if already marked for exit
        if p.exitReadyValidators[pubkeyHex] { continue }

        // Mark validator for exit and log
        p.exitReadyValidators[pubkeyHex] = true
        log.Println(p.c(fmt.Sprintf("Validator %s ready to withdraw, since block %s...", pubkeyHex, exitBlock.String())))

        // Request validator status
        if payload, err := json.Marshal(beaconchain.ClientMessage{
            Message: "get_validator_status",
            Pubkey: pubkeyHex,
        }); err != nil {
            log.Println(p.c(errors.New("Error encoding get validator status payload: " + err.Error())))
        } else if err := p.p.Beacon.Send(payload); err != nil {
            log.Println(p.c(errors.New("Error sending get validator status message: " + err.Error())))
        }

    }

}


/**
 * Handle beacon chain client messages
 */
func (p *WithdrawalProcess) onBeaconClientMessage(messageData []byte) {

    // Parse message
    message := new(beaconchain.ServerMessage)
    if err := json.Unmarshal(messageData, message); err != nil {
        log.Println(p.c(errors.New("Error decoding beacon message: " + err.Error())))
        return
    }

    // Handle message by type
    switch message.Message {

        // Validator status
        case "validator_status":

            // Check validator is ready to exit
            if !p.exitReadyValidators[message.Pubkey] { break }

            // Handle statuses
            switch message.Status.Code {

                // Not exited
                case "inactive": fallthrough
                case "active":
                    if message.Status.Initiated.Exit == 0 {

                        // Log status
                        log.Println(p.c(fmt.Sprintf("Validator %s has not exited yet, exiting...", message.Pubkey)))

                        // Send exit message
                        if payload, err := json.Marshal(beaconchain.ClientMessage{
                            Message: "exit",
                            Pubkey: message.Pubkey,
                        }); err != nil {
                            log.Println(p.c(errors.New("Error encoding exit payload: " + err.Error())))
                        } else if err := p.p.Beacon.Send(payload); err != nil {
                            log.Println(p.c(errors.New("Error sending exit message: " + err.Error())))
                        }

                    } else {
                        log.Println(p.c(fmt.Sprintf("Validator %s is already exiting...", message.Pubkey)))
                    }

                // Exited
                case "exited": fallthrough
                case "withdrawable": fallthrough
                case "withdrawn":
                    log.Println(p.c(fmt.Sprintf("Validator %s has exited successfully...", message.Pubkey)))

            }

        // Success response
        case "success":
            if message.Action == "initiate_exit" {
                log.Println(p.c("Validator initiated exit successfully..."))
            }

        // Error
        case "error":
            log.Println(p.c("A beacon server error occurred: ", message.Error))

    }

}

