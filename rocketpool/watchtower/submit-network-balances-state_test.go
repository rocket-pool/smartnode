package watchtower

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

const smallStateFixture = "../../shared/services/state/testdata/network_state.json.gz"

// stubRewardSplitCalculator returns a deterministic 50/50 split between
// rETH and node rewards, which is sufficient to verify plumbing.
type stubRewardSplitCalculator struct {
	calls []rewardSplitCall
}
type rewardSplitCall struct {
	MegapoolAddress common.Address
	Rewards         *big.Int
}

func (s *stubRewardSplitCalculator) CalculateRewards(megapoolAddress common.Address, rewards *big.Int, _ uint64) (megapool.RewardSplit, error) {
	s.calls = append(s.calls, rewardSplitCall{megapoolAddress, new(big.Int).Set(rewards)})
	half := new(big.Int).Div(rewards, big.NewInt(2))
	remainder := new(big.Int).Sub(rewards, half)
	return megapool.RewardSplit{
		NodeRewards:        half,
		VoterRewards:       big.NewInt(0),
		RethRewards:        remainder,
		ProtocolDAORewards: big.NewInt(0),
	}, nil
}

// stubSmoothingPoolCalculator returns the full smoothing pool balance as the
// rETH share so the test value is deterministic and easy to verify.
type stubSmoothingPoolCalculator struct{}

func (s *stubSmoothingPoolCalculator) GetSmoothingPoolShare(ns *state.NetworkState, _ *types.Header, _ time.Time) (*big.Int, error) {
	return ns.NetworkDetails.SmoothingPoolBalance, nil
}

