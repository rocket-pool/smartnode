package state

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	pDaoPropDetailsBatchSize int = 50
)

// Proposal details
type protocolDaoProposalDetailsRaw struct {
	ID                   uint64
	DAO                  string
	ProposerAddress      common.Address
	TargetBlock          *big.Int
	Message              string
	CreatedTime          *big.Int
	ChallengeWindow      *big.Int
	VotingStartTime      *big.Int
	Phase1EndTime        *big.Int
	Phase2EndTime        *big.Int
	ExpiryTime           *big.Int
	VotingPowerRequired  *big.Int
	VotingPowerFor       *big.Int
	VotingPowerAgainst   *big.Int
	VotingPowerAbstained *big.Int
	VotingPowerToVeto    *big.Int
	IsDestroyed          bool
	IsFinalized          bool
	IsExecuted           bool
	IsVetoed             bool
	VetoQuorum           *big.Int
	Payload              []byte
	PayloadStr           string
	State                uint8
	ProposalBond         *big.Int
	ChallengeBond        *big.Int
	DefeatIndex          *big.Int
}

// Gets a Protocol DAO proposal's details using the efficient multicall contract
func GetProtocolDaoProposalDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, proposalID uint64) (protocol.ProtocolDaoProposalDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	details := protocol.ProtocolDaoProposalDetails{}
	rawDetails := protocolDaoProposalDetailsRaw{}
	details.ID = proposalID

	err := addProposalCalls(contracts, contracts.Multicaller, &rawDetails)
	if err != nil {
		return details, fmt.Errorf("error adding proposal calls: %w", err)
	}

	_, err = contracts.Multicaller.FlexibleCall(true, opts)
	if err != nil {
		return details, fmt.Errorf("error executing multicall: %w", err)
	}

	fixupPdaoProposalDetails(rp, &rawDetails, &details, opts)

	return details, nil
}

// Gets all Protocol DAO proposal details using the efficient multicall contract
func GetAllProtocolDaoProposalDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts) ([]protocol.ProtocolDaoProposalDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	// Get the number of proposals available
	propCount, err := protocol.GetTotalProposalCount(rp, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting proposal count: %w", err)
	}

	// Make the proposal IDs (1-indexed) and return the details
	ids := make([]uint64, propCount)
	for i := range ids {
		ids[i] = uint64(i + 1)
	}
	return getProposalDetails(rp, contracts, ids, opts)
}

