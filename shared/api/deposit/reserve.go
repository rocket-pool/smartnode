package deposit

import (
    "bytes"
    "encoding/hex"
    "errors"

    "github.com/prysmaticlabs/go-ssz"
    "github.com/prysmaticlabs/prysm/shared/bls"
    "github.com/prysmaticlabs/prysm/shared/bytesutil"
    "github.com/prysmaticlabs/prysm/shared/keystore"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Deposit amount in gwei
const DEPOSIT_AMOUNT uint64 = 32000000000


// BLS deposit domain
const DOMAIN_DEPOSIT uint64 = 3


// DepositData data
type DepositData struct {
    Pubkey [48]byte
    WithdrawalCredentials [32]byte
    Amount uint64
    Signature [96]byte
}


// Reserve deposit response type
type CanReserveDepositResponse struct {

    // Status
    Success bool                    `json:"success"`

    // Failure reasons
    HadExistingReservation bool     `json:"hadExistingReservation"`
    DepositsDisabled bool           `json:"depositsDisabled"`
    StakingDurationDisabled bool    `json:"stakingDurationDisabled"`
    PubkeyUsed bool                 `json:"pubkeyUsed"`

}
type ReserveDepositResponse struct {
    Success bool                    `json:"success"`
}


// Check node deposit can be reserved
func CanReserveDeposit(p *services.Provider, validatorKey *keystore.Key, durationId string) (*CanReserveDepositResponse, error) {

    // Response
    response := &CanReserveDepositResponse{}

    // Get validator pubkey
    validatorPubkey := validatorKey.PublicKey.Marshal()

    // Status channels
    hasExistingReservationChannel := make(chan bool)
    depositsDisabledChannel := make(chan bool)
    stakingDurationDisabledChannel := make(chan bool)
    pubkeyUsedChannel := make(chan bool)
    errorChannel := make(chan error)

    // Check node does not have current deposit reservation
    go (func() {
        hasReservation := new(bool)
        if err := p.NodeContract.Call(nil, hasReservation, "getHasDepositReservation"); err != nil {
            errorChannel <- errors.New("Error retrieving deposit reservation status: " + err.Error())
        } else {
            hasExistingReservationChannel <- *hasReservation
        }
    })()

    // Check node deposits are enabled
    go (func() {
        depositsAllowed := new(bool)
        if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, depositsAllowed, "getDepositAllowed"); err != nil {
            errorChannel <- errors.New("Error checking node deposits enabled status: " + err.Error())
        } else {
            depositsDisabledChannel <- !*depositsAllowed
        }
    })()

    // Check staking duration is enabled
    go (func() {
        stakingDurationEnabled := new(bool)
        if err := p.CM.Contracts["rocketMinipoolSettings"].Call(nil, stakingDurationEnabled, "getMinipoolStakingDurationEnabled", durationId); err != nil {
            errorChannel <- errors.New("Error checking staking duration enabled status: " + err.Error())
        } else {
            stakingDurationDisabledChannel <- !*stakingDurationEnabled
        }
    })()

    // Check pubkey is not in use
    go (func() {
        pubkeyUsedKey := eth.KeccakBytes(bytes.Join([][]byte{[]byte("validator.pubkey.used"), validatorPubkey}, []byte{}))
        if pubkeyUsed, err := p.CM.RocketStorage.GetBool(nil, pubkeyUsedKey); err != nil {
            errorChannel <- errors.New("Error retrieving pubkey used status: " + err.Error())
        } else {
            pubkeyUsedChannel <- pubkeyUsed
        }
    })()

    // Receive status
    for received := 0; received < 4; {
        select {
            case response.HadExistingReservation = <-hasExistingReservationChannel:
                received++
            case response.DepositsDisabled = <-depositsDisabledChannel:
                received++
            case response.StakingDurationDisabled = <-stakingDurationDisabledChannel:
                received++
            case response.PubkeyUsed = <-pubkeyUsedChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Update & return response
    response.Success = !(response.HadExistingReservation || response.DepositsDisabled || response.StakingDurationDisabled || response.PubkeyUsed)
    return response, nil

}


// Reserve node deposit
func ReserveDeposit(p *services.Provider, validatorKey *keystore.Key, durationId string) (*ReserveDepositResponse, error) {

    // Get validator pubkey
    validatorPubkey := validatorKey.PublicKey.Marshal()

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
    copy(depositData.Pubkey[:], validatorPubkey)
    copy(depositData.WithdrawalCredentials[:], withdrawalCredentials[:])
    depositData.Amount = DEPOSIT_AMOUNT

    // Get deposit data signing root
    signingRoot, err := ssz.SigningRoot(depositData)
    if err != nil {
        return nil, errors.New("Error retrieving deposit data signing root: " + err.Error())
    }

    // Sign deposit data
    domain := bls.ComputeDomain(bytesutil.Bytes4(DOMAIN_DEPOSIT))
    signature := validatorKey.SecretKey.Sign(signingRoot[:], domain).Marshal()
    copy(depositData.Signature[:], signature)

    // Get deposit data root
    depositDataRoot, err := ssz.HashTreeRoot(depositData)
    if err != nil {
        return nil, errors.New("Error retrieving deposit data hash tree root: " + err.Error())
    }

    // Create deposit reservation
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "depositReserve", durationId, validatorPubkey, signature, depositDataRoot); err != nil {
            return nil, errors.New("Error making deposit reservation: " + err.Error())
        }
    }

    // Return response
    return &ReserveDepositResponse{
        Success: true,
    }, nil

}

