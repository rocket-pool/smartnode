package beacon

import (
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "log"

    "github.com/fatih/color"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
    beaconchain "github.com/rocket-pool/smartnode-cli/rocketpool/services/beacon-chain"
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
    activityProcess := &ActivityProcess{
        c: color.New(ACTIVITY_LOG_COLOR).SprintFunc(),
        p: p,
        activeValidators: make(map[string]bool),
    }

    // Start
    activityProcess.start()

}


/**
 * Start process
 */
func (a *ActivityProcess) start() {

    // Subscribe to beacon chain events
    connectedChannel := make(chan interface{})
    messageChannel := make(chan interface{})
    a.p.Publisher.AddSubscriber("beacon.client.connected", connectedChannel)
    a.p.Publisher.AddSubscriber("beacon.client.message", messageChannel)

    // Handle beacon chain events
    go (func() {
        for {
            select {
                case <-connectedChannel:
                    a.onBeaconClientConnected()
                case eventData := <-messageChannel:
                    event := (eventData).(struct{Client *beaconchain.Client; Message []byte})
                    a.onBeaconClientMessage(event.Message)
            }
        }
    })()

}


/**
 * Handle beacon chain client connections
 */
func (a *ActivityProcess) onBeaconClientConnected() {

    // Request validator statuses
    for _, validator := range a.p.VM.Validators {
        if payload, err := json.Marshal(beaconchain.ClientMessage{
            Message: "get_validator_status",
            Pubkey: hex.EncodeToString(validator.ValidatorPubkey),
        }); err != nil {
            log.Println(a.c(errors.New("Error encoding get validator status payload: " + err.Error())))
        } else if err := a.p.Beacon.Send(payload); err != nil {
            log.Println(a.c(errors.New("Error sending get validator status message: " + err.Error())))
        }
    }

}


/**
 * Handle beacon chain client messages
 */
func (a *ActivityProcess) onBeaconClientMessage(messageData []byte) {

    // Parse message
    message := new(beaconchain.ServerMessage)
    if err := json.Unmarshal(messageData, message); err != nil {
        log.Println(a.c(errors.New("Error decoding beacon message: " + err.Error())))
        return
    }

    // Handle message by type
    switch message.Message {

        // Validator status
        case "validator_status":

            // Check validator pubkey
            found := false
            for _, validator := range a.p.VM.Validators {
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
                    log.Println(a.c(fmt.Sprintf("Validator %s is inactive, waiting until active to send activity...", message.Pubkey)))
                    delete(a.activeValidators, message.Pubkey)

                // Active
                case "active":
                    log.Println(a.c(fmt.Sprintf("Validator %s is active, sending activity...", message.Pubkey)))
                    a.activeValidators[message.Pubkey] = true

                // Exited
                case "exited": fallthrough
                case "withdrawable": fallthrough
                case "withdrawn":
                    log.Println(a.c(fmt.Sprintf("Validator %s has exited, not sending activity...", message.Pubkey)))
                    delete(a.activeValidators, message.Pubkey)

            }

        // Epoch
        case "epoch":

            // Send activity for active validators
            for _, validator := range a.p.VM.Validators {
                pubkeyHex := hex.EncodeToString(validator.ValidatorPubkey)
                if a.activeValidators[pubkeyHex] {
                    log.Println(a.c(fmt.Sprintf("New epoch, sending activity for validator %s...", pubkeyHex)))

                    // Send activity
                    if payload, err := json.Marshal(beaconchain.ClientMessage{
                        Message: "activity",
                        Pubkey: pubkeyHex,
                    }); err != nil {
                        log.Println(a.c(errors.New("Error encoding activity payload: " + err.Error())))
                    } else if err := a.p.Beacon.Send(payload); err != nil {
                        log.Println(a.c(errors.New("Error sending activity message: " + err.Error())))
                    }

                }
            }

        // Success response
        case "success":
            if message.Action == "process_activity" {
                log.Println(a.c("Processed validator activity successfully..."))
            }

        // Error
        case "error":
            log.Println(a.c("A beacon server error occurred: ", message.Error))

    }

}

