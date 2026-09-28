package megapool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rocket-pool/smartnode/bindings/logs"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// PerformanceChallenge contains the data needed to reconstruct a challenge proof.
type PerformanceChallenge struct {
	ChallengeId     *big.Int
	MegapoolAddress common.Address
	ValidatorIds    []uint32
	StartEpoch      uint64
	Participation   []*big.Int
}

type PerformanceChallengeStatus struct {
	Proposer         common.Address
	ResponseDeadline *big.Int
	Responded        bool
	Finalised        bool
}

// GetMegapoolPerformanceChallenges reads challenge events, including events from
// previous deployments of rocketNetworkParticipation. A nil fromBlock scans
// from Rocket Pool's deployment, so overdue unresolved challenges are included.
func GetMegapoolPerformanceChallenges(rp *rocketpool.RocketPool, address common.Address, interval, fromBlock, toBlock *big.Int, opts *bind.CallOpts) ([]PerformanceChallenge, error) {
	contract, err := getRocketNetworkParticipation(rp, opts)
	if err != nil {
		return nil, err
	}
	event, ok := contract.ABI.Events["MegapoolChallenged"]
	if !ok {
		return nil, fmt.Errorf("MegapoolChallenged event not found in rocketNetworkParticipation ABI")
	}
	// Include every address active within this range. Each upgrade identifies
	// the old address; the current contract supplies the last address. Limiting
	// upgrade discovery to the same range keeps subsequent scans incremental.
	upgrade, err := rp.GetContract("rocketDAONodeTrustedUpgrade", opts)
	if err != nil {
		return nil, err
	}
	upgradeEvent, ok := upgrade.ABI.Events["ContractUpgraded"]
	if !ok {
		return nil, fmt.Errorf("ContractUpgraded event not found in rocketDAONodeTrustedUpgrade ABI")
	}
	upgrades, err := logs.GetLogs(rp, []common.Address{*upgrade.Address}, [][]common.Hash{
		{upgradeEvent.ID}, {crypto.Keccak256Hash([]byte("rocketNetworkParticipation"))},
	}, interval, fromBlock, toBlock, nil)
	if err != nil {
		return nil, err
	}
	addresses := []common.Address{*contract.Address}
	seen := map[common.Address]bool{*contract.Address: true}
	for _, entry := range upgrades {
		if len(entry.Topics) != 4 {
			return nil, fmt.Errorf("invalid ContractUpgraded event topics")
		}
		old := common.BytesToAddress(entry.Topics[2].Bytes())
		if !seen[old] {
			addresses = append(addresses, old)
			seen[old] = true
		}
	}
	entries, err := logs.GetLogs(rp, addresses, [][]common.Hash{
		{event.ID}, {common.BytesToHash(address.Bytes())},
	}, interval, fromBlock, toBlock, nil)
	if err != nil {
		return nil, err
	}
	challenges := make([]PerformanceChallenge, 0, len(entries))
	for _, entry := range entries {
		if entry.Removed {
			continue
		}
		challenge, err := decodePerformanceChallenge(event, entry)
		if err != nil {
			return nil, err
		}
		challenges = append(challenges, challenge)
	}
	return challenges, nil
}

func decodePerformanceChallenge(event abi.Event, entry types.Log) (PerformanceChallenge, error) {
	if len(entry.Topics) != 3 || entry.Topics[0] != event.ID {
		return PerformanceChallenge{}, fmt.Errorf("invalid MegapoolChallenged event topics")
	}
	values, err := event.Inputs.NonIndexed().Unpack(entry.Data)
	if err != nil {
		return PerformanceChallenge{}, fmt.Errorf("error unpacking MegapoolChallenged: %w", err)
	}
	if len(values) != 4 {
		return PerformanceChallenge{}, fmt.Errorf("invalid MegapoolChallenged event data")
	}
	ids, idsOK := values[0].([]uint32)
	start, startOK := values[1].(uint64)
	participation, participationOK := values[3].([]*big.Int)
	if !idsOK || !startOK || !participationOK || len(ids) == 0 || len(participation) == 0 {
		return PerformanceChallenge{}, fmt.Errorf("invalid MegapoolChallenged validator list or participation bitmap")
	}
	return PerformanceChallenge{
		ChallengeId: entry.Topics[2].Big(), MegapoolAddress: common.BytesToAddress(entry.Topics[1].Bytes()),
		ValidatorIds: ids, StartEpoch: start, Participation: participation,
	}, nil
}

// GetPerformanceChallengeStatus reads the snapshotted deadline and the two
// terminal flags. Bond settlement is independent of challenge finalisation.
func GetPerformanceChallengeStatus(rp *rocketpool.RocketPool, challengeId *big.Int, opts *bind.CallOpts) (PerformanceChallengeStatus, error) {
	contract, err := getRocketNetworkParticipation(rp, opts)
	if err != nil {
		return PerformanceChallengeStatus{}, err
	}
	var values []interface{}
	if err := contract.Contract.Call(opts, &values, "getChallengeBondDetails", challengeId); err != nil {
		return PerformanceChallengeStatus{}, fmt.Errorf("error getting challenge %s bond details: %w", challengeId, err)
	}
	if len(values) != 5 {
		return PerformanceChallengeStatus{}, fmt.Errorf("invalid challenge bond details")
	}
	status := PerformanceChallengeStatus{Proposer: values[0].(common.Address), ResponseDeadline: values[3].(*big.Int)}
	for key, target := range map[string]*bool{
		"participation.challenge.responded": &status.Responded,
		"participation.challenge.finalised": &status.Finalised,
	} {
		// Solidity uses abi.encodePacked(string, uint256) for these storage keys.
		hash := crypto.Keccak256Hash([]byte(key), common.LeftPadBytes(challengeId.Bytes(), 32))
		value, err := rp.RocketStorage.GetBool(opts, hash)
		if err != nil {
			return PerformanceChallengeStatus{}, fmt.Errorf("error getting challenge %s status: %w", challengeId, err)
		}
		*target = value
	}
	return status, nil
}
