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

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
    beaconchain "github.com/rocket-pool/smartnode-cli/rocketpool/services/beacon-chain"
)


// Config
const CHECK_MINIPOOLS_INTERVAL string = "15s"
var checkMinipoolsInterval, _ = time.ParseDuration(CHECK_MINIPOOLS_INTERVAL)


// Start beacon withdrawal process
func StartWithdrawalProcess(p *services.Provider) {

    // Set of validators ready to exit
    exitReadyValidators := make(map[string]bool)

    // Check staking minipools for withdrawal on interval
    go (func() {
        checkMinipoolsTimer := time.NewTicker(checkMinipoolsInterval)
        for _ = range checkMinipoolsTimer.C {
            checkStakingMinipools(p, &exitReadyValidators)
        }
    })()

    // Subscribe to beacon chain events
    messageChannel := make(chan interface{})
    p.Publisher.AddSubscriber("beacon.client.message", messageChannel)

    // Handle beacon chain events
    go (func() {
        for {
            select {
                case eventData := <-messageChannel:
                    event := (eventData).(struct{Client *beaconchain.Client; Message []byte})
                    withdrawalHandleBeaconClientMessage(p, &exitReadyValidators, event.Message)
            }
        }
    })()

}


// Check staking minipools for withdrawal
func checkStakingMinipools(p *services.Provider, exitReadyValidators *map[string]bool) {

    // Get latest block header
    header, err := p.Client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        log.Println(errors.New("Error retrieving latest block header: " + err.Error()))
        return
    }

    // Log
    log.Println(fmt.Sprintf("Checking staking minipools for withdrawal at block %s...", header.Number.String()))

    // Check minipools
    for _, minipool := range p.VM.Validators {

        // Get minipool validator exit block and pubkey
        var exitBlock big.Int
        exitBlock.Add(minipool.StatusBlock, minipool.StakingDuration)
        pubkeyHex := hex.EncodeToString(minipool.ValidatorPubkey)

        // Check exit block
        if header.Number.Cmp(&exitBlock) == -1 { continue }

        // Check if already marked for exit
        if (*exitReadyValidators)[pubkeyHex] { continue }

        // Mark validator for exit and log
        (*exitReadyValidators)[pubkeyHex] = true
        log.Println(fmt.Sprintf("Validator %s ready to withdraw, since block %s...", pubkeyHex, exitBlock.String()))

        // Request validator status
        if payload, err := json.Marshal(beaconchain.ClientMessage{
            Message: "get_validator_status",
            Pubkey: pubkeyHex,
        }); err != nil {
            log.Println(errors.New("Error encoding get validator status payload: " + err.Error()))
        } else if err := p.Beacon.Send(payload); err != nil {
            log.Println(errors.New("Error sending get validator status message: " + err.Error()))
        }

    }

}


// Handle beacon chain client messages
func withdrawalHandleBeaconClientMessage(p *services.Provider, exitReadyValidators *map[string]bool, messageData []byte) {

    // Parse message
    message := new(beaconchain.ServerMessage)
    if err := json.Unmarshal(messageData, message); err != nil {
        log.Println(errors.New("Error decoding beacon message: " + err.Error()))
        return
    }

    // Handle message by type
    switch message.Message {

        // Validator status
        case "validator_status":

            // Check validator is ready to exit
            if !(*exitReadyValidators)[message.Pubkey] { break }

            // Handle statuses
            switch message.Status.Code {

                // Not exited
                case "inactive": fallthrough
                case "active":
                    if message.Status.Initiated.Exit == 0 {

                        // Log status
                        log.Println(fmt.Sprintf("Validator %s has not exited yet, exiting...", message.Pubkey))

                        // Send exit message
                        if payload, err := json.Marshal(beaconchain.ClientMessage{
                            Message: "exit",
                            Pubkey: message.Pubkey,
                        }); err != nil {
                            log.Println(errors.New("Error encoding exit payload: " + err.Error()))
                        } else if err := p.Beacon.Send(payload); err != nil {
                            log.Println(errors.New("Error sending exit message: " + err.Error()))
                        }

                    } else {
                        log.Println(fmt.Sprintf("Validator %s is already exiting...", message.Pubkey))
                    }

                // Exited
                case "exited": fallthrough
                case "withdrawable": fallthrough
                case "withdrawn":
                    log.Println(fmt.Sprintf("Validator %s has exited successfully...", message.Pubkey))

            }

        // Success response
        case "success":
            if message.Action == "initiate_exit" {
                log.Println("Validator initiated exit successfully...")
            }

        // Error
        case "error":
            log.Println("A beacon server error occurred:", message.Error)

    }

}

