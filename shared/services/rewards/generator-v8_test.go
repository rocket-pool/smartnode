package rewards

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rocket-pool/smartnode/shared/services/rewards/test"
	"github.com/rocket-pool/smartnode/shared/services/state"
)

type RewardsTest struct {
	*testing.T
	rp *test.MockRocketPool
	bc *test.MockBeaconClient
}

func (t *RewardsTest) saveArtifacts(prefix string, result *GenerateTreeResult) {
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("artifacts-%s", t.Name()))
	t.failIf(err)
	rewardsLocalFile := LocalFile[IRewardsFile]{
		fullPath: filepath.Join(tmpDir, fmt.Sprintf("%s-rewards.json", prefix)),
		f:        result.RewardsFile,
	}
	performanceLocalFile := LocalFile[IPerformanceFile]{
		fullPath: filepath.Join(tmpDir, fmt.Sprintf("%s-performance.json", prefix)),
		f:        result.MinipoolPerformanceFile,
	}
	_, err = rewardsLocalFile.Write()
	t.failIf(err)
	_, err = performanceLocalFile.Write()
	t.failIf(err)

	t.Logf("wrote artifacts to %s\n", tmpDir)
}

func newRewardsTest(t *testing.T, index uint64) *RewardsTest {
	rp := test.NewMockRocketPool(t, index)
	out := &RewardsTest{
		T:  t,
		rp: rp,
		bc: test.NewMockBeaconClient(t),
	}
	return out
}

func (t *RewardsTest) failIf(err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func (t *RewardsTest) SetMinipoolPerformance(canonicalMinipoolPerformance IPerformanceFile, networkState *state.NetworkState) {
	addresses := canonicalMinipoolPerformance.GetMinipoolAddresses()
	for _, address := range addresses {

		// Get the minipool's performance
		perf, ok := canonicalMinipoolPerformance.GetMinipoolPerformance(address)
		if !ok {
			t.Fatalf("Minipool %s not found in canonical minipool performance, despite being listed as present", address.Hex())
		}
		missedSlots := perf.GetMissingAttestationSlots()
		pubkey, err := perf.GetPubkey()

		// Get the minipool's validator index
		validatorStatus := networkState.MinipoolValidatorDetails[pubkey]

		if err != nil {
			t.Fatalf("Minipool %s pubkey could not be parsed: %s", address.Hex(), err.Error())
		}
		t.bc.SetMinipoolPerformance(validatorStatus.Index, missedSlots)
	}
}
