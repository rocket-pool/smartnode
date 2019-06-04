package deposit

import (
    "bytes"
    "encoding/hex"
    "errors"
    "fmt"

    //"github.com/prysmaticlabs/prysm/shared/ssz"
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


/*
// DepositInput data
type DepositInput struct {
    pubkey [48]byte
    withdrawalCredentials [32]byte
    proofOfPossession [96]byte
}
*/


// Reserve a node deposit
func reserveDeposit(c *cli.Context, durationId string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        KM: true,
        ClientSync: true,
        CM: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        LoadAbis: []string{"rocketNodeContract"},
    })
    if err != nil {
        return err 
    }

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
                fmt.Println(msg)
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

    // Build proof of possession
    // :TODO: implement once BLS library is available
    proofOfPossessionHex := []byte(
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" +
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" +
        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
    proofOfPossession := make([]byte, hex.DecodedLen(len(proofOfPossessionHex)))
    _,_ = hex.Decode(proofOfPossession, proofOfPossessionHex)

    // Build DepositInput
    // :TODO: implement using SSZ once library is available
    var depositInputLength [4]byte
    depositInputLength[0] = byte(len(pubkey) + len(withdrawalCredentials) + len(proofOfPossession))
    depositInput := bytes.Join([][]byte{depositInputLength[:], pubkey, withdrawalCredentials[:], proofOfPossession}, []byte{})

    /*
    depositInputData := &DepositInput{}
    copy(depositInputData.pubkey[:], pubkey)
    copy(depositInputData.withdrawalCredentials[:], withdrawalCredentials[:])
    copy(depositInputData.proofOfPossession[:], proofOfPossession)
    depositInput := new(bytes.Buffer)
    if err := ssz.Encode(depositInput, depositInputData); err != nil {
        return errors.New("Error encoding DepositInput for deposit reservation: " + err.Error())
    }
    */

    // Create deposit reservation
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        fmt.Println("Making deposit reservation...")
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "depositReserve", durationId, depositInput); err != nil {
            return errors.New("Error making deposit reservation: " + err.Error())
        }
    }

    // Get deposit reservation details
    reservation, err := node.GetReservationDetails(p.NodeContract, p.CM)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Println(fmt.Sprintf(
        "Deposit reservation made successfully, requiring %.2f ETH and %.2f RPL, with a staking duration of %s and expiring at %s",
        eth.WeiToEth(reservation.EtherRequiredWei),
        eth.WeiToEth(reservation.RplRequiredWei),
        reservation.StakingDurationID,
        reservation.ExpiryTime.Format("2006-01-02, 15:04 -0700 MST")))
    return nil

}

