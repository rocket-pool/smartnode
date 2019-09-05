package deposit

import (
    "bytes"
    "encoding/hex"
    "errors"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "github.com/prysmaticlabs/go-ssz"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Deposit amount in gwei
const DEPOSIT_AMOUNT uint64 = 32000000000


// Deposit status
type DepositStatus struct {
    Balances *node.Balances
    Reservation *node.ReservationDetails
}


// DepositData data
type DepositData struct {
    Pubkey [48]byte
    WithdrawalCredentials [32]byte
    Amount uint64
    Signature [96]byte
}


// RocketPool PoolCreated event
type PoolCreated struct {
    Address common.Address
    DurationID [32]byte
    Created *big.Int
}


// Get the node's current deposit status
func getDepositStatus(p *services.Provider) (*DepositStatus, error) {

    // Status channels
    balancesChannel := make(chan *node.Balances)
    reservationChannel := make(chan *node.ReservationDetails)
    errorChannel := make(chan error)

    // Get node balances
    go (func() {
        if balances, err := node.GetBalances(p.NodeContract); err != nil {
            errorChannel <- err
        } else {
            balancesChannel <- balances
        }
    })()

    // Get node deposit reservation details
    go (func() {
        if reservation, err := node.GetReservationDetails(p.NodeContract, p.CM); err != nil {
            errorChannel <- err
        } else {
            reservationChannel <- reservation
        }
    })()

    // Receive status
    var balances *node.Balances
    var reservation *node.ReservationDetails
    for received := 0; received < 2; {
        select {
        case balances = <-balancesChannel:
            received++
        case reservation = <-reservationChannel:
            received++
        case err := <-errorChannel:
            return nil, err
        }
    }

    // Return
    return &DepositStatus{
        Balances: balances,
        Reservation: reservation,
    }, nil

}


// Reserve a node deposit
func reserveDeposit(p *services.Provider, durationId string) error {

    // Generate new validator key
    key, err := p.KM.CreateValidatorKey()
    if err != nil {
        return errors.New("Error generating validator key: " + err.Error())
    }
    pubkey := key.PublicKey.Marshal()

    // Status channels
    successChannel := make(chan bool)
    messageChannel := make(chan string)
    errorChannel := make(chan error)

    // Check node does not have current deposit reservation
    go (func() {
        hasReservation := new(bool)
        if err := p.NodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation status: " + err.Error())
        } else if *hasReservation {
            messageChannel <- "Node has a current deposit reservation, please cancel or complete it"
        } else {
            successChannel <- true
        }
    })()

    // Check node deposits are enabled
    go (func() {
        depositsAllowed := new(bool)
        if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, depositsAllowed, "getDepositAllowed"); err != nil {
            errorChannel <- errors.New("Error checking node deposits enabled status: " + err.Error())
        } else if !*depositsAllowed {
            messageChannel <- "Node deposits are currently disabled in Rocket Pool"
        } else {
            successChannel <- true
        }
    })()

    // Check pubkey is not in use
    go (func() {
        pubkeyUsedKey := eth.KeccakBytes(bytes.Join([][]byte{[]byte("validator.pubkey.used"), pubkey}, []byte{}))
        if pubkeyUsed, err := p.CM.RocketStorage.GetBool(nil, pubkeyUsedKey); err != nil {
            errorChannel <- errors.New("Error retrieving pubkey used status: " + err.Error())
        } else if pubkeyUsed {
            messageChannel <- "The public key is already in use"
        } else {
            successChannel <- true
        }
    })()

    // Receive status
    for received := 0; received < 3; {
        select {
            case <-successChannel:
                received++
            case msg := <-messageChannel:
                fmt.Fprintln(p.Output, msg)
                return nil
            case err := <-errorChannel:
                return err
        }
    }

    // Get RP withdrawal pubkey
    // :TODO: replace with correct withdrawal pubkey once available
    withdrawalPubkeyHex := []byte("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
    withdrawalPubkey := make([]byte, hex.DecodedLen(len(withdrawalPubkeyHex)))
    _,_ = hex.Decode(withdrawalPubkey, withdrawalPubkeyHex)

    // Build withdrawal credentials
    withdrawalCredentials := eth.KeccakBytes(withdrawalPubkey) // Withdrawal pubkey hash
    withdrawalCredentials[0] = 0 // Replace first byte with BLS_WITHDRAWAL_PREFIX_BYTE

    // Build DepositData object
    depositData := &DepositData{}
    copy(depositData.Pubkey[:], pubkey)
    copy(depositData.WithdrawalCredentials[:], withdrawalCredentials[:])
    depositData.Amount = DEPOSIT_AMOUNT

    // Build signature
    signingRoot, err := ssz.SigningRoot(depositData)
    if err != nil {
        return errors.New("Error retrieving deposit data hash tree root: " + err.Error())
    }
    signature := key.SecretKey.Sign(signingRoot[:]).Marshal()

    // Create deposit reservation
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        fmt.Fprintln(p.Output, "Making deposit reservation...")
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "depositReserve", durationId, pubkey, signature); err != nil {
            return errors.New("Error making deposit reservation: " + err.Error())
        }
    }

    // Return
    return nil

}