// Get the details of all protocol DAO proposals
func getProposalDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, ids []uint64, opts *bind.CallOpts) ([]protocol.ProtocolDaoProposalDetails, error) {
	propDetailsRaw := make([]protocolDaoProposalDetailsRaw, len(ids))

	// Get the details in batches
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	count := len(propDetailsRaw)
	for i := 0; i < count; i += pDaoPropDetailsBatchSize {
		i := i
		m := min(i+pDaoPropDetailsBatchSize, count)

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < m; j++ {
				id := ids[j]
				details := &propDetailsRaw[j]
				details.ID = id

				err = addProposalCalls(contracts, mc, details)
				if err != nil {
					return fmt.Errorf("error adding proposal calls: %w", err)
				}
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting Protocol DAO proposal details: %w", err)
	}

	// Postprocessing
	props := make([]protocol.ProtocolDaoProposalDetails, len(ids))
	for i := range propDetailsRaw {
		rawDetails := &propDetailsRaw[i]
		details := &props[i]
		fixupPdaoProposalDetails(rp, rawDetails, details, opts)
	}

	return props, nil
}

// Get the details of a proposal
func addProposalCalls(contracts *NetworkContracts, mc *multicall.MultiCaller, details *protocolDaoProposalDetailsRaw) error {
	id := big.NewInt(0).SetUint64(details.ID)
	allErrors := make([]error, 0)
	addCall := func(contract *rocketpool.Contract, out any, method string, args ...any) {
		allErrors = append(allErrors, mc.AddCall(contract, out, method, args...))
	}
	addCall(contracts.RocketDAOProtocolProposal, &details.ProposerAddress, "getProposer", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.DAO, "getDAO", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.TargetBlock, "getProposalBlock", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.Message, "getMessage", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.VotingStartTime, "getStart", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.Phase1EndTime, "getPhase1End", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.Phase2EndTime, "getPhase2End", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.ExpiryTime, "getExpires", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.CreatedTime, "getCreated", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.VotingPowerRequired, "getVotingPowerRequired", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.VotingPowerFor, "getVotingPowerFor", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.VotingPowerAgainst, "getVotingPowerAgainst", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.VotingPowerAbstained, "getVotingPowerAbstained", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.VotingPowerToVeto, "getVotingPowerVeto", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.IsDestroyed, "getDestroyed", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.IsFinalized, "getFinalised", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.IsExecuted, "getExecuted", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.IsVetoed, "getVetoed", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.VetoQuorum, "getProposalVetoQuorum", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.Payload, "getPayload", id)
	addCall(contracts.RocketDAOProtocolProposal, &details.State, "getState", id)
	addCall(contracts.RocketDAOProtocolVerifier, &details.DefeatIndex, "getDefeatIndex", id)
	addCall(contracts.RocketDAOProtocolVerifier, &details.ProposalBond, "getProposalBond", id)
	addCall(contracts.RocketDAOProtocolVerifier, &details.ChallengeBond, "getChallengeBond", id)
	addCall(contracts.RocketDAOProtocolVerifier, &details.ChallengeWindow, "getChallengePeriod", id)
	for _, err := range allErrors {
		if err != nil {
			return fmt.Errorf("error adding proposal calls: %w", err)
		}
	}
	return nil
}

// Converts a raw proposal to a well-formatted one
func fixupPdaoProposalDetails(rp *rocketpool.RocketPool, rawDetails *protocolDaoProposalDetailsRaw, details *protocol.ProtocolDaoProposalDetails, opts *bind.CallOpts) {
	details.ID = rawDetails.ID
	details.DAO = rawDetails.DAO
	details.ProposerAddress = rawDetails.ProposerAddress
	details.TargetBlock = uint32(rawDetails.TargetBlock.Uint64())
	details.Message = rawDetails.Message
	details.VotingStartTime = time.Unix(rawDetails.VotingStartTime.Int64(), 0)
	details.Phase1EndTime = time.Unix(rawDetails.Phase1EndTime.Int64(), 0)
	details.Phase2EndTime = time.Unix(rawDetails.Phase2EndTime.Int64(), 0)
	details.ExpiryTime = time.Unix(rawDetails.ExpiryTime.Int64(), 0)
	details.CreatedTime = time.Unix(rawDetails.CreatedTime.Int64(), 0)
	details.VotingPowerRequired = rawDetails.VotingPowerRequired
	details.VotingPowerFor = rawDetails.VotingPowerFor
	details.VotingPowerAgainst = rawDetails.VotingPowerAgainst
	details.VotingPowerAbstained = rawDetails.VotingPowerAbstained
	details.VotingPowerToVeto = rawDetails.VotingPowerToVeto
	details.IsDestroyed = rawDetails.IsDestroyed
	details.IsFinalized = rawDetails.IsFinalized
	details.IsExecuted = rawDetails.IsExecuted
	details.IsVetoed = rawDetails.IsVetoed
	details.VetoQuorum = rawDetails.VetoQuorum
	details.Payload = rawDetails.Payload
	details.State = types.ProtocolDaoProposalState(rawDetails.State)
	details.DefeatIndex = rawDetails.DefeatIndex.Uint64()
	details.ProposalBond = rawDetails.ProposalBond
	details.ChallengeBond = rawDetails.ChallengeBond
	details.ChallengeWindow = time.Second * time.Duration(rawDetails.ChallengeWindow.Uint64())

	var err error
	details.PayloadStr, err = protocol.GetProposalPayloadString(rp, rawDetails.Payload, opts)
	if err != nil {
		details.PayloadStr = fmt.Sprintf("<error decoding: %s>", err.Error())
	}
}
