package upgrades

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"golang.org/x/sync/errgroup"
)

// Settings
const UpgradeProposalDetailsBatchSize = 50

// Upgrade proposal details
type UpgradeProposalDetails struct {
	ID             uint64                       `json:"id"`
	State          rptypes.UpgradeProposalState `json:"state"`
	EndTime        *big.Int                     `json:"endTime"`
	Name           string                       `json:"name"`
	Type           [32]byte                     `json:"type"`
	UpgradeAddress common.Address               `json:"upgradeAddress"`
	UpgradeAbi     string                       `json:"upgradeAbi"`
}

// Get all upgrade proposal details
func GetUpgradeProposals(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]UpgradeProposalDetails, error) {

	// Get proposal count
	proposalCount, err := GetTotalUpgradeProposals(rp, opts)
	if err != nil {
		return []UpgradeProposalDetails{}, err
	}

	// Load proposal details in batches
	details := make([]UpgradeProposalDetails, proposalCount)
	for bsi := uint64(0); bsi < proposalCount; bsi += UpgradeProposalDetailsBatchSize {

		// Get batch start & end index
		psi := bsi
		pei := bsi + UpgradeProposalDetailsBatchSize
		if pei > proposalCount {
			pei = proposalCount
		}

		// Load details
		var wg errgroup.Group
		for pi := psi; pi < pei; pi++ {
			pi := pi
			wg.Go(func() error {
				proposalDetails, err := GetUpgradeProposalDetails(rp, pi+1, opts) // Proposals are 1-indexed
				if err == nil {
					details[pi] = proposalDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []UpgradeProposalDetails{}, err
		}

	}

	// Return
	return details, nil

}

// Get a proposal's details
func GetUpgradeProposalDetails(rp *rocketpool.RocketPool, proposalId uint64, opts *bind.CallOpts) (UpgradeProposalDetails, error) {

	// Data
	var wg errgroup.Group
	var state rptypes.UpgradeProposalState
	var endTime *big.Int
	var name string
	var proposalType [32]byte
	var upgradeAddress common.Address
	var upgradeAbi string

	// Load data
	wg.Go(func() error {
		var err error
		name, err = GetUpgradeProposalName(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		proposalType, err = GetUpgradeProposalType(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		upgradeAddress, err = GetUpgradeProposalUpgradeAddress(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		upgradeAbi, err = GetUpgradeProposalUpgradeAbi(rp, proposalId, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		endTime, err = GetUpgradeProposalEndTime(rp, proposalId, opts)
		return err
	})

	wg.Go(func() error {
		var err error
		state, err = GetUpgradeProposalState(rp, proposalId, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return UpgradeProposalDetails{}, err
	}

	// Return
	return UpgradeProposalDetails{
		ID:             proposalId,
		EndTime:        endTime,
		Name:           name,
		Type:           proposalType,
		UpgradeAddress: upgradeAddress,
		UpgradeAbi:     upgradeAbi,
		State:          state,
	}, nil

}
