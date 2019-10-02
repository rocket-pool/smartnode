package minipool

import (
    "encoding/json"
    "errors"
    "fmt"

    "github.com/rocket-pool/smartnode/shared/services"
    beaconchain "github.com/rocket-pool/smartnode/shared/services/beacon-chain"
)


// Activity process
type ActivityProcess struct {
    p *services.Provider
    minipool *Minipool
    validatorActive bool
    stop chan struct{}
    done chan struct{}
}


/**
 * Start beacon activity process
 */
func StartActivityProcess(p *services.Provider, minipool *Minipool, done chan struct{}) {

    // Initialise process
    process := &ActivityProcess{
        p: p,
        minipool: minipool,
        validatorActive: false,
        stop: make(chan struct{}),
        done: done,
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

    // Handle beacon chain events while subscribed
    go (func() {
        subscribed := true
        for subscribed {
            select {
                case <-connectedChannel:
                    go p.onBeaconClientConnected()
                case eventData := <-messageChannel:
                    event := (eventData).(struct{Client *beaconchain.Client; Message []byte})
                    go p.onBeaconClientMessage(event.Message)
                case <-p.stop:
                    p.p.Publisher.RemoveSubscriber("beacon.client.connected", connectedChannel)
                    p.p.Publisher.RemoveSubscriber("beacon.client.message", messageChannel)
                    subscribed = false
            }
        }
    })()

    // Block thread until done
    select {
        case <-p.stop:
            p.p.Log.Println(fmt.Sprintf("Ending minipool %s activity process...", p.minipool.Address.Hex()))
            p.done <- struct{}{}
    }

}


/**
 * Handle beacon chain client connections
 */
func (p *ActivityProcess) onBeaconClientConnected() {

    // Request validator status
    if payload, err := json.Marshal(beaconchain.ClientMessage{
        Message: "get_validator_status",
        Pubkey: p.minipool.Pubkey,
    }); err != nil {
        p.p.Log.Println(errors.New("Error encoding get validator status payload: " + err.Error()))
    } else if err := p.p.Beacon.Send(payload); err != nil {
        p.p.Log.Println(errors.New("Error sending get validator status message: " + err.Error()))
    }

}


/**
 * Handle beacon chain client messages
 */
func (p *ActivityProcess) onBeaconClientMessage(messageData []byte) {

    // Parse message
    message := new(beaconchain.ServerMessage)
    if err := json.Unmarshal(messageData, message); err != nil {
        p.p.Log.Println(errors.New("Error decoding beacon message: " + err.Error()))
        return
    }

    // Handle message by type
    switch message.Message {

        // Validator status
        case "validator_status":

            // Check validator pubkey
            if p.minipool.Pubkey != message.Pubkey { break }

            // Handle statuses
            switch message.Status.Code {

                // Inactive
                case "inactive":
                    p.p.Log.Println(fmt.Sprintf("Validator %s is inactive, waiting until active to send activity...", message.Pubkey))
                    p.validatorActive = false

                // Active
                case "active":
                    p.p.Log.Println(fmt.Sprintf("Validator %s is active, sending activity...", message.Pubkey))
                    p.validatorActive = true

                // Exited
                case "exited": fallthrough
                case "withdrawable": fallthrough
                case "withdrawn":
                    p.p.Log.Println(fmt.Sprintf("Validator %s has exited, not sending activity...", message.Pubkey))
                    p.validatorActive = false
                    close(p.stop)

            }

        // Epoch
        case "epoch":

            // Check validator active status, get pubkey string
            if !p.validatorActive { break }
            pubkeyHex := p.minipool.Pubkey

            // Log activity
            p.p.Log.Println(fmt.Sprintf("New epoch, sending activity for validator %s...", pubkeyHex))

            // Send activity
            if payload, err := json.Marshal(beaconchain.ClientMessage{
                Message: "activity",
                Pubkey: pubkeyHex,
            }); err != nil {
                p.p.Log.Println(errors.New("Error encoding activity payload: " + err.Error()))
            } else if err := p.p.Beacon.Send(payload); err != nil {
                p.p.Log.Println(errors.New("Error sending activity message: " + err.Error()))
            }

        // Success response
        case "success":
            if message.Action == "process_activity" {
                p.p.Log.Println("Processed validator activity successfully...")
            }

        // Error
        case "error":
            p.p.Log.Println("A beacon server error occurred: ", message.Error)

    }

}

