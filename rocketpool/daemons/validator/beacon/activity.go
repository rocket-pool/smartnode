package beacon

import (
    "encoding/json"
    "errors"
    "fmt"
    "strings"

    "github.com/gorilla/websocket"
    "github.com/urfave/cli"
)


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
func StartActivityProcess(c *cli.Context, errorChannel chan error, fatalErrorChannel chan error) {

    // Get node's validator pubkeys
    // :TODO: implement once BLS library is available
    pubkeys := []string{
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd01",
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd02",
    }

    // Open websocket connection to beacon server
    wsConnection, _, err := websocket.DefaultDialer.Dial(c.GlobalString("beacon"), nil)
    if err != nil {
        fatalErrorChannel <- errors.New("Error connecting to beacon server: " + err.Error())
        return
    }
    defer wsConnection.Close()

    // Validator active statuses
    validatorActive := make(map[string]bool)
    for _, pubkey := range pubkeys {
        validatorActive[strings.ToLower(pubkey)] = false
    }

    // Handle beacon messages
    go (func() {
        for {
            if message, err := readMessage(wsConnection); err != nil {
                errorChannel <- err
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
                                fmt.Println(fmt.Sprintf("Validator %s is inactive, waiting until active...", message.Pubkey))
                                validatorActive[strings.ToLower(message.Pubkey)] = false

                            // Active
                            case "active":
                                fmt.Println(fmt.Sprintf("Validator %s is active, sending activity...", message.Pubkey))
                                validatorActive[strings.ToLower(message.Pubkey)] = true

                            // Exited
                            case "exited": fallthrough
                            case "withdrawable": fallthrough
                            case "withdrawn":
                                fmt.Println(fmt.Sprintf("Validator %s has exited, closing connection", message.Pubkey))
                                validatorActive[strings.ToLower(message.Pubkey)] = false
                                wsConnection.Close()

                        }

                    // Epoch
                    case "epoch":

                        // Send activity for active validators
                        for _, pubkey := range pubkeys {
                            if validatorActive[strings.ToLower(pubkey)] {
                                fmt.Println(fmt.Sprintf("New epoch, sending activity for validator %s...", pubkey))

                                // Send activity
                                if payload, err := json.Marshal(ClientMessage{
                                    Message: "activity",
                                    Pubkey: pubkey,
                                }); err != nil {
                                    errorChannel <- errors.New("Error encoding activity payload: " + err.Error())
                                } else if err := wsConnection.WriteMessage(websocket.TextMessage, payload); err != nil {
                                    errorChannel <- errors.New("Error sending activity message: " + err.Error())
                                }

                            }
                        }

                    // Success response
                    case "success":
                        if message.Action == "process_activity" {
                            fmt.Println("Processed validator activity successfully...")
                        }

                    // Error
                    case "error":
                        fmt.Println("A beacon server error occurred:", message.Error)

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
            errorChannel <- errors.New("Error encoding get validator status payload: " + err.Error())
        } else if err := wsConnection.WriteMessage(websocket.TextMessage, payload); err != nil {
            errorChannel <- errors.New("Error sending get validator status message: " + err.Error())
        }
    }
    
    // Block thread
    select {}

}


// Read message from beacon
func readMessage(wsConnection *websocket.Conn) (*ServerMessage, error) {

    // Read message
    messageType, messageData, err := wsConnection.ReadMessage()
    if err != nil {
        return nil, errors.New("Error reading beacon message: " + err.Error())
    }

    // Check message type
    if messageType != websocket.TextMessage {
        return nil, errors.New("Unrecognised beacon message type")
    }

    // Decode message
    message := new(ServerMessage)
    if err := json.Unmarshal(messageData, message); err != nil {
        return nil, errors.New("Error decoding beacon message: " + err.Error())
    }

    // Return
    return message, nil

}

