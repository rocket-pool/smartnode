package rocketpool

import (
    "encoding/hex"
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/prysmaticlabs/go-ssz"

    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/services/validators"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Deposit amount in gwei
const DEPOSIT_AMOUNT uint64 = 32000000000


// RocketNodeAPI NodeAdd event
type NodeAdd struct {
    ID common.Address
    ContractAddress common.Address
    Created *big.Int
}


// RocketPool PoolCreated event
type PoolCreated struct {
    Address common.Address
    DurationID [32]byte
    Created *big.Int
}


// DepositData data
type DepositData struct {
    Pubkey [48]byte
    WithdrawalCredentials [32]byte
    Amount uint64
    Signature [96]byte
}


// Register a node
func RegisterNode(client *ethclient.Client, cm *rocketpool.ContractManager, am *accounts.AccountManager) (*bind.BoundContract, common.Address, error) {

    // Seed node account
    account, err := am.GetNodeAccount()
    if err != nil { return nil, common.Address{}, err }
    if err := test.SeedAccount(client, account.Address, eth.EthToWei(10)); err != nil { return nil, common.Address{}, err }

    // Register node
    txor, err := am.GetNodeAccountTransactor()
    if err != nil { return nil, common.Address{}, err }
    txReceipt, err := eth.ExecuteContractTransaction(client, txor, cm.Addresses["rocketNodeAPI"], cm.Abis["rocketNodeAPI"], "add", "Australia/Brisbane")
    if err != nil { return nil, common.Address{}, err }

    // Get NodeAdd event
    nodeAddEvents, err := eth.GetTransactionEvents(client, txReceipt, cm.Addresses["rocketNodeAPI"], cm.Abis["rocketNodeAPI"], "NodeAdd", NodeAdd{})
    if err != nil {
        return nil, common.Address{}, err
    } else if len(nodeAddEvents) == 0 {
        return nil, common.Address{}, errors.New("Failed to retrieve NodeAdd event")
    }
    nodeAddEvent := (nodeAddEvents[0]).(*NodeAdd)

    // Create and return node contract
    nodeContract, err := cm.NewContract(&nodeAddEvent.ContractAddress, "rocketNodeContract")
    if err != nil { return nil, common.Address{}, err }

    // Return
    return nodeContract, nodeAddEvent.ContractAddress, nil

}


// Create a node deposit reservation
func ReserveNodeDeposit(client *ethclient.Client, cm *rocketpool.ContractManager, am *accounts.AccountManager, km *validators.KeyManager, nodeContractAddress common.Address, durationId string) error {

    // Generate new validator key
    key, err := km.CreateValidatorKey()
    if err != nil { return err }
    pubkey := key.PublicKey.Marshal()

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
    if err != nil { return err }
    signature := key.SecretKey.Sign(signingRoot[:]).Marshal()

    // Reserve deposit
    txor, err := am.GetNodeAccountTransactor()
    if err != nil { return err }
    if _, err := eth.ExecuteContractTransaction(client, txor, &nodeContractAddress, cm.Abis["rocketNodeContract"], "depositReserve", durationId, pubkey, signature); err != nil { return err }

    // Return
    return nil

}


// Create a minipool under a node
func CreateNodeMinipool(client *ethclient.Client, cm *rocketpool.ContractManager, am *accounts.AccountManager, km *validators.KeyManager, nodeContract *bind.BoundContract, nodeContractAddress common.Address, durationId string) (common.Address, error) {

    // Reserve deposit
    if err := ReserveNodeDeposit(client, cm, am, km, nodeContractAddress, durationId); err != nil { return common.Address{}, err }

    // Get required balances
    requiredBalances, err := node.GetRequiredBalances(nodeContract)
    if err != nil { return common.Address{}, err }

    // Seed node contract
    if requiredBalances.EtherWei.Cmp(big.NewInt(0)) == 1 {
        if err := test.SeedAccount(client, nodeContractAddress, requiredBalances.EtherWei); err != nil { return common.Address{}, err }
    }
    if requiredBalances.RplWei.Cmp(big.NewInt(0)) == 1 {
        if err := MintRPL(client, cm, nodeContractAddress, requiredBalances.RplWei); err != nil { return common.Address{}, err }
    }

    // Complete deposit
    txor, err := am.GetNodeAccountTransactor()
    if err != nil { return common.Address{}, err }
    txReceipt, err := eth.ExecuteContractTransaction(client, txor, &nodeContractAddress, cm.Abis["rocketNodeContract"], "deposit")
    if err != nil { return common.Address{}, err }

    // Get minipool created event
    minipoolCreatedEvents, err := eth.GetTransactionEvents(client, txReceipt, cm.Addresses["rocketPool"], cm.Abis["rocketPool"], "PoolCreated", PoolCreated{})
    if err != nil {
        return common.Address{}, err
    } else if len(minipoolCreatedEvents) == 0 {
        return common.Address{}, errors.New("Failed to retrieve PoolCreated event")
    }
    minipoolCreatedEvent := (minipoolCreatedEvents[0]).(*PoolCreated)

    // Return
    return minipoolCreatedEvent.Address, nil

}

