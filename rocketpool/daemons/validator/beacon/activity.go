package beacon

import (
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "log"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
    beaconchain "github.com/rocket-pool/smartnode-cli/rocketpool/services/beacon-chain"
)


// Start beacon activity process
func StartActivityProcess(p *services.Provider) {

    // Active validator set
    activeValidators := make(map[string]bool)

    // Subscribe to beacon chain events
    connectedChannel := make(chan interface{})
    messageChannel := make(chan interface{})
    p.Publisher.AddSubscriber("beacon.client.connected", connectedChannel)
    p.Publisher.AddSubscriber("beacon.client.message", messageChannel)

    // Handle beacon chain events
    go (func() {
        for {
            select {
                case <-connectedChannel:
                    activityHandleBeaconClientConnected(p)
                case eventData := <-messageChannel:
                    event := (eventData).(struct{Client *beaconchain.Client; Message []byte})
                    activityHandleBeaconClientMessage(p, &activeValidators, event.Message)
            }
        }
    })()

}


// Handle beacon chain client connections
func activityHandleBeaconClientConnected(p *services.Provider) {

    // Request validator statuses
    for _, validator := range p.VM.Validators {
        if payload, err := json.Marshal(beaconchain.ClientMessage{
            Message: "get_validator_status",
            Pubkey: hex.EncodeToString(validator.ValidatorPubkey),
        }); err != nil {
            log.Println(errors.New("Error encoding get validator status payload: " + err.Error()))
        } else if err := p.Beacon.Send(payload); err != nil {
            log.Println(errors.New("Error sending get validator status message: " + err.Error()))
        }
    }

}


// Handle beacon chain client messages
func activityHandleBeaconClientMessage(p *services.Provider, activeValidators *map[string]bool, messageData []byte) {

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

            // Check validator pubkey
            found := false
            for _, validator := range p.VM.Validators {
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
                    log.Println(fmt.Sprintf("Validator %s is inactive, waiting until active...", message.Pubkey))
                    delete(*activeValidators, message.Pubkey)

                // Active
                case "active":
                    log.Println(fmt.Sprintf("Validator %s is active, sending activity...", message.Pubkey))
                    (*activeValidators)[message.Pubkey] = true

                // Exited
                case "exited": fallthrough
                case "withdrawable": fallthrough
                case "withdrawn":
                    log.Println(fmt.Sprintf("Validator %s has exited...", message.Pubkey))
                    delete(*activeValidators, message.Pubkey)

            }

        // Epoch
        case "epoch":

            // Send activity for active validators
            for _, validator := range p.VM.Validators {
                pubkeyHex := hex.EncodeToString(validator.ValidatorPubkey)
                if (*activeValidators)[pubkeyHex] {
                    log.Println(fmt.Sprintf("New epoch, sending activity for validator %s...", pubkeyHex))

                    // Send activity
                    if payload, err := json.Marshal(beaconchain.ClientMessage{
                        Message: "activity",
                        Pubkey: pubkeyHex,
                    }); err != nil {
                        log.Println(errors.New("Error encoding activity payload: " + err.Error()))
                    } else if err := p.Beacon.Send(payload); err != nil {
                        log.Println(errors.New("Error sending activity message: " + err.Error()))
                    }

                }
            }

        // Success response
        case "success":
            if message.Action == "process_activity" {
                log.Println("Processed validator activity successfully...")
            }

        // Error
        case "error":
            log.Println("A beacon server error occurred:", message.Error)

    }

}

