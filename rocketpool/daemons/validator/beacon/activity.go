package beacon

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "strings"

    beaconchain "github.com/rocket-pool/smartnode-cli/rocketpool/services/beacon-chain"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/messaging"
)


// Shared vars
var pubkeys []string
var validatorActive map[string]bool


// Client message to server
type ClientMessage struct {
    Message string  `json:"message"`
    Pubkey string   `json:"pubkey"`
}


// Server message to client
type ServerMessage struct {
    Message string  `json:"message"`
    Pubkey string   `json:"pubkey"`
    Status struct {
        Code string `json:"code"`
    }               `json:"status"`
    Action string   `json:"action"`
    Error string    `json:"error"`
}


// Start beacon activity process
func StartActivityProcess(publisher *messaging.Publisher, fatalErrorChannel chan error) {

    // Get node's validator pubkeys
    // :TODO: implement once BLS library is available
    pubkeys = []string{
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd01",
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd02",
    }

    // Validator active statuses
    validatorActive = make(map[string]bool)
    for _, pubkey := range pubkeys {
        validatorActive[strings.ToLower(pubkey)] = false
    }

    // Subscribe to events
    connectedChannel := make(chan interface{})
    messageChannel := make(chan interface{})
    publisher.AddSubscriber("beacon.client.connected", connectedChannel)
    publisher.AddSubscriber("beacon.client.message", messageChannel)

    // Handle events
    for {
        select {
            case eventData := <-connectedChannel:
                client := (eventData).(*beaconchain.Client)
                onBeaconClientConnected(client)
            case eventData := <-messageChannel:
                event := (eventData).(struct{Client *beaconchain.Client; Message []byte})
                onBeaconClientMessage(event.Client, event.Message)
        }
    }

}


// Handle beacon chain client connections
func onBeaconClientConnected(client *beaconchain.Client) {

    // Request validator statuses
    for _, pubkey := range pubkeys {
        if payload, err := json.Marshal(ClientMessage{
            Message: "get_validator_status",
            Pubkey: pubkey,
        }); err != nil {
            log.Println(errors.New("Error encoding get validator status payload: " + err.Error()))
        } else if err := client.Send(payload); err != nil {
            log.Println(errors.New("Error sending get validator status message: " + err.Error()))
        }
    }

}


// Handle beacon chain client messages
func onBeaconClientMessage(client *beaconchain.Client, messageData []byte) {

    // Parse message
    message := new(ServerMessage)
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
            for _, pubkey := range pubkeys {
                if strings.ToLower(pubkey) == strings.ToLower(message.Pubkey) {
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
                    validatorActive[strings.ToLower(message.Pubkey)] = false

                // Active
                case "active":
                    log.Println(fmt.Sprintf("Validator %s is active, sending activity...", message.Pubkey))
                    validatorActive[strings.ToLower(message.Pubkey)] = true

                // Exited
                case "exited": fallthrough
                case "withdrawable": fallthrough
                case "withdrawn":
                    log.Println(fmt.Sprintf("Validator %s has exited...", message.Pubkey))
                    validatorActive[strings.ToLower(message.Pubkey)] = false

            }

        // Epoch
        case "epoch":

            // Send activity for active validators
            for _, pubkey := range pubkeys {
                if validatorActive[strings.ToLower(pubkey)] {
                    log.Println(fmt.Sprintf("New epoch, sending activity for validator %s...", pubkey))

                    // Send activity
                    if payload, err := json.Marshal(ClientMessage{
                        Message: "activity",
                        Pubkey: pubkey,
                    }); err != nil {
                        log.Println(errors.New("Error encoding activity payload: " + err.Error()))
                    } else if err := client.Send(payload); err != nil {
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

