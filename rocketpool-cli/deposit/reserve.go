package deposit

import (
    "bytes"
    "encoding/hex"
    "errors"
    "fmt"

    "github.com/prysmaticlabs/go-ssz"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Deposit amount in gwei
const DEPOSIT_AMOUNT uint64 = 32000000000


// DepositData data
type DepositData struct {
    Pubkey [48]byte
    WithdrawalCredentials [32]byte
    Amount uint64
    Signature [96]byte
}


// Reserve a node deposit
func reserveDeposit(c *cli.Context, durationId string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        KM: true,
        CM: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        LoadAbis: []string{"rocketNodeContract"},
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

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

    // Get deposit reservation details
    reservation, err := node.GetReservationDetails(p.NodeContract, p.CM)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Fprintln(p.Output, fmt.Sprintf(
        "Deposit reservation made successfully, requiring %.2f ETH and %.2f RPL, with a staking duration of %s and expiring at %s",
        eth.WeiToEth(reservation.EtherRequiredWei),
        eth.WeiToEth(reservation.RplRequiredWei),
        reservation.StakingDurationID,
        reservation.ExpiryTime.Format("2006-01-02, 15:04 -0700 MST")))
    return nil

}

