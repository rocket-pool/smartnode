package deposit

import (
    "bytes"
    "encoding/hex"
    "errors"
    "fmt"

    //"github.com/prysmaticlabs/prysm/shared/ssz"
    "github.com/urfave/cli"

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

    // Command setup
    am, rp, nodeContract, message, err := setup(c, []string{"rocketNodeAPI", "rocketNodeSettings"});
    if message != "" {
        fmt.Println(message)
        return nil
    }
    if err != nil {
        return err
    }

    // Check node deposits are enabled
    depositsAllowed := new(bool)
    err = rp.Contracts["rocketNodeSettings"].Call(nil, depositsAllowed, "getDepositAllowed")
    if err != nil {
        return errors.New("Error checking node deposits enabled status: " + err.Error())
    }
    if !*depositsAllowed {
        fmt.Println("Node deposits are currently disabled in Rocket Pool")
        return nil
    }

    // Check node does not have current deposit reservation
    hasReservation := new(bool)
    err = nodeContract.Call(nil, hasReservation, "getHasDepositReservation")
    if err != nil {
        return errors.New("Error retrieving deposit reservation status: " + err.Error())
    }
    if *hasReservation {
        fmt.Println("Node has a current deposit reservation, please cancel or complete it")
        return nil
    }

    // Get node's validator pubkey
    // :TODO: implement once BLS library is available
    pubkeyHex := []byte("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcd01")
    pubkey := make([]byte, hex.DecodedLen(len(pubkeyHex)))
    _,_ = hex.Decode(pubkey, pubkeyHex)

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
    err = ssz.Encode(depositInput, depositInputData)
    if err != nil {
        return errors.New("Error encoding DepositInput for deposit reservation: " + err.Error())
    }
    */

    // Get node account transactor
    nodeAccountTransactor, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Create deposit reservation
    _, err = nodeContract.Transact(nodeAccountTransactor, "depositReserve", durationId, depositInput)
    if err != nil {
        return errors.New("Error making deposit reservation: " + err.Error())
    }

    // Get deposit reservation details
    reservation, err := node.GetReservationDetails(nodeContract, rp)
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

