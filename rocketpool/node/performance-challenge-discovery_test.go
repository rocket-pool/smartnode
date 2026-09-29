package node

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/megapool"
)

type challengeDiscoverySource struct {
	headers   map[uint64]*types.Header
	entries   []megapool.PerformanceChallenge
	statuses  map[string]megapool.PerformanceChallengeStatus
	ranges    [][2]*big.Int
	statusErr error
	eventsErr error
	statusAt  *big.Int
}

func (s *challengeDiscoverySource) header(block uint64) (*types.Header, error) {
	return s.headers[block], nil
}
func (s *challengeDiscoverySource) events(_ common.Address, from, to *big.Int) ([]megapool.PerformanceChallenge, error) {
	s.ranges = append(s.ranges, [2]*big.Int{from, to})
	return s.entries, s.eventsErr
}
func (s *challengeDiscoverySource) status(id, block *big.Int) (megapool.PerformanceChallengeStatus, error) {
	s.statusAt = block
	return s.statuses[id.String()], s.statusErr
}

func discoveryFixture() (*challengeDiscoverySource, common.Address) {
	address := common.HexToAddress("0x1234")
	source := &challengeDiscoverySource{
		headers: map[uint64]*types.Header{
			100: {Number: big.NewInt(100), Time: 10000},
			110: {Number: big.NewInt(110), Time: 11000},
		},
		statuses: map[string]megapool.PerformanceChallengeStatus{},
	}
	for _, id := range []int64{7, 8, 9} {
		source.entries = append(source.entries, megapool.PerformanceChallenge{
			ChallengeId: big.NewInt(id), MegapoolAddress: address, ValidatorIds: []uint32{2, 6},
			StartEpoch: 42, Participation: []*big.Int{big.NewInt(5)},
		})
		source.statuses[big.NewInt(id).String()] = megapool.PerformanceChallengeStatus{ResponseDeadline: big.NewInt(10001)}
	}
	return source, address
}

func TestDiscoverPerformanceChallenges(t *testing.T) {
	source, address := discoveryFixture()
	source.statuses["8"] = megapool.PerformanceChallengeStatus{Responded: true}
	source.statuses["9"] = megapool.PerformanceChallengeStatus{Finalised: true}
	// Duplicate events and another node's events must not create extra work.
	source.entries = append(source.entries, source.entries[0], megapool.PerformanceChallenge{
		ChallengeId: big.NewInt(10), MegapoolAddress: common.HexToAddress("0x5678"),
	})
	var discovery performanceChallengeDiscovery
	active, timestamp, err := discovery.discover(source, address, 100)
	if err != nil || len(active) != 1 || timestamp != 10000 {
		t.Fatalf("got %v, %d, %v", active, timestamp, err)
	}
	if active[0].challengeId.Int64() != 7 || len(active[0].validatorIds) != 2 || active[0].startEpoch != 42 || active[0].participationCallData[0].Int64() != 5 {
		t.Fatalf("challenge event data was not preserved: %+v", active[0])
	}
	if source.ranges[0][0] != nil || source.ranges[0][1].Int64() != 100 || source.statusAt.Int64() != 100 {
		t.Fatal("initial history and status must be read through the pinned block")
	}

	// No new events: an old challenge is still returned after its deadline.
	source.entries = nil
	active, _, err = discovery.discover(source, address, 110)
	if err != nil || len(active) != 1 || !active[0].expired(11000) || source.ranges[1][0].Int64() != 101 {
		t.Fatalf("old unresolved challenge was lost: %v, %v", active, err)
	}
	// Repeating the same block rechecks status without scanning an invalid range.
	source.statuses["7"] = megapool.PerformanceChallengeStatus{Finalised: true}
	active, _, err = discovery.discover(source, address, 110)
	if err != nil || len(active) != 0 || len(source.ranges) != 2 {
		t.Fatalf("finalised challenge still active or unchanged block rescanned: %v, %v", active, err)
	}
}

