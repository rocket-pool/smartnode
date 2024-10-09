package rewards

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/rewards/test"
	"github.com/rocket-pool/smartnode/shared/services/rewards/test/assets"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

type v8Test struct {
	*testing.T
	rp *test.MockRocketPool
	bc *test.MockBeaconClient
}

func (t *v8Test) saveArtifacts(prefix string, result *GenerateTreeResult) {
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("artifacts-%s", t.Name()))
	t.failIf(err)
	rewardsLocalFile := LocalFile[IRewardsFile]{
		fullPath: filepath.Join(tmpDir, fmt.Sprintf("%s-rewards.json", prefix)),
		f:        result.RewardsFile,
	}
	performanceLocalFile := LocalFile[IMinipoolPerformanceFile]{
		fullPath: filepath.Join(tmpDir, fmt.Sprintf("%s-minipool-performance.json", prefix)),
		f:        result.MinipoolPerformanceFile,
	}
	_, err = rewardsLocalFile.Write()
	t.failIf(err)
	_, err = performanceLocalFile.Write()
	t.failIf(err)

	t.Logf("wrote artifacts to %s\n", tmpDir)
}

func newV8Test(t *testing.T) *v8Test {
	rp := test.NewMockRocketPool(t)
	out := &v8Test{
		T:  t,
		rp: rp,
		bc: test.NewMockBeaconClient(t),
	}
	return out
}

func (t *v8Test) failIf(err error) {
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func (t *v8Test) SetMinipoolPerformance(canonicalMinipoolPerformance IMinipoolPerformanceFile, networkState *state.NetworkState) {
	addresses := canonicalMinipoolPerformance.GetMinipoolAddresses()
	for _, address := range addresses {

		// Get the minipool's performance
		perf, ok := canonicalMinipoolPerformance.GetSmoothingPoolPerformance(address)
		if !ok {
			t.Fatalf("Minipool %s not found in canonical minipool performance, despite being listed as present", address.Hex())
		}
		missedSlots := perf.GetMissingAttestationSlots()
		pubkey, err := perf.GetPubkey()

		// Get the minipool's validator index
		validatorStatus := networkState.ValidatorDetails[pubkey]

		if err != nil {
			t.Fatalf("Minipool %s pubkey could not be parsed: %s", address.Hex(), err.Error())
		}
		t.bc.SetMinipoolPerformance(validatorStatus.Index, missedSlots)
	}
}

// TestV8Mainnet builds a tree using serialized state for a mainnet interval that used v8
// and checks that the resulting artifacts match their canonical values.
func TestV8Mainnet(tt *testing.T) {
	t := newV8Test(tt)

	canonical, err := DeserializeRewardsFile(assets.GetMainnet20RewardsJSON())
	t.failIf(err)

	canonicalPerformance, err := DeserializeMinipoolPerformanceFile(assets.GetMainnet20MinipoolPerformanceJSON())
	t.failIf(err)

	state := assets.GetMainnet20RewardsState()
	t.Logf("pending rpl rewards: %s", state.NetworkDetails.PendingRPLRewards.String())

	t.bc.SetState(state)

	// Some interval info needed for mocks
	consensusStartBlock := canonical.GetConsensusStartBlock()
	executionStartBlock := canonical.GetExecutionStartBlock()
	consensusEndBlock := canonical.GetConsensusEndBlock()

	// Create a new treeGeneratorImpl_v8
	logger := log.NewColorLogger(color.Faint)
	generator := newTreeGeneratorImpl_v8(
		&logger,
		t.Name(),
		state.NetworkDetails.RewardIndex,
		canonical.GetStartTime(),
		canonical.GetEndTime(),
		consensusEndBlock,
		&types.Header{
			Number: big.NewInt(int64(canonical.GetExecutionEndBlock())),
			Time:   assets.Mainnet20ELHeaderTime,
		},
		canonical.GetIntervalsPassed(),
		state,
	)

	// Load the mock up
	t.rp.SetRewardSnapshotEvent(assets.GetRewardSnapshotEventInterval19())
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock-1), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock - 1})
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock})
	t.rp.SetHeaderByNumber(big.NewInt(int64(executionStartBlock)), &types.Header{Time: uint64(canonical.GetStartTime().Unix())})

	// Set the critical duties slots
	t.bc.SetCriticalDutiesSlots(assets.GetMainnet20CriticalDutiesSlots())

	// Set the minipool performance
	t.SetMinipoolPerformance(canonicalPerformance, state)

	artifacts, err := generator.generateTree(
		t.rp,
		"mainnet",
		make([]common.Address, 0),
		t.bc,
	)
	t.failIf(err)

	// Save the artifacts if verbose mode is enabled
	if testing.Verbose() {
		t.saveArtifacts("", artifacts)
	}

	t.Logf("merkle root: %s\n", artifacts.RewardsFile.GetMerkleRoot())
	if artifacts.RewardsFile.GetMerkleRoot() != canonical.GetMerkleRoot() {
		t.Fatalf("Merkle root does not match %s", canonical.GetMerkleRoot())
	} else {
		t.Logf("merkle root matches %s", canonical.GetMerkleRoot())
	}
}
