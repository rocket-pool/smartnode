package dao

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/rocketpool-go/utils/strings"
)

// Settings
const (
	ProposalDAONamesBatchSize = 50
	ProposalDetailsBatchSize  = 10
)

// Proposal details
type ProposalDetails struct {
	ID              uint64                `json:"id"`
	DAO             string                `json:"dao"`
	ProposerAddress common.Address        `json:"proposerAddress"`
	Message         string                `json:"message"`
	CreatedTime     uint64                `json:"createdTime"`
	StartTime       uint64                `json:"startTime"`
	EndTime         uint64                `json:"endTime"`
	ExpiryTime      uint64                `json:"expiryTime"`
	VotesRequired   float64               `json:"votesRequired"`
	VotesFor        float64               `json:"votesFor"`
	VotesAgainst    float64               `json:"votesAgainst"`
	MemberVoted     bool                  `json:"memberVoted"`
	MemberSupported bool                  `json:"memberSupported"`
	IsCancelled     bool                  `json:"isCancelled"`
	IsExecuted      bool                  `json:"isExecuted"`
	Payload         []byte                `json:"payload"`
	PayloadStr      string                `json:"payloadStr"`
	State           rptypes.ProposalState `json:"state"`
}

// Get all proposal details
func GetProposals(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]ProposalDetails, error) {

	// Get proposal count
	proposalCount, err := GetProposalCount(rp, opts)
	if err != nil {
		return []ProposalDetails{}, err
	}

	// Load proposal details in batches
	details := make([]ProposalDetails, proposalCount)
	for bsi := uint64(0); bsi < proposalCount; bsi += ProposalDetailsBatchSize {

		// Get batch start & end index
		psi := bsi
		pei := bsi + ProposalDetailsBatchSize
		if pei > proposalCount {
			pei = proposalCount
		}

		// Load details
		var wg errgroup.Group
		for pi := psi; pi < pei; pi++ {
			pi := pi
			wg.Go(func() error {
				proposalDetails, err := GetProposalDetails(rp, pi+1, opts) // Proposals are 1-indexed
				if err == nil {
					details[pi] = proposalDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []ProposalDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get all proposal details with member data
func GetProposalsWithMember(rp *rocketpool.RocketPool, memberAddress common.Address, opts *bind.CallOpts) ([]ProposalDetails, error) {

	// Get proposal count
	proposalCount, err := GetProposalCount(rp, opts)
	if err != nil {
		return []ProposalDetails{}, err
	}

	// Load proposal details in batches
	details := make([]ProposalDetails, proposalCount)
	for bsi := uint64(0); bsi < proposalCount; bsi += ProposalDetailsBatchSize {

		// Get batch start & end index
		psi := bsi
		pei := bsi + ProposalDetailsBatchSize
		if pei > proposalCount {
			pei = proposalCount
		}

		// Load details
		var wg errgroup.Group
		for pi := psi; pi < pei; pi++ {
			pi := pi
			wg.Go(func() error {
				proposalDetails, err := GetProposalDetailsWithMember(rp, pi+1, memberAddress, opts) // Proposals are 1-indexed
				if err == nil {
					details[pi] = proposalDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []ProposalDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get DAO proposal details
func GetDAOProposals(rp *rocketpool.RocketPool, daoName string, opts *bind.CallOpts) ([]ProposalDetails, error) {

	// Get DAO proposal IDs
	proposalIds, err := GetDAOProposalIDs(rp, daoName, opts)
	if err != nil {
		return []ProposalDetails{}, err
	}

	// Load proposal details in batches
	details := make([]ProposalDetails, len(proposalIds))
	for bsi := 0; bsi < len(proposalIds); bsi += ProposalDetailsBatchSize {

		// Get batch start & end index
		psi := bsi
		pei := bsi + ProposalDetailsBatchSize
		if pei > len(proposalIds) {
			pei = len(proposalIds)
		}

		// Load details
		var wg errgroup.Group
		for pi := psi; pi < pei; pi++ {
			pi := pi
			wg.Go(func() error {
				proposalDetails, err := GetProposalDetails(rp, proposalIds[pi], opts)
				if err == nil {
					details[pi] = proposalDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []ProposalDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get DAO proposal details with member data
func GetDAOProposalsWithMember(rp *rocketpool.RocketPool, daoName string, memberAddress common.Address, opts *bind.CallOpts) ([]ProposalDetails, error) {

	// Get DAO proposal IDs
	proposalIds, err := GetDAOProposalIDs(rp, daoName, opts)
	if err != nil {
		return []ProposalDetails{}, err
	}

	// Load proposal details in batches
	details := make([]ProposalDetails, len(proposalIds))
	for bsi := 0; bsi < len(proposalIds); bsi += ProposalDetailsBatchSize {

		// Get batch start & end index
		psi := bsi
		pei := bsi + ProposalDetailsBatchSize
		if pei > len(proposalIds) {
			pei = len(proposalIds)
		}

		// Load details
		var wg errgroup.Group
		for pi := psi; pi < pei; pi++ {
			pi := pi
			wg.Go(func() error {
				proposalDetails, err := GetProposalDetailsWithMember(rp, proposalIds[pi], memberAddress, opts)
				if err == nil {
					details[pi] = proposalDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []ProposalDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get the IDs of proposals filtered by a DAO
func GetDAOProposalIDs(rp *rocketpool.RocketPool, daoName string, opts *bind.CallOpts) ([]uint64, error) {

	// Get proposal count
	proposalCount, err := GetProposalCount(rp, opts)
	if err != nil {
		return []uint64{}, err
	}

	// Load proposal DAO names in batches
	proposalDaoNames := make([]string, proposalCount)
	for bsi := uint64(0); bsi < proposalCount; bsi += ProposalDAONamesBatchSize {

		// Get batch start & end index
		psi := bsi
		pei := bsi + ProposalDAONamesBatchSize
		if pei > proposalCount {
			pei = proposalCount
		}

		// Load details
		var wg errgroup.Group
		for pi := psi; pi < pei; pi++ {
			pi := pi
			wg.Go(func() error {
				proposalDaoName, err := GetProposalDAO(rp, pi+1, opts) // Proposals are 1-indexed
				if err == nil {
					proposalDaoNames[pi] = proposalDaoName
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []uint64{}, err
		}

	}

	// Get & return IDs for DAO proposals
	ids := []uint64{}
	for pi, proposalDaoName := range proposalDaoNames {
		if proposalDaoName == daoName {
			ids = append(ids, uint64(pi+1)) // Proposals are 1-indexed
		}
	}
	return ids, nil

}

// Get a proposal's details
func GetProposalDetails(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (ProposalDetails, error) {

	// Data
	var wg errgroup.Group
	var dao string
	var proposerAddress common.Address
	var message string
	var createdTime uint64
	var startTime uint64
	var endTime uint64
	var expiryTime uint64
	var votesRequired float64
	var votesFor float64
	var votesAgainst float64
	var isCancelled bool
	var isExecuted bool
	var payload []byte
	var state rptypes.ProposalState

	// Load data
	wg.Go(func() error {
		var err error
		dao, err = GetProposalDAO(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		proposerAddress, err = GetProposalProposerAddress(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		message, err = GetProposalMessage(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		createdTime, err = GetProposalCreatedTime(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		startTime, err = GetProposalStartTime(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		endTime, err = GetProposalEndTime(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		expiryTime, err = GetProposalExpiryTime(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		votesRequired, err = GetProposalVotesRequired(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		votesFor, err = GetProposalVotesFor(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		votesAgainst, err = GetProposalVotesAgainst(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		isCancelled, err = GetProposalIsCancelled(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		isExecuted, err = GetProposalIsExecuted(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		payload, err = GetProposalPayload(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		state, err = GetProposalState(rp, proposalId, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return ProposalDetails{}, err
	}

	// Get proposal payload string
	payloadStr, err := GetProposalPayloadString(rp, dao, payload, opts)
	if err != nil {
		payloadStr = "(unknown)"
	}

	// Return
	return ProposalDetails{
		ID:              proposalId,
		DAO:             dao,
		ProposerAddress: proposerAddress,
		Message:         message,
		CreatedTime:     createdTime,
		StartTime:       startTime,
		EndTime:         endTime,
		ExpiryTime:      expiryTime,
		VotesRequired:   votesRequired,
		VotesFor:        votesFor,
		VotesAgainst:    votesAgainst,
		IsCancelled:     isCancelled,
		IsExecuted:      isExecuted,
		Payload:         payload,
		PayloadStr:      payloadStr,
		State:           state,
	}, nil

}

// Get a proposal's details with member data
func GetProposalDetailsWithMember(rp *rocketpool.RocketPool, proposalId uint64, memberAddress common.Address, opts *bind.CallOpts) (ProposalDetails, error) {

	// Data
	var wg errgroup.Group
	var details ProposalDetails
	var memberVoted bool
	var memberSupported bool

	// Load data
	wg.Go(func() error {
		var err error
		details, err = GetProposalDetails(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		memberVoted, err = GetProposalMemberVoted(rp, proposalId, memberAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		memberSupported, err = GetProposalMemberSupported(rp, proposalId, memberAddress, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return ProposalDetails{}, err
	}

	// Return
	details.MemberVoted = memberVoted
	details.MemberSupported = memberSupported
	return details, nil

}

// Get the proposal count
func GetProposalCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return 0, err
	}
	proposalCount := new(*big.Int)
	if err := rocketDAOProposal.Call(opts, proposalCount, "getTotal"); err != nil {
		return 0, fmt.Errorf("Could not get proposal count: %w", err)
	}
	return (*proposalCount).Uint64(), nil
}

// Proposal details
func GetProposalDAO(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (string, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return "", err
	}
	daoName := new(string)
	if err := rocketDAOProposal.Call(opts, daoName, "getDAO", big.NewInt(int64(proposalId))); err != nil {
		return "", fmt.Errorf("Could not get proposal %d DAO: %w", proposalId, err)
	}
	return strings.Sanitize(*daoName), nil
}
func GetProposalProposerAddress(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (common.Address, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	proposerAddress := new(common.Address)
	if err := rocketDAOProposal.Call(opts, proposerAddress, "getProposer", big.NewInt(int64(proposalId))); err != nil {
		return common.Address{}, fmt.Errorf("Could not get proposal %d proposer address: %w", proposalId, err)
	}
	return *proposerAddress, nil
}
func GetProposalMessage(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (string, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return "", err
	}
	message := new(string)
	if err := rocketDAOProposal.Call(opts, message, "getMessage", big.NewInt(int64(proposalId))); err != nil {
		return "", fmt.Errorf("Could not get proposal %d message: %w", proposalId, err)
	}
	return strings.Sanitize(*message), nil
}
func GetProposalCreatedTime(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return 0, err
	}
	createdTime := new(*big.Int)
	if err := rocketDAOProposal.Call(opts, createdTime, "getCreated", big.NewInt(int64(proposalId))); err != nil {
		return 0, fmt.Errorf("Could not get proposal %d created time: %w", proposalId, err)
	}
	return (*createdTime).Uint64(), nil
}
func GetProposalStartTime(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return 0, err
	}
	startTime := new(*big.Int)
	if err := rocketDAOProposal.Call(opts, startTime, "getStart", big.NewInt(int64(proposalId))); err != nil {
		return 0, fmt.Errorf("Could not get proposal %d start time: %w", proposalId, err)
	}
	return (*startTime).Uint64(), nil
}
func GetProposalEndTime(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return 0, err
	}
	endTime := new(*big.Int)
	if err := rocketDAOProposal.Call(opts, endTime, "getEnd", big.NewInt(int64(proposalId))); err != nil {
		return 0, fmt.Errorf("Could not get proposal %d end time: %w", proposalId, err)
	}
	return (*endTime).Uint64(), nil
}
func GetProposalExpiryTime(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (uint64, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return 0, err
	}
	expiryTime := new(*big.Int)
	if err := rocketDAOProposal.Call(opts, expiryTime, "getExpires", big.NewInt(int64(proposalId))); err != nil {
		return 0, fmt.Errorf("Could not get proposal %d expiry time: %w", proposalId, err)
	}
	return (*expiryTime).Uint64(), nil
}
func GetProposalVotesRequired(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (float64, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return 0, err
	}
	votesRequired := new(*big.Int)
	if err := rocketDAOProposal.Call(opts, votesRequired, "getVotesRequired", big.NewInt(int64(proposalId))); err != nil {
		return 0, fmt.Errorf("Could not get proposal %d votes required: %w", proposalId, err)
	}
	return eth.WeiToEth(*votesRequired), nil
}
func GetProposalVotesFor(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (float64, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return 0, err
	}
	votesFor := new(*big.Int)
	if err := rocketDAOProposal.Call(opts, votesFor, "getVotesFor", big.NewInt(int64(proposalId))); err != nil {
		return 0, fmt.Errorf("Could not get proposal %d votes for: %w", proposalId, err)
	}
	return eth.WeiToEth(*votesFor), nil
}
func GetProposalVotesAgainst(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (float64, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return 0, err
	}
	votesAgainst := new(*big.Int)
	if err := rocketDAOProposal.Call(opts, votesAgainst, "getVotesAgainst", big.NewInt(int64(proposalId))); err != nil {
		return 0, fmt.Errorf("Could not get proposal %d votes against: %w", proposalId, err)
	}
	return eth.WeiToEth(*votesAgainst), nil
}
func GetProposalIsCancelled(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return false, err
	}
	cancelled := new(bool)
	if err := rocketDAOProposal.Call(opts, cancelled, "getCancelled", big.NewInt(int64(proposalId))); err != nil {
		return false, fmt.Errorf("Could not get proposal %d cancelled status: %w", proposalId, err)
	}
	return *cancelled, nil
}
func GetProposalIsExecuted(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (bool, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return false, err
	}
	executed := new(bool)
	if err := rocketDAOProposal.Call(opts, executed, "getExecuted", big.NewInt(int64(proposalId))); err != nil {
		return false, fmt.Errorf("Could not get proposal %d executed status: %w", proposalId, err)
	}
	return *executed, nil
}
func GetProposalPayload(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) ([]byte, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return []byte{}, err
	}
	payload := new([]byte)
	if err := rocketDAOProposal.Call(opts, payload, "getPayload", big.NewInt(int64(proposalId))); err != nil {
		return []byte{}, fmt.Errorf("Could not get proposal %d payload: %w", proposalId, err)
	}
	return *payload, nil
}
func GetProposalPayloadStr(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (string, error) {
	dao, err := GetProposalDAO(rp, proposalId, opts)
	if err != nil {
		return "", err
	}
	payload, err := GetProposalPayload(rp, proposalId, opts)
	if err != nil {
		return "", err
	}
	payloadStr, err := GetProposalPayloadString(rp, dao, payload, opts)
	if err != nil {
		payloadStr = "(unknown)"
	}
	return payloadStr, nil
}
func GetProposalState(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (rptypes.ProposalState, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return 0, err
	}
	state := new(uint8)
	if err := rocketDAOProposal.Call(opts, state, "getState", big.NewInt(int64(proposalId))); err != nil {
		return 0, fmt.Errorf("Could not get proposal %d state: %w", proposalId, err)
	}
	return rptypes.ProposalState(*state), nil
}

// Get whether a member has voted on a proposal
func GetProposalMemberVoted(rp *rocketpool.RocketPool, proposalId uint64, memberAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return false, err
	}
	voted := new(bool)
	if err := rocketDAOProposal.Call(opts, voted, "getReceiptHasVoted", big.NewInt(int64(proposalId)), memberAddress); err != nil {
		return false, fmt.Errorf("Could not get proposal %d member %s voted status: %w", proposalId, memberAddress.Hex(), err)
	}
	return *voted, nil
}

// Get whether a member has voted in support of a proposal
func GetProposalMemberSupported(rp *rocketpool.RocketPool, proposalId uint64, memberAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketDAOProposal, err := getRocketDAOProposal(rp, opts)
	if err != nil {
		return false, err
	}
	supported := new(bool)
	if err := rocketDAOProposal.Call(opts, supported, "getReceiptSupported", big.NewInt(int64(proposalId)), memberAddress); err != nil {
		return false, fmt.Errorf("Could not get proposal %d member %s supported status: %w", proposalId, memberAddress.Hex(), err)
	}
	return *supported, nil
}

// Get contracts
var rocketDAOProposalLock sync.Mutex

func getRocketDAOProposal(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDAOProposalLock.Lock()
	defer rocketDAOProposalLock.Unlock()
	return rp.GetContract("rocketDAOProposal", opts)
}
