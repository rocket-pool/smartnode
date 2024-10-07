package test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/rocketpool-go/rewards"
)

// MockRocketPool is a EC mock specifically for testing treegen.
// At a high level our approach is to provide two options to the tester:
// 1) Use a recording of request/response data from production to emulate a canonical tree
// 2) Allow for full response customization.
//
// The former is useful for ensuring that no regressions arise during refactors that should
// otherwise be nonfunction, ie, not impact the merkle root.
//
// The latter is useful to probe specific behaviors such as opt-in/opt-out eligibility timing,
// node weight, smoothing pool status, etc.
//
// Because recording responses ties the test to a specific version of the contracts and therefor
// the client-side code, the interface we need to mock should be as minimized as possible, and the
// recorded data should tightly match that interface. That is, no recorded response should encode
// something like the contract address data are being requested from, but instead the high-level
// function name and arguments.
type MockRocketPool struct {
	t                    *testing.T
	rewardSnapshotEvents map[uint64]rewards.RewardsEvent
	headers              map[uint64]*types.Header
}

func NewMockRocketPool(t *testing.T) *MockRocketPool {
	return &MockRocketPool{t: t}
}

func (mock *MockRocketPool) GetNetworkEnabled(networkId *big.Int, opts *bind.CallOpts) (bool, error) {
	mock.t.Logf("GetNetworkEnabled(%+v, %+v)", networkId, opts)
	return true, nil
}

func (mock *MockRocketPool) HeaderByNumber(_ context.Context, number *big.Int) (*types.Header, error) {
	mock.t.Logf("HeaderByNumber(%+v)", number)
	if header, ok := mock.headers[number.Uint64()]; ok {
		return header, nil
	}
	return nil, fmt.Errorf("header not found in mock for %d, please set it with SetHeaderByNumber", number.Uint64())
}

func (mock *MockRocketPool) SetHeaderByNumber(number *big.Int, header *types.Header) {
	if mock.headers == nil {
		mock.headers = make(map[uint64]*types.Header)
	}
	mock.headers[number.Uint64()] = header
}

func (mock *MockRocketPool) GetRewardsEvent(index uint64, _ []common.Address, opts *bind.CallOpts) (bool, rewards.RewardsEvent, error) {
	mock.t.Logf("GetRewardsEvent(%+v, %+v)", index, opts)

	if event, ok := mock.rewardSnapshotEvents[index]; ok {
		return true, event, nil
	}
	return false, rewards.RewardsEvent{}, nil
}

func (mock *MockRocketPool) GetRewardSnapshotEvent(previousRewardsPoolAddresses []common.Address, interval uint64, opts *bind.CallOpts) (rewards.RewardsEvent, error) {
	mock.t.Logf("GetRewardSnapshotEvent(%+v, %+v, %+v)", previousRewardsPoolAddresses, interval, opts)
	if event, ok := mock.rewardSnapshotEvents[interval]; ok {
		return event, nil
	}
	return rewards.RewardsEvent{}, nil
}

func (mock *MockRocketPool) SetRewardSnapshotEvent(event rewards.RewardsEvent) {
	if mock.rewardSnapshotEvents == nil {
		mock.rewardSnapshotEvents = make(map[uint64]rewards.RewardsEvent)
	}
	mock.rewardSnapshotEvents[event.Index.Uint64()] = event
}
