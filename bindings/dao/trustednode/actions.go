package trustednode

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Estimate the gas of Join
func EstimateJoinGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrustedActions.GetTransactionGasInfo(opts, "actionJoin")
}

// Join the trusted node DAO
// Requires an executed invite proposal
func Join(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAONodeTrustedActions.Transact(opts, "actionJoin")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error joining the trusted node DAO: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of Leave
func EstimateLeaveGas(rp *rocketpool.RocketPool, rplBondRefundAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrustedActions.GetTransactionGasInfo(opts, "actionLeave", rplBondRefundAddress)
}

// Leave the trusted node DAO
// Requires an executed leave proposal
func Leave(rp *rocketpool.RocketPool, rplBondRefundAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAONodeTrustedActions.Transact(opts, "actionLeave", rplBondRefundAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error leaving the trusted node DAO: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of MakeChallenge
func EstimateMakeChallengeGas(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrustedActions.GetTransactionGasInfo(opts, "actionChallengeMake", memberAddress)
}

// Make a challenge against a node
func MakeChallenge(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAONodeTrustedActions.Transact(opts, "actionChallengeMake", memberAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error challenging trusted node DAO member %s: %w", memberAddress.Hex(), err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of DecideChallenge
func EstimateDecideChallengeGas(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrustedActions.GetTransactionGasInfo(opts, "actionChallengeDecide", memberAddress)
}

// Decide a challenge against a node
func DecideChallenge(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAONodeTrustedActions, err := getRocketDAONodeTrustedActions(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAONodeTrustedActions.Transact(opts, "actionChallengeDecide", memberAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error deciding the challenge against trusted node DAO member %s: %w", memberAddress.Hex(), err)
	}
	return tx.Hash(), nil
}

// Returns the most recent block number that the number of trusted nodes changed since fromBlock
func GetLatestMemberCountChangedBlock(rp *rocketpool.RocketPool, fromBlock uint64, intervalSize *big.Int, opts *bind.CallOpts) (uint64, error) {
	// Get contracts
	rocketDaoNodeTrustedActions, err := getRocketDAONodeTrustedActions(rp, opts)
	if err != nil {
		return 0, err
	}
	// Construct a filter query for relevant logs
	addressFilter := []common.Address{*rocketDaoNodeTrustedActions.Address}
	topicFilter := [][]common.Hash{{rocketDaoNodeTrustedActions.ABI.Events["ActionJoined"].ID, rocketDaoNodeTrustedActions.ABI.Events["ActionLeave"].ID, rocketDaoNodeTrustedActions.ABI.Events["ActionKick"].ID, rocketDaoNodeTrustedActions.ABI.Events["ActionChallengeDecided"].ID}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, big.NewInt(int64(fromBlock)), nil, nil)
	if err != nil {
		return 0, err
	}

	for i := range logs {
		log := logs[len(logs)-i-1]
		if log.Topics[0] == rocketDaoNodeTrustedActions.ABI.Events["ActionChallengeDecided"].ID {
			values := make(map[string]interface{})
			// Decode the event
			if rocketDaoNodeTrustedActions.ABI.Events["ActionChallengeDecided"].Inputs.UnpackIntoMap(values, log.Data) != nil {
				return 0, err
			}
			if values["success"].(bool) {
				return log.BlockNumber, nil
			}
		} else {
			return log.BlockNumber, nil
		}
	}
	return fromBlock, nil
}

// Get contracts
var rocketDAONodeTrustedActionsLock sync.Mutex

func getRocketDAONodeTrustedActions(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAONodeTrustedActionsLock.Lock()
	defer rocketDAONodeTrustedActionsLock.Unlock()
	return rp.GetContract("rocketDAONodeTrustedActions", opts)
}
