package beacon

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/gorilla/websocket"
    "github.com/urfave/cli"
)


// Config
const RECONNECT_INTERVAL string = "15s"


// Shared vars
var immediate, _ = time.ParseDuration("0s")
var reconnectInterval, _ = time.ParseDuration(RECONNECT_INTERVAL)
var connectionTimer *time.Timer
var pubkeys []string


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
func StartActivityProcess(c *cli.Context, fatalErrorChannel chan error) {

    // Get node's validator pubkeys
    // :TODO: implement once BLS library is available
    pubkeys = []string{
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd01",
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd02",
    }

    // Initialise beacon connection timer
    connectionTimer = time.NewTimer(immediate)
    for _ = range connectionTimer.C {
        connectToBeacon(c.GlobalString("beacon"))
    }

}


// Connect to beacon server and start validation
func connectToBeacon(providerUrl string) {

    // Open websocket connection to beacon server
    wsConnection, _, err := websocket.DefaultDialer.Dial(providerUrl, nil)
    if err != nil {

        // Log connection errors and retry
        log.Println(errors.New("Error connecting to beacon server: " + err.Error()))
        log.Println(fmt.Sprintf("Retrying in %s...", reconnectInterval.String()))
        connectionTimer.Reset(reconnectInterval)
        return

    }

    // Log success & defer close
    log.Println("Connected to beacon server at", providerUrl)
    defer wsConnection.Close()

    // Validator active statuses
    validatorActive := make(map[string]bool)
    for _, pubkey := range pubkeys {
        validatorActive[strings.ToLower(pubkey)] = false
    }

    // Handle beacon messages
    closed := make(chan struct{})
    go (func() {
        defer close(closed)
        for {
            if message, err, didClose := readMessage(wsConnection); err != nil {
                log.Println(err)
                if didClose { return }
            } else {
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
                                } else if err := wsConnection.WriteMessage(websocket.TextMessage, payload); err != nil {
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
        }
    })()

    // Request validator statuses
    for _, pubkey := range pubkeys {
        if payload, err := json.Marshal(ClientMessage{
            Message: "get_validator_status",
            Pubkey: pubkey,
        }); err != nil {
            log.Println(errors.New("Error encoding get validator status payload: " + err.Error()))
        } else if err := wsConnection.WriteMessage(websocket.TextMessage, payload); err != nil {
            log.Println(errors.New("Error sending get validator status message: " + err.Error()))
        }
    }

    // Block thread until closed, reconnect
    select {
        case <-closed:
            log.Println(fmt.Sprintf("Connection closed, reconnecting in %s...", reconnectInterval.String()))
            connectionTimer.Reset(reconnectInterval)
    }

}


// Read message from beacon
func readMessage(wsConnection *websocket.Conn) (*ServerMessage, error, bool) {

    // Read message
    messageType, messageData, err := wsConnection.ReadMessage()
    if err != nil {
        return nil, errors.New("Error reading beacon message: " + err.Error()), websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err)
    }

    // Check message type
    if messageType != websocket.TextMessage {
        return nil, errors.New("Unrecognised beacon message type"), false
    }

    // Decode message
    message := new(ServerMessage)
    if err := json.Unmarshal(messageData, message); err != nil {
        return nil, errors.New("Error decoding beacon message: " + err.Error()), false
    }

    // Return
    return message, nil, false

}