// Complete a node deposit
func completeDeposit(p *services.Provider) (*PoolCreated, error) {

    // Status channels
    successChannel := make(chan bool)
    messageChannel := make(chan string)
    errorChannel := make(chan error)

    // Check node has current deposit reservation
    go (func() {
        hasReservation := new(bool)
        if err := p.NodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation status: " + err.Error())
        } else if !*hasReservation {
            messageChannel <- "Node does not have a current deposit reservation, please make one with `rocketpool deposit reserve durationID`"
        } else {
            successChannel <- true
        }
    })()

    // Check node deposits are enabled
    go (func() {
        depositsAllowed := new(bool)
        if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, depositsAllowed, "getDepositAllowed"); err != nil {
            errorChannel <- errors.New("Error checking node deposits enabled status: " + err.Error())
        } else if !*depositsAllowed {
            messageChannel <- "Node deposits are currently disabled in Rocket Pool"
        } else {
            successChannel <- true
        }
    })()

    // Check minipool creation is enabled
    go (func() {
        minipoolCreationAllowed := new(bool)
        if err := p.CM.Contracts["rocketMinipoolSettings"].Call(nil, minipoolCreationAllowed, "getMinipoolCanBeCreated"); err != nil {
            errorChannel <- errors.New("Error checking minipool creation enabled status: " + err.Error())
        } else if !*minipoolCreationAllowed {
            messageChannel <- "Minipool creation is currently disabled in Rocket Pool"
        } else {
            successChannel <- true
        }
    })()

    // Receive status
    for received := 0; received < 3; {
        select {
            case <-successChannel:
                received++
            case msg := <-messageChannel:
                fmt.Fprintln(p.Output, msg)
                return nil, nil
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Get deposit reservation validator pubkey
    validatorPubkey := new([]byte)
    if err := p.NodeContract.Call(nil, validatorPubkey, "getDepositReserveValidatorPubkey"); err != nil {
        return nil, errors.New("Error retrieving deposit reservation validator pubkey: " + err.Error())
    }

    // Check for local validator key
    if _, err := p.KM.GetValidatorKey(*validatorPubkey); err != nil {
        return nil, errors.New("Local validator key matching deposit reservation validator pubkey not found")
    }

    // Data channels
    accountBalancesChannel := make(chan *node.Balances)
    nodeBalancesChannel := make(chan *node.Balances)
    requiredBalancesChannel := make(chan *node.Balances)
    depositDurationIDChannel := make(chan string)

    // Get node account balances
    go (func() {
        nodeAccount, _ := p.AM.GetNodeAccount()
        if accountBalances, err := node.GetAccountBalances(nodeAccount.Address, p.Client, p.CM); err != nil {
            errorChannel <- err
        } else {
            accountBalancesChannel <- accountBalances
        }
    })()

    // Get node balances
    go (func() {
        if nodeBalances, err := node.GetBalances(p.NodeContract); err != nil {
            errorChannel <- err
        } else {
            nodeBalancesChannel <- nodeBalances
        }
    })()

    // Get node balance requirements
    go (func() {
        if requiredBalances, err := node.GetRequiredBalances(p.NodeContract); err != nil {
            errorChannel <- err
        } else {
            requiredBalancesChannel <- requiredBalances
        }
    })()

    // Get deposit duration ID
    go (func() {
        durationID := new(string)
        if err := p.NodeContract.Call(nil, durationID, "getDepositReserveDurationID"); err != nil {
            errorChannel <- errors.New("Error retrieving deposit duration ID: " + err.Error())
        } else {
            depositDurationIDChannel <- *durationID
        }
    })()

    // Receive data
    var accountBalances *node.Balances
    var nodeBalances *node.Balances
    var requiredBalances *node.Balances
    var depositDurationID string
    for received := 0; received < 4; {
        select {
            case accountBalances = <-accountBalancesChannel:
                received++
            case nodeBalances = <-nodeBalancesChannel:
                received++
            case requiredBalances = <-requiredBalancesChannel:
                received++
            case depositDurationID = <-depositDurationIDChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Check node ether balance and get required deposit transaction value
    depositTransactionValueWei := new(big.Int)
    if nodeBalances.EtherWei.Cmp(requiredBalances.EtherWei) < 0 {

        // Get remaining ether balance required
        remainingEtherRequiredWei := new(big.Int)
        remainingEtherRequiredWei.Sub(requiredBalances.EtherWei, nodeBalances.EtherWei)

        // Check node account balance
        if accountBalances.EtherWei.Cmp(remainingEtherRequiredWei) < 0 {
            fmt.Fprintln(p.Output, fmt.Sprintf("Node balance of %.2f ETH plus account balance of %.2f ETH is not enough to cover requirement of %.2f ETH", eth.WeiToEth(nodeBalances.EtherWei), eth.WeiToEth(accountBalances.EtherWei), eth.WeiToEth(requiredBalances.EtherWei)))
            return nil, nil
        }

        // Confirm transfer of remaining required ether
        response := cliutils.Prompt(p.Input, p.Output, fmt.Sprintf("Node contract requires %.2f ETH to complete deposit, would you like to pay now from your node account? [y/n]", eth.WeiToEth(remainingEtherRequiredWei)), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            fmt.Fprintln(p.Output, "Deposit not completed")
            return nil, nil
        }

        // Set deposit transaction value
        depositTransactionValueWei.Set(remainingEtherRequiredWei)

    }

    // Check node RPL balance and transfer remaining required RPL
    if nodeBalances.RplWei.Cmp(requiredBalances.RplWei) < 0 {

        // Get remaining RPL balance required
        remainingRplRequiredWei := new(big.Int)
        remainingRplRequiredWei.Sub(requiredBalances.RplWei, nodeBalances.RplWei)

        // Check node account balance
        if accountBalances.RplWei.Cmp(remainingRplRequiredWei) < 0 {
            fmt.Fprintln(p.Output, fmt.Sprintf("Node balance of %.2f RPL plus account balance of %.2f RPL is not enough to cover requirement of %.2f RPL", eth.WeiToEth(nodeBalances.RplWei), eth.WeiToEth(accountBalances.RplWei), eth.WeiToEth(requiredBalances.RplWei)))
            return nil, nil
        }

        // Confirm transfer of remaining required RPL
        response := cliutils.Prompt(p.Input, p.Output, fmt.Sprintf("Node contract requires %.2f RPL to complete deposit, would you like to pay now from your node account? [y/n]", eth.WeiToEth(remainingRplRequiredWei)), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            fmt.Fprintln(p.Output, "Deposit not completed")
            return nil, nil
        }

        // Transfer remaining required RPL
        if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
            return nil, err
        } else {
            fmt.Fprintln(p.Output, "Transferring RPL to node contract...")
            if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses["rocketPoolToken"], p.CM.Abis["rocketPoolToken"], "transfer", p.NodeContractAddress, remainingRplRequiredWei); err != nil {
                return nil, errors.New("Error transferring RPL to node contract: " + err.Error())
            }
        }

    }

    // Complete deposit
    txor, err := p.AM.GetNodeAccountTransactor()
    if err != nil { return nil, err }
    txor.Value = depositTransactionValueWei
    fmt.Fprintln(p.Output, "Completing deposit...")
    txReceipt, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "deposit")
    if err != nil {
        return nil, errors.New("Error completing deposit: " + err.Error())
    }

    // Get minipool created event
    minipoolCreatedEvents, err := eth.GetTransactionEvents(p.Client, txReceipt, p.CM.Addresses["rocketPool"], p.CM.Abis["rocketPool"], "PoolCreated", PoolCreated{})
    if err != nil {
        return nil, errors.New("Error retrieving deposit transaction minipool created event: " + err.Error())
    } else if len(minipoolCreatedEvents) == 0 {
        return nil, errors.New("Could not retrieve deposit transaction minipool created event")
    }
    minipoolCreatedEvent := (minipoolCreatedEvents[0]).(*PoolCreated)

    // Process deposit queue for duration
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else {
        fmt.Fprintln(p.Output, "Processing deposit queue...")
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses["rocketDepositQueue"], p.CM.Abis["rocketDepositQueue"], "assignChunks", depositDurationID); err != nil {
            return nil, errors.New("Error processing deposit queue: " + err.Error())
        }
    }

    // Return
    return minipoolCreatedEvent, nil

}


// Cancel a node deposit reservation
func cancelDeposit(p *services.Provider) error {

    // Check node has current deposit reservation
    hasReservation := new(bool)
    if err := p.NodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
        return errors.New("Error retrieving deposit reservation status: " + err.Error())
    } else if !*hasReservation {
        fmt.Fprintln(p.Output, "Node does not have a current deposit reservation")
        return nil
    }

    // Cancel deposit reservation
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        fmt.Fprintln(p.Output, "Canceling deposit reservation...")
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "depositReserveCancel"); err != nil {
            return errors.New("Error canceling deposit reservation: " + err.Error())
        }
    }

    // Return
    return nil

}