func TestDiscoveryRetriesFailedScansAndRebuildsAfterReorg(t *testing.T) {
	source, address := discoveryFixture()
	var discovery performanceChallengeDiscovery
	source.eventsErr = errors.New("log RPC failed")
	if _, _, err := discovery.discover(source, address, 100); err == nil || discovery.checkpoint != nil {
		t.Fatal("failed initial scan advanced cursor")
	}
	source.eventsErr = nil
	if _, _, err := discovery.discover(source, address, 100); err != nil {
		t.Fatal(err)
	}
	source.statusErr = errors.New("status RPC failed")
	if _, _, err := discovery.discover(source, address, 110); err == nil || discovery.checkpoint.Number.Int64() != 100 {
		t.Fatal("failed status lookup advanced cursor")
	}
	source.statusErr = nil
	if _, _, err := discovery.discover(source, address, 110); err != nil {
		t.Fatal(err)
	}
	if source.ranges[len(source.ranges)-1][0].Int64() != 101 {
		t.Fatal("retry skipped unprocessed blocks")
	}
	// Orphaned challenges disappear, and previously resolved challenges can be
	// rediscovered from the replacement chain instead of trusting the old cache.
	source.headers[110] = &types.Header{Number: big.NewInt(110), Time: 11001, Extra: []byte("replacement")}
	source.entries = source.entries[1:2]
	active, _, err := discovery.discover(source, address, 110)
	if err != nil || len(active) != 1 || active[0].challengeId.Int64() != 8 || source.ranges[len(source.ranges)-1][0] != nil {
		t.Fatalf("reorg did not rebuild discovery: %v, %v", active, err)
	}
}

func TestDiscoveryWithoutChallenges(t *testing.T) {
	source, address := discoveryFixture()
	source.entries = nil
	var discovery performanceChallengeDiscovery
	active, _, err := discovery.discover(source, address, 100)
	if err != nil || len(active) != 0 {
		t.Fatalf("empty history generated challenge work: %v, %v", active, err)
	}
}

func TestChallengeDeadlineBoundary(t *testing.T) {
	challenge := megapoolPerformanceChallenge{responseDeadline: big.NewInt(100)}
	for _, timestamp := range []uint64{99, 100, 101} {
		if challenge.expired(timestamp) != (timestamp > 100) {
			t.Errorf("wrong finalisation eligibility at %d", timestamp)
		}
	}
}

func TestDiscoverMinipoolsWithoutMegapool(t *testing.T) {
	source, _ := discoveryFixture()
	owner, member := common.Address{1}, common.Address{2}
	source.entries = []megapool.PerformanceChallenge{
		{
			ChallengeId:       big.NewInt(7),
			NodeAddress:       owner,
			MinipoolAddresses: []common.Address{member},
			StartEpoch:        42,
			Participation:     []*big.Int{big.NewInt(5)},
		},
	}
	var discovery performanceChallengeDiscovery
	got, _, err := discovery.discover(source, common.Address{}, 100)
	if err != nil || len(got) != 1 {
		t.Fatalf("minipool discovery failed: %v, %v", got, err)
	}
	binding := got[0].binding()
	if binding.NodeAddress != owner || len(binding.MinipoolAddresses) != 1 || binding.MinipoolAddresses[0] != member {
		t.Fatalf("lost minipool membership: %+v", binding)
	}
}

func TestPerformanceChallengeTaskScope(t *testing.T) {
	for _, test := range []struct {
		name                                   string
		own, proposed, enforcer, expired, want bool
	}{
		{
			name:     "always defend own validators",
			own:      true,
			proposed: false,
			enforcer: false,
			expired:  false,
			want:     true,
		},
		{
			name:     "do not defend own proposal",
			own:      true,
			proposed: true,
			enforcer: true,
			expired:  false,
			want:     false,
		},
		{
			name:     "third-party defense opt out",
			own:      false,
			proposed: false,
			enforcer: false,
			expired:  false,
			want:     false,
		},
		{
			name:     "third-party defense opt in",
			own:      false,
			proposed: false,
			enforcer: true,
			expired:  false,
			want:     true,
		},
		{
			name:     "finalise own proposal",
			own:      false,
			proposed: true,
			enforcer: false,
			expired:  true,
			want:     true,
		},
		{
			name:     "finalise own validators",
			own:      true,
			proposed: false,
			enforcer: false,
			expired:  true,
			want:     true,
		},
		{
			name:     "third-party finalisation opt out",
			own:      false,
			proposed: false,
			enforcer: false,
			expired:  true,
			want:     false,
		},
		{
			name:     "third-party finalisation opt in",
			own:      false,
			proposed: false,
			enforcer: true,
			expired:  true,
			want:     true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldHandlePerformanceChallenge(test.own, test.proposed, test.enforcer, test.expired); got != test.want {
				t.Fatalf("got %t, want %t", got, test.want)
			}
		})
	}
}
