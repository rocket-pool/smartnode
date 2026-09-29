package rocketpool

import (
	"context"
	"math/big"
	"net/url"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func challengeMinipoolValues(addresses []common.Address, start uint64, participation []*big.Int) url.Values {
	values := challengePerformanceValues(common.Address{}, nil, start, participation)
	values.Del("validatorIds")
	members := make([]string, len(addresses))
	for i, address := range addresses {
		members[i] = address.Hex()
	}
	values.Set("minipoolAddresses", strings.Join(members, ","))
	return values
}

func (c *Client) CanChallengeMinipoolPerformance(addresses []common.Address, startEpoch uint64, participation []*big.Int) (api.CanChallengeMegapoolPerformanceResponse, error) {
	values := challengeMinipoolValues(addresses, startEpoch, participation)
	return c.callAPICtx[api.CanChallengeMegapoolPerformanceResponse](context.Background(), "GET", "/api/minipool/can-challenge-performance", values, "Could not get can challenge minipool performance status")
}

// ChallengeMinipoolPerformance submits a target-vote performance challenge
// against a same-node minipool list. This call has no client-side
// deadline because the transaction requires downloading a beacon state to
// build the slot proof.
func (c *Client) ChallengeMinipoolPerformance(addresses []common.Address, startEpoch uint64, participation []*big.Int, maxBond *big.Int) (api.ChallengeMegapoolPerformanceResponse, error) {
	values := challengeMinipoolValues(addresses, startEpoch, participation)
	values.Set("maxBond", maxBond.String())
	return c.callAPICtx[api.ChallengeMegapoolPerformanceResponse](context.Background(), "POST", "/api/minipool/challenge-performance", values, "Could not challenge minipool performance")
}
