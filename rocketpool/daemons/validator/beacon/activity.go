package beacon

import (
    "encoding/json"
    "errors"
    "fmt"
    "strings"

    "github.com/gorilla/websocket"
    "github.com/urfave/cli"
)


// Start beacon activity process
func StartActivityProcess(c *cli.Context, errorChannel chan error, fatalErrorChannel chan error) {

    // Get node's validator pubkeys
    // :TODO: implement once BLS library is available
    pubkeys := []string{
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd01",
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd02",
    }

    // Open websocket connection to beacon chain
    wsConnection, _, err := websocket.DefaultDialer.Dial(c.GlobalString("beacon"), nil)
    if err != nil {
        fatalErrorChannel <- errors.New("Error connecting to beacon chain: " + err.Error())
        return
    }
    defer wsConnection.Close()

    // Validator active statuses
    validatorActive := make(map[string]bool)
    for _, pubkey := range pubkeys {
        validatorActive[strings.ToLower(pubkey)] = false
    }

    // Handle server messages
    go (func() {
        for {
            if messageType, messageData, err := wsConnection.ReadMessage(); err != nil {
                errorChannel <- errors.New("Error reading beacon chain message: " + err.Error())
            } else {

                // Check message type
                if messageType != websocket.TextMessage {
                    errorChannel <- errors.New("Unrecognised beacon chain message type")
                } else {

                    // Decode message
                    message := struct{
                        Message string  `json:"message"`
                        Pubkey string   `json:"pubkey"`
                        Status struct {
                            Code string     `json:"code"`
                        }               `json:"status"`
                        Action string   `json:"action"`
                        Error string    `json:"error"`
                    }{}
                    if err := json.Unmarshal(messageData, &message); err != nil {
                        errorChannel <- errors.New("Error decoding beacon chain message: " + err.Error())
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
                                        fmt.Println("Validator is inactive, waiting until active...")
                                        validatorActive[strings.ToLower(message.Pubkey)] = false

                                    // Active
                                    case "active":
                                        fmt.Println("Validator is active, sending activity...")
                                        validatorActive[strings.ToLower(message.Pubkey)] = true

                                    // Exited
                                    case "exited": fallthrough
                                    case "withdrawable": fallthrough
                                    case "withdrawn":
                                        fmt.Println("Validator has exited, closing connection")
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
                                        if payload, err := json.Marshal(struct{
                                            Message string  `json:"message"`
                                            Pubkey string   `json:"pubkey"`
                                        }{
                                            "activity",
                                            pubkey,
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
                                fmt.Println("A server error occurred:", message.Error)

                        }
                    }

                }

            }
        }
    })()

    // Request validator statuses
    for _, pubkey := range pubkeys {
        if payload, err := json.Marshal(struct{
            Message string  `json:"message"`
            Pubkey string   `json:"pubkey"`
        }{
            "get_validator_status",
            pubkey,
        }); err != nil {
            errorChannel <- errors.New("Error encoding get validator status payload: " + err.Error())
        } else if err := wsConnection.WriteMessage(websocket.TextMessage, payload); err != nil {
            errorChannel <- errors.New("Error sending get validator status message: " + err.Error())
        }
    }
    
    // Block thread
    select {}

}

