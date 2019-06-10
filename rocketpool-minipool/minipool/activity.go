package minipool

import (
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "log"

    "github.com/fatih/color"

    "github.com/rocket-pool/smartnode-cli/shared/services"
    beaconchain "github.com/rocket-pool/smartnode-cli/shared/services/beacon-chain"
)


// Config
const ACTIVITY_LOG_COLOR = color.FgBlue


// Activity process
type ActivityProcess struct {
    c func(a ...interface{}) string
    p *services.Provider
    activeValidators map[string]bool
}


/**
 * Start beacon activity process
 */
func StartActivityProcess(p *services.Provider) {

    // Initialise process
    process := &ActivityProcess{
        c: color.New(ACTIVITY_LOG_COLOR).SprintFunc(),
        p: p,
        activeValidators: make(map[string]bool),
    }

    // Start
    process.start()

}


/**
 * Start process
 */
func (p *ActivityProcess) start() {

    // Subscribe to beacon chain events
    connectedChannel := make(chan interface{})
    messageChannel := make(chan interface{})
    p.p.Publisher.AddSubscriber("beacon.client.connected", connectedChannel)
    p.p.Publisher.AddSubscriber("beacon.client.message", messageChannel)

    // Handle beacon chain events
    go (func() {
        for {
            select {
                case <-connectedChannel:
                    p.onBeaconClientConnected()
                case eventData := <-messageChannel:
                    event := (eventData).(struct{Client *beaconchain.Client; Message []byte})
                    p.onBeaconClientMessage(event.Message)
            }
        }
    })()

}


/**
 * Handle beacon chain client connections
 */
func (p *ActivityProcess) onBeaconClientConnected() {

    // Request validator statuses
    for _, validator := range p.p.VM.Validators {
        if payload, err := json.Marshal(beaconchain.ClientMessage{
            Message: "get_validator_status",
            Pubkey: hex.EncodeToString(validator.ValidatorPubkey),
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
func (p *ActivityProcess) onBeaconClientMessage(messageData []byte) {

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

            // Check validator pubkey
            found := false
            for _, validator := range p.p.VM.Validators {
                if hex.EncodeToString(validator.ValidatorPubkey) == message.Pubkey {
                    found = true
                    break
                }
            }
            if !found { break }

            // Handle statuses
            switch message.Status.Code {

                // Inactive
                case "inactive":
                    log.Println(p.c(fmt.Sprintf("Validator %s is inactive, waiting until active to send activity...", message.Pubkey)))
                    delete(p.activeValidators, message.Pubkey)

                // Active
                case "active":
                    log.Println(p.c(fmt.Sprintf("Validator %s is active, sending activity...", message.Pubkey)))
                    p.activeValidators[message.Pubkey] = true

                // Exited
                case "exited": fallthrough
                case "withdrawable": fallthrough
                case "withdrawn":
                    log.Println(p.c(fmt.Sprintf("Validator %s has exited, not sending activity...", message.Pubkey)))
                    delete(p.activeValidators, message.Pubkey)

            }

        // Epoch
        case "epoch":

            // Send activity for active validators
            for _, validator := range p.p.VM.Validators {
                pubkeyHex := hex.EncodeToString(validator.ValidatorPubkey)
                if p.activeValidators[pubkeyHex] {
                    log.Println(p.c(fmt.Sprintf("New epoch, sending activity for validator %s...", pubkeyHex)))

                    // Send activity
                    if payload, err := json.Marshal(beaconchain.ClientMessage{
                        Message: "activity",
                        Pubkey: pubkeyHex,
                    }); err != nil {
                        log.Println(p.c(errors.New("Error encoding activity payload: " + err.Error())))
                    } else if err := p.p.Beacon.Send(payload); err != nil {
                        log.Println(p.c(errors.New("Error sending activity message: " + err.Error())))
                    }

                }
            }

        // Success response
        case "success":
            if message.Action == "process_activity" {
                log.Println(p.c("Processed validator activity successfully..."))
            }

        // Error
        case "error":
            log.Println(p.c("A beacon server error occurred: ", message.Error))

    }

}

