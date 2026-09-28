package rocketpool

import (
	"math/big"
	"net/url"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func (c *Client) CanReleaseChallengeBond(id *big.Int) (api.CanSettleChallengeBondResponse, error) {
	return c.callAPI[api.CanSettleChallengeBondResponse]("GET", "/api/node/can-release-challenge-bond", url.Values{"challengeId": {id.String()}}, "Could not check challenge bond release")
}

func (c *Client) ReleaseChallengeBond(id *big.Int) (api.SettleChallengeBondResponse, error) {
	return c.callAPI[api.SettleChallengeBondResponse]("POST", "/api/node/release-challenge-bond", url.Values{"challengeId": {id.String()}}, "Could not release challenge bond")
}

func (c *Client) CanClaimChallengeReward(id *big.Int) (api.CanSettleChallengeBondResponse, error) {
	return c.callAPI[api.CanSettleChallengeBondResponse]("GET", "/api/node/can-claim-challenge-reward", url.Values{"challengeId": {id.String()}}, "Could not check challenge reward claim")
}

func (c *Client) ClaimChallengeReward(id *big.Int) (api.SettleChallengeBondResponse, error) {
	return c.callAPI[api.SettleChallengeBondResponse]("POST", "/api/node/claim-challenge-reward", url.Values{"challengeId": {id.String()}}, "Could not claim challenge reward")
}