func TestGetNetworkBalancesFromState(t *testing.T) {
	provider, err := state.NewStaticNetworkStateProviderFromFile(smallStateFixture)
	if err != nil {
		t.Fatalf("loading state: %v", err)
	}

	ns, err := provider.GetHeadState()
	if err != nil {
		t.Fatalf("GetHeadState: %v", err)
	}

	logger := log.NewColorLogger(0)
	cfg := config.NewRocketPoolConfig("", false)
	rewardCalc := &stubRewardSplitCalculator{}
	spCalc := &stubSmoothingPoolCalculator{}

	elBlockHeader := &types.Header{
		Number: new(big.Int).SetUint64(ns.ElBlockNumber),
	}
	slotTime := time.Unix(int64(ns.BeaconConfig.GenesisTime+ns.BeaconSlotNumber*ns.BeaconConfig.SecondsPerSlot), 0)

	task := &submitNetworkBalances{
		log: &logger,
		cfg: cfg,
	}

	balances, err := task.getNetworkBalancesFromState(ns, elBlockHeader, slotTime, rewardCalc, spCalc)
	if err != nil {
		t.Fatalf("getNetworkBalancesFromState: %v", err)
	}

	if balances.Block != ns.ElBlockNumber {
		t.Errorf("Block: got %d, want %d", balances.Block, ns.ElBlockNumber)
	}

	// DepositPool must equal DepositPoolUserBalance from the state
	if balances.DepositPool.Cmp(ns.NetworkDetails.DepositPoolUserBalance) != 0 {
		t.Errorf("DepositPool: got %s, want %s", balances.DepositPool, ns.NetworkDetails.DepositPoolUserBalance)
	}

	// RETHContract must equal RETHBalance from the state
	if balances.RETHContract.Cmp(ns.NetworkDetails.RETHBalance) != 0 {
		t.Errorf("RETHContract: got %s, want %s", balances.RETHContract, ns.NetworkDetails.RETHBalance)
	}

	// RETHSupply must equal TotalRETHSupply from the state
	if balances.RETHSupply.Cmp(ns.NetworkDetails.TotalRETHSupply) != 0 {
		t.Errorf("RETHSupply: got %s, want %s", balances.RETHSupply, ns.NetworkDetails.TotalRETHSupply)
	}

	// SmoothingPoolShare must equal SmoothingPoolBalance (per our stub)
	if balances.SmoothingPoolShare.Cmp(ns.NetworkDetails.SmoothingPoolBalance) != 0 {
		t.Errorf("SmoothingPoolShare: got %s, want %s", balances.SmoothingPoolShare, ns.NetworkDetails.SmoothingPoolBalance)
	}

	// Verify minipool totals are non-negative
	if balances.MinipoolsTotal.Sign() < 0 {
		t.Errorf("MinipoolsTotal is negative: %s", balances.MinipoolsTotal)
	}
	if balances.MinipoolsStaking.Sign() < 0 {
		t.Errorf("MinipoolsStaking is negative: %s", balances.MinipoolsStaking)
	}
	// MinipoolsStaking must not exceed MinipoolsTotal
	if balances.MinipoolsStaking.Cmp(balances.MinipoolsTotal) > 0 {
		t.Errorf("MinipoolsStaking (%s) > MinipoolsTotal (%s)", balances.MinipoolsStaking, balances.MinipoolsTotal)
	}

	// The fixture has 10 minipools; verify the total is non-zero for staking validators
	if len(ns.MinipoolDetails) > 0 && balances.MinipoolsTotal.Sign() == 0 {
		t.Error("MinipoolsTotal is zero but the fixture has minipools")
	}

	// NodeCreditBalance must be the sum of all nodes' DepositCreditBalance
	expectedCredit := big.NewInt(0)
	for _, node := range ns.NodeDetails {
		expectedCredit.Add(expectedCredit, node.DepositCreditBalance)
	}
	if balances.NodeCreditBalance.Cmp(expectedCredit) != 0 {
		t.Errorf("NodeCreditBalance: got %s, want %s", balances.NodeCreditBalance, expectedCredit)
	}

	// DistributorShareTotal must be the sum of all nodes' DistributorBalanceUserETH
	expectedDistributor := big.NewInt(0)
	for _, node := range ns.NodeDetails {
		expectedDistributor.Add(expectedDistributor, node.DistributorBalanceUserETH)
	}
	if balances.DistributorShareTotal.Cmp(expectedDistributor) != 0 {
		t.Errorf("DistributorShareTotal: got %s, want %s", balances.DistributorShareTotal, expectedDistributor)
	}

	// Verify megapool fields are non-negative
	if balances.MegapoolsUserShareTotal.Sign() < 0 {
		t.Errorf("MegapoolsUserShareTotal is negative: %s", balances.MegapoolsUserShareTotal)
	}
	if balances.MegapoolStaking.Sign() < 0 {
		t.Errorf("MegapoolStaking is negative: %s", balances.MegapoolStaking)
	}

	// The fixture has megapool details; verify they were loaded
	if len(ns.MegapoolDetails) == 0 {
		t.Fatal("MegapoolDetails is empty — fixture may not have been regenerated with the megapool_details json tag")
	}

	// MegapoolsUserShareTotal should equal the sum of UserCapital for megapools
	// that appear in MegapoolDetails AND have validators in MegapoolToPubkeysMap
	// (the balance loop iterates MegapoolDetails, which is keyed by address).
	expectedUserCapital := big.NewInt(0)
	for addr, mp := range ns.MegapoolDetails {
		if _, hasPubkeys := ns.MegapoolToPubkeysMap[addr]; hasPubkeys {
			expectedUserCapital.Add(expectedUserCapital, mp.UserCapital)
		}
	}
	if balances.MegapoolsUserShareTotal.Cmp(expectedUserCapital) != 0 {
		t.Errorf("MegapoolsUserShareTotal: got %s, want %s (sum of UserCapital for megapools with validators)", balances.MegapoolsUserShareTotal, expectedUserCapital)
	}

	t.Logf("Balances summary:")
	t.Logf("  Block:                   %d", balances.Block)
	t.Logf("  DepositPool:             %s", balances.DepositPool)
	t.Logf("  MinipoolsTotal:          %s", balances.MinipoolsTotal)
	t.Logf("  MinipoolsStaking:        %s", balances.MinipoolsStaking)
	t.Logf("  MegapoolsUserShareTotal: %s", balances.MegapoolsUserShareTotal)
	t.Logf("  MegapoolStaking:         %s", balances.MegapoolStaking)
	t.Logf("  DistributorShareTotal:   %s", balances.DistributorShareTotal)
	t.Logf("  SmoothingPoolShare:      %s", balances.SmoothingPoolShare)
	t.Logf("  RETHContract:            %s", balances.RETHContract)
	t.Logf("  RETHSupply:              %s", balances.RETHSupply)
	t.Logf("  NodeCreditBalance:       %s", balances.NodeCreditBalance)
	t.Logf("  RewardCalc calls:        %d", len(rewardCalc.calls))
}
