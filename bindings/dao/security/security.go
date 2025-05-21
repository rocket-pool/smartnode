package security

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/strings"
	"golang.org/x/sync/errgroup"
)

// Settings
const (
	MemberAddressBatchSize = 50
	MemberDetailsBatchSize = 20
)

// Member details
type SecurityDAOMemberDetails struct {
	Address    common.Address `json:"address"`
	Exists     bool           `json:"exists"`
	ID         string         `json:"id"`
	JoinedTime uint64         `json:"joinedTime"`
}

// Get all member details
func GetMembers(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]SecurityDAOMemberDetails, error) {
	// Get member addresses
	memberAddresses, err := GetMemberAddresses(rp, opts)
	if err != nil {
		return []SecurityDAOMemberDetails{}, err
	}

	// Load member details in batches
	details := make([]SecurityDAOMemberDetails, len(memberAddresses))
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
			return []SecurityDAOMemberDetails{}, err
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
func GetMemberDetails(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (SecurityDAOMemberDetails, error) {
	// Data
	var wg errgroup.Group
	var exists bool
	var id string
	var joinedTime uint64

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
		joinedTime, err = GetMemberJoinedTime(rp, memberAddress, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return SecurityDAOMemberDetails{}, err
	}

	// Return
	return SecurityDAOMemberDetails{
		Address:    memberAddress,
		Exists:     exists,
		ID:         id,
		JoinedTime: joinedTime,
	}, nil
}

// Get the amount of member votes need for a proposal to pass (as a fraction of 1e18)
func GetMemberQuorumVotesRequired(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketDAOSecurity, err := getRocketDAOSecurity(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketDAOSecurity.Call(opts, value, "getMemberQuorumVotesRequired"); err != nil {
		return nil, fmt.Errorf("error getting security DAO quorum votes required: %w", err)
	}
	return *value, nil
}

// Get the member count
func GetMemberCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketDAOSecurity, err := getRocketDAOSecurity(rp, opts)
	if err != nil {
		return 0, err
	}
	memberCount := new(*big.Int)
	if err := rocketDAOSecurity.Call(opts, memberCount, "getMemberCount"); err != nil {
		return 0, fmt.Errorf("error getting security DAO member count: %w", err)
	}
	return (*memberCount).Uint64(), nil
}

// Get a member address by index
func GetMemberAt(rp *rocketpool.RocketPool, index uint64, opts *bind.CallOpts) (common.Address, error) {
	rocketDAOSecurity, err := getRocketDAOSecurity(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	memberAddress := new(common.Address)
	if err := rocketDAOSecurity.Call(opts, memberAddress, "getMemberAt", big.NewInt(int64(index))); err != nil {
		return common.Address{}, fmt.Errorf("error getting security DAO member %d address: %w", index, err)
	}
	return *memberAddress, nil
}

// Member details
func GetMemberExists(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketDAOSecurity, err := getRocketDAOSecurity(rp, opts)
	if err != nil {
		return false, err
	}
	exists := new(bool)
	if err := rocketDAOSecurity.Call(opts, exists, "getMemberIsValid", memberAddress); err != nil {
		return false, fmt.Errorf("error getting security DAO member %s exists status: %w", memberAddress.Hex(), err)
	}
	return *exists, nil
}
func GetMemberID(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (string, error) {
	rocketDAOSecurity, err := getRocketDAOSecurity(rp, opts)
	if err != nil {
		return "", err
	}
	id := new(string)
	if err := rocketDAOSecurity.Call(opts, id, "getMemberID", memberAddress); err != nil {
		return "", fmt.Errorf("error getting security DAO member %s ID: %w", memberAddress.Hex(), err)
	}
	return strings.Sanitize(*id), nil
}
func GetMemberJoinedTime(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketDAOSecurity, err := getRocketDAOSecurity(rp, opts)
	if err != nil {
		return 0, err
	}
	joinedTime := new(*big.Int)
	if err := rocketDAOSecurity.Call(opts, joinedTime, "getMemberJoinedTime", memberAddress); err != nil {
		return 0, fmt.Errorf("error getting security DAO member %s joined time: %w", memberAddress.Hex(), err)
	}
	return (*joinedTime).Uint64(), nil
}

// Get the time that a proposal for a member was executed at
func GetMemberInviteProposalExecutedTime(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	return GetMemberProposalExecutedTime(rp, "invited", memberAddress, opts)
}
func GetMemberLeaveProposalExecutedTime(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	return GetMemberProposalExecutedTime(rp, "leave", memberAddress, opts)
}
func GetMemberProposalExecutedTime(rp *rocketpool.RocketPool, proposalType string, memberAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketDAOSecurity, err := getRocketDAOSecurity(rp, opts)
	if err != nil {
		return 0, err
	}
	proposalExecutedTime := new(*big.Int)
	if err := rocketDAOSecurity.Call(opts, proposalExecutedTime, "getMemberProposalExecutedTime", proposalType, memberAddress); err != nil {
		return 0, fmt.Errorf("error getting security DAO %s proposal executed time for member %s: %w", proposalType, memberAddress.Hex(), err)
	}
	return (*proposalExecutedTime).Uint64(), nil
}

// Get contracts
var rocketDAOSecurityLock sync.Mutex

func getRocketDAOSecurity(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOSecurityLock.Lock()
	defer rocketDAOSecurityLock.Unlock()
	return rp.GetContract("rocketDAOSecurity", opts)
}
