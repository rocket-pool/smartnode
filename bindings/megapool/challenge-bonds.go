package megapool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/logs"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// GetChallengeBondCandidates finds bonds proposed by this node and defeated
// challenges of either pool type. Defeat events do not identify the responder;
// callers must filter these candidates using getChallengeBondDetails.
func GetChallengeBondCandidates(rp *rocketpool.RocketPool, node common.Address, interval, fromBlock, toBlock *big.Int, opts *bind.CallOpts) ([]*big.Int, error) {
	contract, err := getRocketNetworkParticipation(rp, opts)
	if err != nil {
		return nil, err
	}
	addresses, err := participationContractAddresses(rp, contract, interval, fromBlock, toBlock, opts)
	if err != nil {
		return nil, err
	}
	locked, ok := contract.ABI.Events["ChallengeBondLocked"]
	if !ok {
		return nil, fmt.Errorf("ChallengeBondLocked event missing from participation ABI")
	}
	megapoolDefeated, ok := contract.ABI.Events["MegapoolChallengeDefeated"]
	if !ok {
		return nil, fmt.Errorf("MegapoolChallengeDefeated event missing from participation ABI")
	}
	minipoolDefeated, ok := contract.ABI.Events["MinipoolChallengeDefeated"]
	if !ok {
		return nil, fmt.Errorf("MinipoolChallengeDefeated event missing from participation ABI")
	}
	lockedLogs, err := logs.GetLogs(rp, addresses, [][]common.Hash{{locked.ID}, nil, {common.BytesToHash(node.Bytes())}}, interval, fromBlock, toBlock, nil)
	if err != nil {
		return nil, err
	}
	defeatedLogs, err := logs.GetLogs(rp, addresses, [][]common.Hash{{megapoolDefeated.ID, minipoolDefeated.ID}}, interval, fromBlock, toBlock, nil)
	if err != nil {
		return nil, err
	}
	ids := []*big.Int{}
	seen := map[string]bool{}
	add := func(id *big.Int) {
		if !seen[id.String()] {
			ids = append(ids, id)
			seen[id.String()] = true
		}
	}
	for _, entry := range lockedLogs {
		if entry.Removed {
			continue
		}
		if len(entry.Topics) != 3 || entry.Topics[0] != locked.ID {
			return nil, fmt.Errorf("invalid ChallengeBondLocked event")
		}
		add(entry.Topics[1].Big())
	}
	for _, entry := range defeatedLogs {
		if entry.Removed {
			continue
		}
		if len(entry.Topics) != 1 {
			return nil, fmt.Errorf("invalid challenge defeat event")
		}
		event := megapoolDefeated
		if entry.Topics[0] == minipoolDefeated.ID {
			event = minipoolDefeated
		} else if entry.Topics[0] != megapoolDefeated.ID {
			return nil, fmt.Errorf("unexpected challenge defeat event")
		}
		values, err := event.Inputs.NonIndexed().Unpack(entry.Data)
		if err != nil || len(values) != 1 {
			return nil, fmt.Errorf("invalid challenge defeat data: %v", err)
		}
		add(values[0].(*big.Int))
	}
	return ids, nil
}
