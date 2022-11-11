package trustednode

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/strings"
)

// Settings
const (
	MemberAddressBatchSize = 50
	MemberDetailsBatchSize = 20
)

// Proposal details
type MemberDetails struct {
	Address                common.Address `json:"address"`
	Exists                 bool           `json:"exists"`
	ID                     string         `json:"id"`
	Url                    string         `json:"url"`
	JoinedTime             uint64         `json:"joinedTime"`
	LastProposalTime       uint64         `json:"lastProposalTime"`
	RPLBondAmount          *big.Int       `json:"rplBondAmount"`
	UnbondedValidatorCount uint64         `json:"unbondedValidatorCount"`
}

// Get all member details
func GetMembers(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]MemberDetails, error) {

	// Get member addresses
	memberAddresses, err := GetMemberAddresses(rp, opts)
	if err != nil {
		return []MemberDetails{}, err
	}

	// Load member details in batches
	details := make([]MemberDetails, len(memberAddresses))
	for bsi := 0; bsi < len(memberAddresses); bsi += MemberDetailsBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MemberDetailsBatchSize
		if mei > len(memberAddresses) {
			mei = len(memberAddresses)
		}

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				memberAddress := memberAddresses[mi]
				memberDetails, err := GetMemberDetails(rp, memberAddress, opts)
				if err == nil {
					details[mi] = memberDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []MemberDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get all member addresses
func GetMemberAddresses(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]common.Address, error) {

	// Get member count
	memberCount, err := GetMemberCount(rp, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Load member addresses in batches
	addresses := make([]common.Address, memberCount)
	for bsi := uint64(0); bsi < memberCount; bsi += MemberAddressBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := bsi + MemberAddressBatchSize
		if mei > memberCount {
			mei = memberCount
		}

		// Load addresses
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address, err := GetMemberAt(rp, mi, opts)
				if err == nil {
					addresses[mi] = address
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []common.Address{}, err
		}

	}

	// Return
	return addresses, nil

}

// Get a member's details
func GetMemberDetails(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (MemberDetails, error) {

	// Data
	var wg errgroup.Group
	var exists bool
	var id string
	var url string
	var joinedTime uint64
	var lastProposalTime uint64
	var rplBondAmount *big.Int
	var unbondedValidatorCount uint64

	// Load data
	wg.Go(func() error {
		var err error
		exists, err = GetMemberExists(rp, memberAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		id, err = GetMemberID(rp, memberAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		url, err = GetMemberUrl(rp, memberAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		joinedTime, err = GetMemberJoinedTime(rp, memberAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		lastProposalTime, err = GetMemberLastProposalTime(rp, memberAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		rplBondAmount, err = GetMemberRPLBondAmount(rp, memberAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		unbondedValidatorCount, err = GetMemberUnbondedValidatorCount(rp, memberAddress, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return MemberDetails{}, err
	}

	// Return
	return MemberDetails{
		Address:                memberAddress,
		Exists:                 exists,
		ID:                     id,
		Url:                    url,
		JoinedTime:             joinedTime,
		LastProposalTime:       lastProposalTime,
		RPLBondAmount:          rplBondAmount,
		UnbondedValidatorCount: unbondedValidatorCount,
	}, nil

}

// Get the minimum member count
func GetMinimumMemberCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return 0, err
	}
	minMemberCount := new(*big.Int)
	if err := rocketDAONodeTrusted.Call(opts, minMemberCount, "getMemberMinRequired"); err != nil {
		return 0, fmt.Errorf("Could not get trusted node DAO minimum member count: %w", err)
	}
	return (*minMemberCount).Uint64(), nil
}

// Get the member count
func GetMemberCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return 0, err
	}
	memberCount := new(*big.Int)
	if err := rocketDAONodeTrusted.Call(opts, memberCount, "getMemberCount"); err != nil {
		return 0, fmt.Errorf("Could not get trusted node DAO member count: %w", err)
	}
	return (*memberCount).Uint64(), nil
}

// Get a member address by index
func GetMemberAt(rp *rocketpool.RocketPool, index uint64, opts *bind.CallOpts) (common.Address, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	memberAddress := new(common.Address)
	if err := rocketDAONodeTrusted.Call(opts, memberAddress, "getMemberAt", big.NewInt(int64(index))); err != nil {
		return common.Address{}, fmt.Errorf("Could not get trusted node DAO member %d address: %w", index, err)
	}
	return *memberAddress, nil
}

// Member details
func GetMemberExists(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return false, err
	}
	exists := new(bool)
	if err := rocketDAONodeTrusted.Call(opts, exists, "getMemberIsValid", memberAddress); err != nil {
		return false, fmt.Errorf("Could not get trusted node DAO member %s exists status: %w", memberAddress.Hex(), err)
	}
	return *exists, nil
}
func GetMemberID(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (string, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return "", err
	}
	id := new(string)
	if err := rocketDAONodeTrusted.Call(opts, id, "getMemberID", memberAddress); err != nil {
		return "", fmt.Errorf("Could not get trusted node DAO member %s ID: %w", memberAddress.Hex(), err)
	}
	return strings.Sanitize(*id), nil
}
func GetMemberUrl(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (string, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return "", err
	}
	url := new(string)
	if err := rocketDAONodeTrusted.Call(opts, url, "getMemberUrl", memberAddress); err != nil {
		return "", fmt.Errorf("Could not get trusted node DAO member %s URL: %w", memberAddress.Hex(), err)
	}
	return strings.Sanitize(*url), nil
}
func GetMemberJoinedTime(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return 0, err
	}
	joinedTime := new(*big.Int)
	if err := rocketDAONodeTrusted.Call(opts, joinedTime, "getMemberJoinedTime", memberAddress); err != nil {
		return 0, fmt.Errorf("Could not get trusted node DAO member %s joined time: %w", memberAddress.Hex(), err)
	}
	return (*joinedTime).Uint64(), nil
}
func GetMemberLastProposalTime(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return 0, err
	}
	lastProposalTime := new(*big.Int)
	if err := rocketDAONodeTrusted.Call(opts, lastProposalTime, "getMemberLastProposalTime", memberAddress); err != nil {
		return 0, fmt.Errorf("Could not get trusted node DAO member %s last proposal time: %w", memberAddress.Hex(), err)
	}
	return (*lastProposalTime).Uint64(), nil
}
func GetMemberRPLBondAmount(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return nil, err
	}
	rplBondAmount := new(*big.Int)
	if err := rocketDAONodeTrusted.Call(opts, rplBondAmount, "getMemberRPLBondAmount", memberAddress); err != nil {
		return nil, fmt.Errorf("Could not get trusted node DAO member %s RPL bond amount: %w", memberAddress.Hex(), err)
	}
	return *rplBondAmount, nil
}
func GetMemberUnbondedValidatorCount(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return 0, err
	}
	unbondedValidatorCount := new(*big.Int)
	if err := rocketDAONodeTrusted.Call(opts, unbondedValidatorCount, "getMemberUnbondedValidatorCount", memberAddress); err != nil {
		return 0, fmt.Errorf("Could not get trusted node DAO member %s unbonded validator count: %w", memberAddress.Hex(), err)
	}
	return (*unbondedValidatorCount).Uint64(), nil
}

// Get the time that a proposal for a member was executed at
func GetMemberInviteProposalExecutedTime(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	return GetMemberProposalExecutedTime(rp, "invited", memberAddress, opts)
}
func GetMemberLeaveProposalExecutedTime(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	return GetMemberProposalExecutedTime(rp, "leave", memberAddress, opts)
}
func GetMemberReplaceProposalExecutedTime(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	return GetMemberProposalExecutedTime(rp, "replace", memberAddress, opts)
}
func GetMemberProposalExecutedTime(rp *rocketpool.RocketPool, proposalType string, memberAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return 0, err
	}
	proposalExecutedTime := new(*big.Int)
	if err := rocketDAONodeTrusted.Call(opts, proposalExecutedTime, "getMemberProposalExecutedTime", proposalType, memberAddress); err != nil {
		return 0, fmt.Errorf("Could not get trusted node DAO %s proposal executed time for member %s: %w", proposalType, memberAddress.Hex(), err)
	}
	return (*proposalExecutedTime).Uint64(), nil
}

