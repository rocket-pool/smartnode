package rewards

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/rewards"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/trustednode"
)

// Interface assertion
var _ RewardsExecutionClient = &defaultRewardsExecutionClient{}

// An implementation of RewardsExecutionClient that uses
// rocketpool-go to access chain data.
//
// Importantly, this struct instantiates rocketpool.RocketPool and passes it
// to the old fashioned rocketpool-go getters that take it as an argument
// but it also fulfills the requirements of an interface used for dependency injection
// in tests.
type defaultRewardsExecutionClient struct {
	*rocketpool.RocketPool
}

func NewRewardsExecutionClient(rp *rocketpool.RocketPool) *defaultRewardsExecutionClient {
	out := new(defaultRewardsExecutionClient)
	out.RocketPool = rp
	return out
}

func (client *defaultRewardsExecutionClient) GetNetworkEnabled(networkId *big.Int, opts *bind.CallOpts) (bool, error) {
	return trustednode.GetNetworkEnabled(client.RocketPool, networkId, opts)
}

func (client *defaultRewardsExecutionClient) HeaderByNumber(ctx context.Context, block *big.Int) (*ethtypes.Header, error) {
	return client.RocketPool.Client.HeaderByNumber(ctx, block)
}

func (client *defaultRewardsExecutionClient) GetRewardsEvent(index uint64, rocketRewardsPoolAddresses []common.Address, opts *bind.CallOpts) (bool, rewards.RewardsEvent, error) {
	return rewards.GetRewardsEvent(client.RocketPool, index, rocketRewardsPoolAddresses, opts)
}

func (client *defaultRewardsExecutionClient) GetRewardSnapshotEvent(previousRewardsPoolAddresses []common.Address, interval uint64, opts *bind.CallOpts) (rewards.RewardsEvent, error) {

	found, event, err := client.GetRewardsEvent(interval, previousRewardsPoolAddresses, opts)
	if err != nil {
		return rewards.RewardsEvent{}, fmt.Errorf("error getting rewards event for interval %d: %w", interval, err)
	}
	if !found {
		return rewards.RewardsEvent{}, fmt.Errorf("interval %d event not found", interval)
	}

	return event, nil

}

func (client *defaultRewardsExecutionClient) GetRewardIndex(opts *bind.CallOpts) (*big.Int, error) {
	return client.RocketPool.GetRewardIndex(opts)
}

func (client *defaultRewardsExecutionClient) GetContract(contractName string, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	return client.RocketPool.GetContract(contractName, opts)
}

func (client *defaultRewardsExecutionClient) BalanceAt(ctx context.Context, address common.Address, blockNumber *big.Int) (*big.Int, error) {
	return client.RocketPool.Client.BalanceAt(ctx, address, blockNumber)
}

func (client *defaultRewardsExecutionClient) Client() *rocketpool.RocketPool {
	return client.RocketPool
}
