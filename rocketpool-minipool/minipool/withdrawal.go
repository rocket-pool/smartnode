package minipool

import (
    "context"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "math/big"
    "time"

    "github.com/rocket-pool/smartnode/shared/services"
    beaconchain "github.com/rocket-pool/smartnode/shared/services/beacon-chain"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Config
const CHECK_MINIPOOLS_INTERVAL string = "1m"
var checkMinipoolsInterval, _ = time.ParseDuration(CHECK_MINIPOOLS_INTERVAL)


// Withdrawal process
type WithdrawalProcess struct {
    p *services.Provider
    minipool *Minipool
    minipoolExiting bool
    stop chan struct{}
    done chan struct{}
}


/**
 * Start beacon withdrawal process
 */
func StartWithdrawalProcess(p *services.Provider, minipool *Minipool, done chan struct{}) {

    // Initialise process
    process := &WithdrawalProcess{
        p: p,
        minipool: minipool,
        minipoolExiting: false,
        stop: make(chan struct{}),
        done: done,
    }

    // Start
    process.start()

}


/**
 * Start process
 */
func (p *WithdrawalProcess) start() {

    // Check minipool for withdrawal on interval while checking
    go (func() {
        checkMinipoolsTimer := time.NewTicker(checkMinipoolsInterval)
        checking := true
        for checking {
            select {
                case <-checkMinipoolsTimer.C:
                    p.checkWithdrawal()
                case <-p.stop:
                    checkMinipoolsTimer.Stop()
                    checking = false
            }
        }
    })()

    // Subscribe to beacon chain events
    messageChannel := make(chan interface{})
    p.p.Publisher.AddSubscriber("beacon.client.message", messageChannel)

    // Handle beacon chain events while subscribed
    go (func() {
        subscribed := true
        for subscribed {
            select {
                case eventData := <-messageChannel:
                    event := (eventData).(struct{Client *beaconchain.Client; Message []byte})
                    go p.onBeaconClientMessage(event.Message)
                case <-p.stop:
                    p.p.Publisher.RemoveSubscriber("beacon.client.message", messageChannel)
                    subscribed = false
            }
        }
    })()

    // Block thread until done
    select {
        case <-p.stop:
            p.p.Log.Println(fmt.Sprintf("Ending minipool %s withdrawal process...", p.minipool.Address.Hex()))
            p.done <- struct{}{}
    }

}


/**
 * Check minipool for withdrawal
 */
func (p *WithdrawalProcess) checkWithdrawal() {

    // Wait for node to sync
    eth.WaitSync(p.p.Client, true, false)

    // Check minipool contract still exists
    if code, err := p.p.Client.CodeAt(context.Background(), *(p.minipool.Address), nil); err != nil {
        p.p.Log.Println(errors.New("Error retrieving contract code at minipool address: " + err.Error()))
        return
    } else if len(code) == 0 {
        p.p.Log.Println(fmt.Sprintf("Minipool %s no longer exists...", p.minipool.Address.Hex()))
        close(p.stop)
        return
    }

    // Get latest block header
    header, err := p.p.Client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        p.p.Log.Println(errors.New("Error retrieving latest block header: " + err.Error()))
        return
    }

    // Get minipool status
    status, err := minipool.GetStatus(p.p.CM, p.minipool.Address)
    if err != nil {
        p.p.Log.Println(errors.New("Error retrieving minipool status: " + err.Error()))
        return
    }

    // Log
    p.p.Log.Println(fmt.Sprintf("Checking minipool %s for withdrawal at block %s...", p.minipool.Address.Hex(), header.Number.String()))

    // Check minipool status
    if status.Status > minipool.STAKING {
        p.p.Log.Println(fmt.Sprintf("Minipool %s has already progressed beyond staking...", p.minipool.Address.Hex()))
        close(p.stop)
        return
    } else if status.Status < minipool.STAKING {
        p.p.Log.Println(fmt.Sprintf("Minipool %s is not staking yet...", p.minipool.Address.Hex()))
        return
    }

    // Get minipool exit block
    var exitBlock big.Int
    exitBlock.Add(status.StatusBlock, status.StakingDuration)

    // Check exit block
    if header.Number.Cmp(&exitBlock) == -1 {
        p.p.Log.Println(fmt.Sprintf("Minipool %s is not ready to withdraw until block %s...", p.minipool.Address.Hex(), exitBlock.String()))
        return
    }

    // Check if already marked for exit
    if p.minipoolExiting { return }

    // Mark minipool for exit and log
    p.minipoolExiting = true
    p.p.Log.Println(fmt.Sprintf("Minipool %s is ready to withdraw, since block %s...", p.minipool.Address.Hex(), exitBlock.String()))

    // Request validator status
    if payload, err := json.Marshal(beaconchain.ClientMessage{
        Message: "get_validator_status",
        Pubkey: hex.EncodeToString(status.ValidatorPubkey),
    }); err != nil {
        p.p.Log.Println(errors.New("Error encoding get validator status payload: " + err.Error()))
    } else if err := p.p.Beacon.Send(payload); err != nil {
        p.p.Log.Println(errors.New("Error sending get validator status message: " + err.Error()))
    }

}


/**
 * Handle beacon chain client messages
 */
func (p *WithdrawalProcess) onBeaconClientMessage(messageData []byte) {

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

            // Check validator pubkey and minipool status
            if hex.EncodeToString(p.minipool.Key.PublicKey.Marshal()) != message.Pubkey { break }
            if !p.minipoolExiting { break }

            // Handle statuses
            switch message.Status.Code {

                // Not exited
                case "inactive": fallthrough
                case "active":
                    if message.Status.Initiated.Exit == 0 {

                        // Log status
                        p.p.Log.Println(fmt.Sprintf("Minipool %s has not exited yet, exiting...", p.minipool.Address.Hex()))

                        // Send exit message
                        if payload, err := json.Marshal(beaconchain.ClientMessage{
                            Message: "exit",
                            Pubkey: message.Pubkey,
                        }); err != nil {
                            p.p.Log.Println(errors.New("Error encoding exit payload: " + err.Error()))
                        } else if err := p.p.Beacon.Send(payload); err != nil {
                            p.p.Log.Println(errors.New("Error sending exit message: " + err.Error()))
                        }

                    } else {
                        p.p.Log.Println(fmt.Sprintf("Minipool %s is already exiting...", p.minipool.Address.Hex()))
                    }

                // Exited
                case "exited": fallthrough
                case "withdrawable": fallthrough
                case "withdrawn":
                    p.p.Log.Println(fmt.Sprintf("Minipool %s has exited successfully...", p.minipool.Address.Hex()))
                    close(p.stop)

            }

        // Success response
        case "success":
            if message.Action == "initiate_exit" {
                p.p.Log.Println(fmt.Sprintf("Minipool %s initiated exit successfully...", p.minipool.Address.Hex()))
            }

        // Error
        case "error":
            p.p.Log.Println("A beacon server error occurred: ", message.Error)

    }

}

