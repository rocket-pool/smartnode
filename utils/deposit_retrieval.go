package utils

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// BeaconDepositEvent represents a DepositEvent event raised by the BeaconDeposit contract.
type BeaconDepositEvent struct {
	Pubkey                []byte
	WithdrawalCredentials []byte
	Amount                []byte
	Signature             []byte
	Index                 []byte
	Raw                   types.Log // Blockchain specific contextual infos
}


// Formatted Beacon deposit event data
type DepositData struct {
    Pubkey rptypes.ValidatorPubkey          `json:"pubkey"`
    WithdrawalCredentials common.Hash       `json:"withdrawalCredentials"`
    Amount uint64                           `json:"amount"`
    Signature rptypes.ValidatorSignature    `json:"signature"`
    TxHash common.Hash                      `json:"txHash"`
}


// Gets all of the deposit contract's deposit events for the provided minipool addresses 
func GetMinipoolDeposits(rp *rocketpool.RocketPool, minipoolAddresses []common.Address, startBlock *big.Int, intervalSize *big.Int, opts *bind.CallOpts) ( map[common.Address][]DepositData, error ) {

    // Get the deposit contract wrapper
    casperDeposit, err := getCasperDeposit(rp)
    if err != nil {
        return nil, err
    }

    // Create the initial map and pubkey lookup
    depositMap := make(map[common.Address][]DepositData)
    pubkeyLookup := make(map[rptypes.ValidatorPubkey]common.Address)
    for _, address := range minipoolAddresses {
        pubkey, err := minipool.GetMinipoolPubkey(rp, address, nil)
        if err != nil {
            return nil, err
        }
        pubkeyLookup[pubkey] = address
        depositMap[address] = []DepositData{}
    }

    // Get the deposit events
    addressFilter := []common.Address{*casperDeposit.Address}
    topicFilter := [][]common.Hash{{casperDeposit.ABI.Events["DepositEvent"].ID}}
    logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, startBlock, nil, nil)
    if err != nil {
        return nil, err
    }

    // Process each event
    for _, log := range logs {
        depositEvent := new(BeaconDepositEvent)
        casperDeposit.Contract.UnpackLog(depositEvent, "DepositEvent", log)

        // Check if this is a deposit for one of the minipools we're looking for
        pubkey := rptypes.BytesToValidatorPubkey(depositEvent.Pubkey)
        minipoolAddress, exists := pubkeyLookup[pubkey]
        if exists {
            // Convert the deposit amount from little-endian binary to a uint64
            var amount uint64
            buf := bytes.NewReader(depositEvent.Amount)
            err = binary.Read(buf, binary.LittleEndian, &amount)
            if err != nil {
                return nil, err
            }

            // Create the deposit data wrapper and add it to this minipool's collection
            depositData := DepositData{
                Pubkey: pubkey,
                WithdrawalCredentials: common.BytesToHash(depositEvent.WithdrawalCredentials),
                Amount: amount,
                Signature: rptypes.BytesToValidatorSignature(depositEvent.Signature),
                TxHash: log.TxHash,
            }
            depositMap[minipoolAddress] = append(depositMap[minipoolAddress], depositData)
        }
    }

    return depositMap, nil
}


// Get contracts
var casperDepositLock sync.Mutex
func getCasperDeposit(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    casperDepositLock.Lock()
    defer casperDepositLock.Unlock()
    return rp.GetContract("casperDeposit")
}

