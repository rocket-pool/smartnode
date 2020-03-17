package main

import (
    "encoding/hex"
    "errors"
    "fmt"
    "log"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/prysmaticlabs/go-ssz"
    "github.com/prysmaticlabs/prysm/shared/bls"
    "github.com/prysmaticlabs/prysm/shared/bytesutil"

    "github.com/rocket-pool/smartnode/shared/contracts"
    "github.com/rocket-pool/smartnode/shared/services/validators"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Config
const MIN_VALIDATOR_COUNT int = 64
const DEPOSIT_AMOUNT uint64 = 32000000000
const DOMAIN_DEPOSIT uint64 = 3
const DEPOSIT_CONTRACT_ADDRESS string = "0xb50eA9565646e5Ed39688694b283cb185A3CC130"
const GAS_LIMIT uint64 = 8000000


// DepositData data
type DepositData struct {
    Pubkey [48]byte
    WithdrawalCredentials [32]byte
    Amount uint64
    Signature [96]byte
}


// Run
func main() {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { log.Fatal(errors.New("Could not create account manager: " + err.Error())) }

    // Create key manager
    km, err := test.NewInitKeyManager("foobarbaz")
    if err != nil { log.Fatal(errors.New("Could not create key manager: " + err.Error())) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { log.Fatal(errors.New("Could not connect to ethereum client: " + err.Error())) }

    // Initialise deposit contract
    depositContract, err := contracts.NewDepositContract(common.HexToAddress(DEPOSIT_CONTRACT_ADDRESS), client)
    if err != nil { log.Fatal(errors.New("Could not initialise deposit contract: " + err.Error())) }

    // Seed node account
    account, err := am.GetNodeAccount()
    if err != nil { log.Fatal(errors.New("Could not get node account: " + err.Error())) }
    if err := test.SeedAccount(client, account.Address, eth.EthToWei(32000)); err != nil { log.Fatal(errors.New("Could not seed node account: " + err.Error())) }

    // Create validators
    for vi := 0; vi < MIN_VALIDATOR_COUNT; vi++ {

        // Get deposit data
        depositData, depositDataRoot, err := getValidatorDepositData(km)
        if err != nil { log.Fatal(errors.New("Could not get validator deposit data: " + err.Error())) }

        // Get account txor
        txor, err := am.GetNodeAccountTransactor()
        if err != nil { log.Fatal(errors.New("Could not get node account transactor: " + err.Error())) }
        txor.Value = eth.EthToWei(32)
        txor.GasLimit = GAS_LIMIT

        // Deposit
        _, err = depositContract.Deposit(txor, depositData.Pubkey[:], depositData.WithdrawalCredentials[:], depositData.Signature[:], depositDataRoot)
        if err != nil { log.Fatal(errors.New("Could not deposit to deposit contract: " + err.Error())) }

        // Log
        fmt.Println(fmt.Sprintf("Validator %d registered successfully.", vi))

    }

    // Log
    fmt.Println(fmt.Sprintf("%d validators registered successfully.", MIN_VALIDATOR_COUNT))

}


// Create validator & get deposit data
func getValidatorDepositData(km *validators.KeyManager) (*DepositData, [32]byte, error) {

    // Generate new validator key
    key, err := km.CreateValidatorKey()
    if err != nil { return nil, [32]byte{}, err }
    pubkey := key.PublicKey.Marshal()

    // Get RP withdrawal pubkey
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

    // Get deposit data signing root
    signingRoot, err := ssz.SigningRoot(depositData)
    if err != nil { return nil, [32]byte{}, err }

    // Sign deposit data
    domain := bls.ComputeDomain(bytesutil.ToBytes4(bytesutil.Bytes4(DOMAIN_DEPOSIT)))
    signature := key.SecretKey.Sign(signingRoot[:], domain).Marshal()
    copy(depositData.Signature[:], signature)

    // Get deposit data root
    depositDataRoot, err := ssz.HashTreeRoot(depositData)
    if err != nil { return nil, [32]byte{}, err }

    // Return
    return depositData, depositDataRoot, nil

}