// Get a member's replacement address if being replaced
func GetMemberReplacementAddress(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (common.Address, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	replacementAddress := new(common.Address)
	if err := rocketDAONodeTrusted.Call(opts, replacementAddress, "getMemberReplacedAddress", "new", memberAddress); err != nil {
		return common.Address{}, fmt.Errorf("Could not get trusted node DAO member %s replacement address: %w", memberAddress.Hex(), err)
	}
	return *replacementAddress, nil
}

// Get whether a member has an active challenge against them
func GetMemberIsChallenged(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, opts)
	if err != nil {
		return false, err
	}
	isChallenged := new(bool)
	if err := rocketDAONodeTrusted.Call(opts, isChallenged, "getMemberIsChallenged", memberAddress); err != nil {
		return false, fmt.Errorf("Could not get trusted node DAO member %s is challenged status: %w", memberAddress.Hex(), err)
	}
	return *isChallenged, nil
}

// Estimate the gas of BootstrapBool
func EstimateBootstrapBoolGas(rp *rocketpool.RocketPool, contractName, settingPath string, value bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrusted.GetTransactionGasInfo(opts, "bootstrapSettingBool", contractName, settingPath, value)
}

// Bootstrap a bool setting
func BootstrapBool(rp *rocketpool.RocketPool, contractName, settingPath string, value bool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAONodeTrusted.Transact(opts, "bootstrapSettingBool", contractName, settingPath, value)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not bootstrap trusted node setting %s.%s: %w", contractName, settingPath, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of BootstrapUint
func EstimateBootstrapUintGas(rp *rocketpool.RocketPool, contractName, settingPath string, value *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrusted.GetTransactionGasInfo(opts, "bootstrapSettingUint", contractName, settingPath, value)
}

// Bootstrap a uint256 setting
func BootstrapUint(rp *rocketpool.RocketPool, contractName, settingPath string, value *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAONodeTrusted.Transact(opts, "bootstrapSettingUint", contractName, settingPath, value)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not bootstrap trusted node setting %s.%s: %w", contractName, settingPath, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of BootstrapMember
func EstimateBootstrapMemberGas(rp *rocketpool.RocketPool, id, url string, nodeAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	url = strings.Sanitize(url)
	return rocketDAONodeTrusted.GetTransactionGasInfo(opts, "bootstrapMember", id, url, nodeAddress)
}

// Bootstrap a DAO member
func BootstrapMember(rp *rocketpool.RocketPool, id, url string, nodeAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	url = strings.Sanitize(url)
	tx, err := rocketDAONodeTrusted.Transact(opts, "bootstrapMember", id, url, nodeAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not bootstrap trusted node member %s: %w", id, err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of BootstrapUpgrade
func EstimateBootstrapUpgradeGas(rp *rocketpool.RocketPool, upgradeType, contractName, contractAbi string, contractAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	compressedAbi, err := rocketpool.EncodeAbiStr(contractAbi)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDAONodeTrusted.GetTransactionGasInfo(opts, "bootstrapUpgrade", upgradeType, contractName, compressedAbi, contractAddress)
}

// Bootstrap a contract upgrade
func BootstrapUpgrade(rp *rocketpool.RocketPool, upgradeType, contractName, contractAbi string, contractAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	compressedAbi, err := rocketpool.EncodeAbiStr(contractAbi)
	if err != nil {
		return common.Hash{}, err
	}
	rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDAONodeTrusted.Transact(opts, "bootstrapUpgrade", upgradeType, contractName, compressedAbi, contractAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not bootstrap contract '%s' upgrade (%s): %w", contractName, upgradeType, err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketDAONodeTrustedLock sync.Mutex

func getRocketDAONodeTrusted(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAONodeTrustedLock.Lock()
	defer rocketDAONodeTrustedLock.Unlock()
	return rp.GetContract("rocketDAONodeTrusted", opts)
}
