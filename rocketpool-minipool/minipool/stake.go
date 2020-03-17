package minipool

import (
    "bytes"
    "encoding/hex"
    "errors"

    "github.com/prysmaticlabs/go-ssz"
    "github.com/prysmaticlabs/prysm/shared/bls"
    "github.com/prysmaticlabs/prysm/shared/bytesutil"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
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


// Stake minipool
func Stake(p *services.Provider, pool *Minipool) error {

    // Check minipool status
    if status, err := minipool.GetStatusCode(p.CM, pool.Address); err != nil {
        return errors.New("Error retrieving minipool status: " + err.Error())
    } else if status != minipool.PRELAUNCH {
        return nil
    }

    // Get Rocket Pool withdrawal credentials
    withdrawalCredentialsBytes32 := new([32]byte)
    if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, withdrawalCredentialsBytes32, "getWithdrawalCredentials"); err != nil {
        return errors.New("Error retrieving Rocket Pool withdrawal credentials: " + err.Error())
    }
    withdrawalCredentials := (*withdrawalCredentialsBytes32)[:]

    // Check withdrawal credentials
    if bytes.Equal(withdrawalCredentials, make([]byte, 32)) {
        return errors.New("Rocket Pool withdrawal credentials have not been initialized")
    }

    // Generate new validator key
    validatorKey, err := p.KM.CreateValidatorKey()
    if err != nil { return err }
    validatorPubkey := validatorKey.PublicKey.Marshal()

    // Build DepositData object
    depositData := &DepositData{}
    copy(depositData.Pubkey[:], validatorPubkey)
    copy(depositData.WithdrawalCredentials[:], withdrawalCredentials)
    depositData.Amount = DEPOSIT_AMOUNT

    // Get deposit data signing root
    signingRoot, err := ssz.SigningRoot(depositData)
    if err != nil {
        return errors.New("Error retrieving deposit data signing root: " + err.Error())
    }

    // Sign deposit data
    domain := bls.ComputeDomain(bytesutil.ToBytes4(bytesutil.Bytes4(DOMAIN_DEPOSIT)))
    signature := validatorKey.SecretKey.Sign(signingRoot[:], domain).Marshal()
    copy(depositData.Signature[:], signature)

    // Get deposit data root
    depositDataRoot, err := ssz.HashTreeRoot(depositData)
    if err != nil {
        return errors.New("Error retrieving deposit data hash tree root: " + err.Error())
    }

    // Stake minipool
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], "stakeMinipool", pool.Address, validatorPubkey, signature, depositDataRoot); err != nil {
            return errors.New("Error staking minipool: " + err.Error())
        }
    }

    // Log
    p.Log.Println("Successfully staked minipool...")

    // Encode validator pubkey and add to minipool data
    validatorPubkeyHex := make([]byte, hex.EncodedLen(len(validatorPubkey)))
    hex.Encode(validatorPubkeyHex, validatorPubkey)
    validatorPubkeyStr := string(validatorPubkeyHex)
    pool.Pubkey = validatorPubkeyStr

    // Return
    return nil

}

