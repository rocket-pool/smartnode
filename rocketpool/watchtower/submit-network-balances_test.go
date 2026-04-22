package watchtower

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/state"
)

// ============================================================
// Helpers
// ============================================================

func ethToWei(eth float64) *big.Int {
	return new(big.Int).Mul(
		big.NewInt(int64(eth*1e9)),
		big.NewInt(1e9),
	)
}

func newNetworkBalances() networkBalances {
	return networkBalances{
		DepositPool:             big.NewInt(0),
		MinipoolsTotal:          big.NewInt(0),
		MinipoolsStaking:        big.NewInt(0),
		MegapoolsUserShareTotal: big.NewInt(0),
		MegapoolStaking:         big.NewInt(0),
		DistributorShareTotal:   big.NewInt(0),
		SmoothingPoolShare:      big.NewInt(0),
		RETHContract:            big.NewInt(0),
		RETHSupply:              big.NewInt(0),
		NodeCreditBalance:       big.NewInt(0),
	}
}

var testGenesisTime = time.Date(2020, 12, 1, 12, 0, 0, 0, time.UTC)

func makeEth2Config(secondsPerSlot, slotsPerEpoch uint64) beacon.Eth2Config {
	return beacon.Eth2Config{
		GenesisTime:    uint64(testGenesisTime.Unix()),
		SecondsPerSlot: secondsPerSlot,
		SlotsPerEpoch:  slotsPerEpoch,
	}
}

// makeParentBeaconRoot returns a deterministic hash used as ParentBeaconRoot
// in EL headers so the stub's blocksByID map can be keyed on its hex string.
func makeParentBeaconRoot(slot uint64) common.Hash {
	return common.BigToHash(big.NewInt(int64(slot)))
}

// ============================================================
// applyMaxRethDelta
// ============================================================

func TestApplyMaxRethDelta_NoClampNeeded(t *testing.T) {
	b := newNetworkBalances()
	b.OriginalTotalBalanceWei = ethToWei(1000)
	b.OriginalRatioWei = eth.EthToWei(1.05)

	// Last rate is 1.04, new is 1.05: delta = 0.01
	lastRate := 1.04
	// Max delta is 0.02 ETH in wei — larger than actual change, so no clamp
	maxDelta := eth.EthToWei(0.02)

	b.applyMaxRethDelta(maxDelta, lastRate)

	if b.ClampedTotalBalanceWei.Cmp(b.OriginalTotalBalanceWei) != 0 {
		t.Errorf("expected no clamping: got %s, want %s", b.ClampedTotalBalanceWei, b.OriginalTotalBalanceWei)
	}
	if b.ClampedRatioWei.Cmp(b.OriginalRatioWei) != 0 {
		t.Errorf("expected clamped ratio == original ratio")
	}
}

func TestApplyMaxRethDelta_ClampUpwardIncrease(t *testing.T) {
	b := newNetworkBalances()
	// Ratio jumped from 1.00 to 1.10 — a 0.10 increase
	lastRate := 1.00
	b.OriginalRatioWei = eth.EthToWei(1.10)
	b.OriginalTotalBalanceWei = ethToWei(1100) // arbitrary consistent total

	// Max allowed delta is 0.05
	maxDelta := eth.EthToWei(0.05)

	b.applyMaxRethDelta(maxDelta, lastRate)

	expectedClampedRatio := new(big.Int).Add(eth.EthToWei(lastRate), maxDelta) // 1.05
	if b.ClampedRatioWei.Cmp(expectedClampedRatio) != 0 {
		t.Errorf("clamped ratio: got %s, want %s", b.ClampedRatioWei, expectedClampedRatio)
	}
	// Clamped total must be less than original total (ratio was reduced)
	if b.ClampedTotalBalanceWei.Cmp(b.OriginalTotalBalanceWei) >= 0 {
		t.Errorf("expected clamped total < original total")
	}
}

func TestApplyMaxRethDelta_ClampDownwardDecrease(t *testing.T) {
	b := newNetworkBalances()
	// Ratio dropped from 1.10 to 1.00 — a 0.10 decrease
	lastRate := 1.10
	b.OriginalRatioWei = eth.EthToWei(1.00)
	b.OriginalTotalBalanceWei = ethToWei(1000)

	// Max allowed delta is 0.05
	maxDelta := eth.EthToWei(0.05)

	b.applyMaxRethDelta(maxDelta, lastRate)

	expectedClampedRatio := new(big.Int).Sub(eth.EthToWei(lastRate), maxDelta) // 1.05
	if b.ClampedRatioWei.Cmp(expectedClampedRatio) != 0 {
		t.Errorf("clamped ratio: got %s, want %s", b.ClampedRatioWei, expectedClampedRatio)
	}
	// Clamped total must be greater than original total (ratio was raised)
	if b.ClampedTotalBalanceWei.Cmp(b.OriginalTotalBalanceWei) <= 0 {
		t.Errorf("expected clamped total > original total")
	}
}

func TestApplyMaxRethDelta_ExactlyAtBoundary(t *testing.T) {
	// Delta == maxDelta exactly: should NOT clamp (boundary is exclusive via >)
	b := newNetworkBalances()
	lastRate := 1.00
	b.OriginalRatioWei = eth.EthToWei(1.05)
	b.OriginalTotalBalanceWei = ethToWei(1050)
	maxDelta := eth.EthToWei(0.05) // exactly matches the change

	b.applyMaxRethDelta(maxDelta, lastRate)

	if b.ClampedTotalBalanceWei.Cmp(b.OriginalTotalBalanceWei) != 0 {
		t.Errorf("expected no clamping at exact boundary")
	}
}

func TestApplyMaxRethDelta_ZeroRatioChange(t *testing.T) {
	b := newNetworkBalances()
	lastRate := 1.05
	b.OriginalRatioWei = eth.EthToWei(1.05)
	b.OriginalTotalBalanceWei = ethToWei(1050)
	maxDelta := eth.EthToWei(0.01)

	b.applyMaxRethDelta(maxDelta, lastRate)

	if b.ClampedTotalBalanceWei.Cmp(b.OriginalTotalBalanceWei) != 0 {
		t.Errorf("expected no clamping when ratio unchanged")
	}
}

// ============================================================
// calculateTotalEthAndRethRate
// ============================================================

func TestCalculateTotalEthAndRethRate_BasicSummation(t *testing.T) {
	b := newNetworkBalances()
	b.DepositPool = ethToWei(100)
	b.MinipoolsTotal = ethToWei(200)
	b.MegapoolsUserShareTotal = ethToWei(150)
	b.RETHContract = ethToWei(50)
	b.DistributorShareTotal = ethToWei(30)
	b.SmoothingPoolShare = ethToWei(20)
	b.NodeCreditBalance = ethToWei(10) // should be SUBTRACTED
	b.RETHSupply = ethToWei(500)
	b.MinipoolsStaking = ethToWei(180)
	b.MegapoolStaking = ethToWei(120)

	// Expected total = 100+200+150+50+30+20 - 10 = 540
	expectedTotal := ethToWei(540)

	maxDelta := eth.EthToWei(999) // large enough to never clamp
	b.calculateTotalEthAndRethRate(maxDelta, 0)

	if b.OriginalTotalBalanceWei.Cmp(expectedTotal) != 0 {
		t.Errorf("total ETH: got %s, want %s", b.OriginalTotalBalanceWei, expectedTotal)
	}
}

func TestCalculateTotalEthAndRethRate_NodeCreditIsSubtracted(t *testing.T) {
	b := newNetworkBalances()
	b.DepositPool = ethToWei(500)
	b.RETHSupply = ethToWei(400)
	b.NodeCreditBalance = ethToWei(100)

	maxDelta := eth.EthToWei(999)
	b.calculateTotalEthAndRethRate(maxDelta, 0)

	// 500 - 100 = 400
	expected := ethToWei(400)
	if b.OriginalTotalBalanceWei.Cmp(expected) != 0 {
		t.Errorf("node credit not subtracted correctly: got %s, want %s", b.OriginalTotalBalanceWei, expected)
	}
}

func TestCalculateTotalEthAndRethRate_TotalStakingAggregated(t *testing.T) {
	b := newNetworkBalances()
	b.MinipoolsStaking = ethToWei(300)
	b.MegapoolStaking = ethToWei(200)
	b.RETHSupply = ethToWei(1)

	maxDelta := eth.EthToWei(999)
	b.calculateTotalEthAndRethRate(maxDelta, 0)

	expected := ethToWei(500)
	if b.TotalStaking.Cmp(expected) != 0 {
		t.Errorf("TotalStaking: got %s, want %s", b.TotalStaking, expected)
	}
}

func TestCalculateTotalEthAndRethRate_RatioCalculation(t *testing.T) {
	b := newNetworkBalances()
	// Total ETH = 1100 (after credits), supply = 1000 → ratio = 1.1
	b.DepositPool = ethToWei(1100)
	b.RETHSupply = ethToWei(1000)

	maxDelta := eth.EthToWei(999) // no clamp
	b.calculateTotalEthAndRethRate(maxDelta, 1.0)

	expectedRatio := eth.EthToWei(1.1)
	// Allow 1 wei of rounding tolerance
	diff := new(big.Int).Abs(new(big.Int).Sub(b.OriginalRatioWei, expectedRatio))
	if diff.Cmp(big.NewInt(2)) > 0 {
		t.Errorf("ratio: got %s, want ~%s", b.OriginalRatioWei, expectedRatio)
	}
}

// ============================================================
// getMinipoolBalanceDetails
// ============================================================

func newMinipoolState(blockEpoch uint64) *state.NetworkState {
	return &state.NetworkState{
		BeaconSlotNumber: blockEpoch * 32,
		BeaconConfig: beacon.Eth2Config{
			SlotsPerEpoch: 32,
		},
		MinipoolValidatorDetails: map[rptypes.ValidatorPubkey]beacon.ValidatorStatus{},
	}
}

func newMpd(status rptypes.MinipoolStatus, depositType rptypes.MinipoolDeposit) *rpstate.NativeMinipoolDetails {
	return &rpstate.NativeMinipoolDetails{
		Status:                            status,
		DepositType:                       depositType,
		UserDepositBalance:                ethToWei(16),
		NodeDepositBalance:                ethToWei(16),
		NodeRefundBalance:                 big.NewInt(0),
		Balance:                           big.NewInt(0),
		UserShareOfBalanceIncludingBeacon: ethToWei(16),
		Version:                           3,
	}
}

func TestGetMinipoolBalanceDetails_VacantMinipool(t *testing.T) {
	task := &submitNetworkBalances{}
	mpd := newMpd(rptypes.Staking, rptypes.Full)
	mpd.IsVacant = true
	s := newMinipoolState(100)

	result := task.getMinipoolBalanceDetails(mpd, s, nil)

	if result.UserBalance.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("vacant minipool should return zero balance, got %s", result.UserBalance)
	}
}

func TestGetMinipoolBalanceDetails_DissolvedMinipool(t *testing.T) {
	task := &submitNetworkBalances{}
	mpd := newMpd(rptypes.Dissolved, rptypes.Full)
	s := newMinipoolState(100)

	result := task.getMinipoolBalanceDetails(mpd, s, nil)

	if result.UserBalance.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("dissolved minipool should return zero balance, got %s", result.UserBalance)
	}
}

func TestGetMinipoolBalanceDetails_InitializedMinipool(t *testing.T) {
	task := &submitNetworkBalances{}
	mpd := newMpd(rptypes.Initialized, rptypes.Half)
	mpd.UserDepositBalance = ethToWei(8)
	s := newMinipoolState(100)

	result := task.getMinipoolBalanceDetails(mpd, s, nil)

	if result.UserBalance.Cmp(ethToWei(8)) != 0 {
		t.Errorf("initialized minipool: got %s, want 8 ETH", result.UserBalance)
	}
	if result.IsStaking {
		t.Error("initialized minipool should not be staking")
	}
}

func TestGetMinipoolBalanceDetails_PrelaunchMinipool(t *testing.T) {
	task := &submitNetworkBalances{}
	mpd := newMpd(rptypes.Prelaunch, rptypes.Half)
	mpd.UserDepositBalance = ethToWei(8)
	s := newMinipoolState(100)

	result := task.getMinipoolBalanceDetails(mpd, s, nil)

	if result.UserBalance.Cmp(ethToWei(8)) != 0 {
		t.Errorf("prelaunch minipool: got %s, want 8 ETH", result.UserBalance)
	}
}

func TestGetMinipoolBalanceDetails_ValidatorNotYetActive(t *testing.T) {
	task := &submitNetworkBalances{}
	mpd := newMpd(rptypes.Staking, rptypes.Half)
	mpd.UserDepositBalance = ethToWei(8)

	pubkey := rptypes.ValidatorPubkey{0x02}
	mpd.Pubkey = pubkey

	blockEpoch := uint64(100)
	s := newMinipoolState(blockEpoch)
	s.MinipoolValidatorDetails[pubkey] = beacon.ValidatorStatus{
		Exists:          true,
		ActivationEpoch: 150, // future epoch — not yet active
		ExitEpoch:       FarFutureEpoch,
		Balance:         8000,
	}

	result := task.getMinipoolBalanceDetails(mpd, s, nil)

	if result.UserBalance.Cmp(ethToWei(8)) != 0 {
		t.Errorf("pending validator should use userDepositBalance: got %s", result.UserBalance)
	}
}

func TestGetMinipoolBalanceDetails_ValidatorDoesNotExist(t *testing.T) {
	task := &submitNetworkBalances{}
	mpd := newMpd(rptypes.Staking, rptypes.Half)
	mpd.UserDepositBalance = ethToWei(8)
	mpd.Pubkey = rptypes.ValidatorPubkey{0x03}

	// No entry in MinipoolValidatorDetails → Exists == false
	s := newMinipoolState(100)

	result := task.getMinipoolBalanceDetails(mpd, s, nil)

	if result.UserBalance.Cmp(ethToWei(8)) != 0 {
		t.Errorf("non-existent validator should use userDepositBalance: got %s", result.UserBalance)
	}
}

func TestGetMinipoolBalanceDetails_FullMinipoolInRefundQueue(t *testing.T) {
	// Full minipool with zero userDepositBalance → subtract 16 ETH from user share
	task := &submitNetworkBalances{}
	mpd := newMpd(rptypes.Staking, rptypes.Full)
	mpd.UserDepositBalance = big.NewInt(0)
	mpd.UserShareOfBalanceIncludingBeacon = ethToWei(32)

	pubkey := rptypes.ValidatorPubkey{0x04}
	mpd.Pubkey = pubkey

	blockEpoch := uint64(100)
	s := newMinipoolState(blockEpoch)
	s.MinipoolValidatorDetails[pubkey] = beacon.ValidatorStatus{
		Exists:          true,
		ActivationEpoch: 50,
		ExitEpoch:       FarFutureEpoch,
		Balance:         32000,
	}

	result := task.getMinipoolBalanceDetails(mpd, s, nil)

	// 32 ETH user share - 16 ETH = 16 ETH
	expected := ethToWei(16)
	if result.UserBalance.Cmp(expected) != 0 {
		t.Errorf("full minipool refund queue: got %s, want %s", result.UserBalance, expected)
	}
}

func TestGetMinipoolBalanceDetails_IsStakingFlag_ExitedValidator(t *testing.T) {
	// Validator whose ExitEpoch <= blockEpoch should NOT be staking
	task := &submitNetworkBalances{}
	mpd := newMpd(rptypes.Staking, rptypes.Half)
	mpd.UserDepositBalance = ethToWei(8)
	mpd.UserShareOfBalanceIncludingBeacon = ethToWei(8)

	pubkey := rptypes.ValidatorPubkey{0x05}
	mpd.Pubkey = pubkey

	blockEpoch := uint64(200)
	s := newMinipoolState(blockEpoch)
	s.MinipoolValidatorDetails[pubkey] = beacon.ValidatorStatus{
		Exists:          true,
		ActivationEpoch: 50,
		ExitEpoch:       100, // already exited before blockEpoch
		Balance:         8000,
	}

	result := task.getMinipoolBalanceDetails(mpd, s, nil)

	if result.IsStaking {
		t.Error("exited validator should not be marked as staking")
	}
}

func TestGetMinipoolBalanceDetails_IsStakingFlag_ActiveValidator(t *testing.T) {
	task := &submitNetworkBalances{}
	mpd := newMpd(rptypes.Staking, rptypes.Half)
	mpd.UserDepositBalance = ethToWei(8)
	mpd.UserShareOfBalanceIncludingBeacon = ethToWei(8)

	pubkey := rptypes.ValidatorPubkey{0x06}
	mpd.Pubkey = pubkey

	blockEpoch := uint64(100)
	s := newMinipoolState(blockEpoch)
	s.MinipoolValidatorDetails[pubkey] = beacon.ValidatorStatus{
		Exists:          true,
		ActivationEpoch: 50,
		ExitEpoch:       FarFutureEpoch,
		Balance:         8000,
	}

	result := task.getMinipoolBalanceDetails(mpd, s, nil)

	if !result.IsStaking {
		t.Error("active validator should be marked as staking")
	}
}

// ============================================================
// Stubs
// ============================================================

// stubBeaconClient implements the subset of beacon.Client used by
// FindNextSubmissionTarget and FindLastBlockWithExecutionPayload.
type stubBeaconClient struct {
	beaconHead    beacon.BeaconHead
	beaconHeadErr error
	// blocksByID maps blockId string → BeaconBlock (used by GetBeaconBlock)
	blocksByID        map[string]beacon.BeaconBlock
	getBeaconBlockErr error
	blocksBySlot      map[uint64]beacon.BeaconBlock
}

func (s *stubBeaconClient) GetBeaconHead() (beacon.BeaconHead, error) {
	return s.beaconHead, s.beaconHeadErr
}

func (s *stubBeaconClient) GetBeaconBlock(blockId string) (beacon.BeaconBlock, bool, error) {
	if s.getBeaconBlockErr != nil {
		return beacon.BeaconBlock{}, false, s.getBeaconBlockErr
	}
	// blocksByID is keyed by hex hash (used for ParentBeaconRoot lookups)
	if block, ok := s.blocksByID[blockId]; ok {
		return block, true, nil
	}
	// blocksBySlot is keyed by decimal slot string (used by FindLastBlockWithExecutionPayload)
	if slot, err := strconv.ParseUint(blockId, 10, 64); err == nil {
		if block, ok := s.blocksBySlot[slot]; ok {
			return block, true, nil
		}
	}
	return beacon.BeaconBlock{}, false, nil
}

func (s *stubBeaconClient) GetClientType() (beacon.BeaconClientType, error) {
	return 0, nil
}
func (s *stubBeaconClient) GetSyncStatus() (beacon.SyncStatus, error) {
	return beacon.SyncStatus{}, nil
}
func (s *stubBeaconClient) GetEth2Config() (beacon.Eth2Config, error) {
	return beacon.Eth2Config{}, nil
}
func (s *stubBeaconClient) GetEth2DepositContract() (beacon.Eth2DepositContract, error) {
	return beacon.Eth2DepositContract{}, nil
}
func (s *stubBeaconClient) GetAttestations(blockId string) ([]beacon.AttestationInfo, bool, error) {
	return nil, false, nil
}
func (s *stubBeaconClient) GetBeaconBlockHeader(blockId string) (beacon.BeaconBlockHeader, bool, error) {
	return beacon.BeaconBlockHeader{}, false, nil
}
func (s *stubBeaconClient) GetValidatorStatusByIndex(index string, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	return beacon.ValidatorStatus{}, nil
}
func (s *stubBeaconClient) GetValidatorStatus(pubkey rptypes.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	return beacon.ValidatorStatus{}, nil
}
func (s *stubBeaconClient) GetAllValidators() ([]beacon.ValidatorStatus, error) { return nil, nil }
func (s *stubBeaconClient) GetValidatorStatuses(pubkeys []rptypes.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (map[rptypes.ValidatorPubkey]beacon.ValidatorStatus, error) {
	return nil, nil
}
func (s *stubBeaconClient) GetValidatorIndex(pubkey rptypes.ValidatorPubkey) (string, error) {
	return "", nil
}
func (s *stubBeaconClient) GetValidatorSyncDuties(indices []string, epoch uint64) (map[string]bool, error) {
	return nil, nil
}
func (s *stubBeaconClient) GetValidatorProposerDuties(indices []string, epoch uint64) (map[string]uint64, error) {
	return nil, nil
}
func (s *stubBeaconClient) GetValidatorBalances(indices []string, opts *beacon.ValidatorStatusOptions) (map[string]*big.Int, error) {
	return nil, nil
}
func (s *stubBeaconClient) GetValidatorBalancesSafe(indices []string, opts *beacon.ValidatorStatusOptions) (map[string]*big.Int, error) {
	return nil, nil
}
func (s *stubBeaconClient) GetDomainData(domainType []byte, epoch uint64, useGenesisFork bool) ([]byte, error) {
	return nil, nil
}
func (s *stubBeaconClient) ExitValidator(validatorIndex string, epoch uint64, signature rptypes.ValidatorSignature) error {
	return nil
}
func (s *stubBeaconClient) Close() error { return nil }
func (s *stubBeaconClient) GetEth1DataForEth2Block(blockId string) (beacon.Eth1Data, bool, error) {
	return beacon.Eth1Data{}, false, nil
}
func (s *stubBeaconClient) GetCommitteesForEpoch(epoch *uint64) (beacon.Committees, error) {
	return nil, nil
}
func (s *stubBeaconClient) ChangeWithdrawalCredentials(validatorIndex string, fromBlsPubkey rptypes.ValidatorPubkey, toExecutionAddress common.Address, signature rptypes.ValidatorSignature) error {
	return nil
}
func (s *stubBeaconClient) GetBeaconStateSSZ(slot uint64) (*beacon.BeaconStateSSZ, error) {
	return nil, nil
}
func (s *stubBeaconClient) GetBeaconBlockSSZ(slot uint64) (*beacon.BeaconBlockSSZ, bool, error) {
	return nil, false, nil
}

type stubExecutionClient struct {
	headers   map[uint64]*types.Header
	headerErr error
}

func (s *stubExecutionClient) HeaderByNumber(_ context.Context, number *big.Int) (*types.Header, error) {
	if s.headerErr != nil {
		return nil, s.headerErr
	}
	h, ok := s.headers[number.Uint64()]
	if !ok {
		return nil, nil
	}
	return h, nil
}

func (s *stubExecutionClient) CodeAt(_ context.Context, _ common.Address, _ *big.Int) ([]byte, error) {
	return nil, nil
}
func (s *stubExecutionClient) CallContract(_ context.Context, _ ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	return nil, nil
}
func (s *stubExecutionClient) HeaderByHash(_ context.Context, _ common.Hash) (*types.Header, error) {
	return nil, nil
}
func (s *stubExecutionClient) PendingCodeAt(_ context.Context, _ common.Address) ([]byte, error) {
	return nil, nil
}
func (s *stubExecutionClient) PendingNonceAt(_ context.Context, _ common.Address) (uint64, error) {
	return 0, nil
}
func (s *stubExecutionClient) SuggestGasPrice(_ context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (s *stubExecutionClient) SuggestGasTipCap(_ context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (s *stubExecutionClient) EstimateGas(_ context.Context, _ ethereum.CallMsg) (uint64, error) {
	return 0, nil
}
func (s *stubExecutionClient) SendTransaction(_ context.Context, _ *types.Transaction) error {
	return nil
}
func (s *stubExecutionClient) FilterLogs(_ context.Context, _ ethereum.FilterQuery) ([]types.Log, error) {
	return nil, nil
}
func (s *stubExecutionClient) SubscribeFilterLogs(_ context.Context, _ ethereum.FilterQuery, _ chan<- types.Log) (ethereum.Subscription, error) {
	return nil, nil
}
func (s *stubExecutionClient) TransactionReceipt(_ context.Context, _ common.Hash) (*types.Receipt, error) {
	return nil, nil
}
func (s *stubExecutionClient) BlockNumber(_ context.Context) (uint64, error) { return 0, nil }
func (s *stubExecutionClient) BalanceAt(_ context.Context, _ common.Address, _ *big.Int) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (s *stubExecutionClient) TransactionByHash(_ context.Context, _ common.Hash) (*types.Transaction, bool, error) {
	return nil, false, nil
}
func (s *stubExecutionClient) NonceAt(_ context.Context, _ common.Address, _ *big.Int) (uint64, error) {
	return 0, nil
}
func (s *stubExecutionClient) SyncProgress(_ context.Context) (*ethereum.SyncProgress, error) {
	return nil, nil
}
func (s *stubExecutionClient) LatestBlockTime(_ context.Context) (time.Time, error) {
	return time.Time{}, nil
}
func (s *stubExecutionClient) ChainID(_ context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

// TestFindNextSubmissionTarget_FirstEverSubmission covers the case where
// lastSubmissionBlock == 0 (no prior submission exists).  The function must
// skip the HeaderByNumber / GetBeaconBlock calls for the last submission and
// derive the next target purely from the reference timestamp.
func TestFindNextSubmissionTarget_FirstEverSubmission(t *testing.T) {
	cfg := makeEth2Config(12, 32)

	// Reference timestamp = genesis; interval = 1 day.
	referenceTimestamp := testGenesisTime.Unix()
	intervalSeconds := int64(86400) // 1 day

	// Finalized epoch is 3 days worth of slots into the chain.
	threeDaysOfSlots := uint64(3 * 86400 / 12)
	finalizedEpoch := threeDaysOfSlots / cfg.SlotsPerEpoch

	// The expected submission target is at timestamp = referenceTimestamp +
	// 2*interval (the highest multiple of intervalSeconds that is still <=
	// the start of the finalizedEpoch).
	finalizedEpochStartSlot := finalizedEpoch * cfg.SlotsPerEpoch
	finalizedEpochTimestamp := testGenesisTime.Add(
		time.Duration(finalizedEpochStartSlot*cfg.SecondsPerSlot) * time.Second,
	)
	// Walk forward from referenceTimestamp to find expected maxSubmissionTimestamp.
	expectedMaxTs := referenceTimestamp
	for n := int64(1); ; n++ {
		ts := referenceTimestamp + n*intervalSeconds
		if ts > finalizedEpochTimestamp.Unix() {
			break
		}
		expectedMaxTs = ts
	}
	expectedSlot := uint64(expectedMaxTs-testGenesisTime.Unix()) / cfg.SecondsPerSlot

	// Target EL block number returned by FindLastBlockWithExecutionPayload.
	targetELBlock := uint64(5000)

	bc := &stubBeaconClient{
		beaconHead: beacon.BeaconHead{
			FinalizedEpoch: finalizedEpoch,
		},
		blocksBySlot: map[uint64]beacon.BeaconBlock{
			expectedSlot: {
				Slot:                 expectedSlot,
				HasExecutionPayload:  true,
				ExecutionBlockNumber: targetELBlock,
			},
		},
	}

	ec := &stubExecutionClient{
		headers: map[uint64]*types.Header{
			targetELBlock: {Number: big.NewInt(int64(targetELBlock))},
		},
	}

	// lastSubmissionBlock = 0 → first ever submission, rp.Client is never called.
	slot, slotTime, header, valid, err := utils.FindNextSubmissionTarget(
		nil, // rp — unused when lastSubmissionBlock == 0
		cfg,
		bc,
		ec,
		0, // lastSubmissionBlock
		referenceTimestamp,
		intervalSeconds,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Fatal("expected valid=true for a first submission with a finalized target")
	}
	if slot != expectedSlot {
		t.Errorf("slot: got %d, want %d", slot, expectedSlot)
	}
	if header.Number.Uint64() != targetELBlock {
		t.Errorf("EL block number: got %d, want %d", header.Number.Uint64(), targetELBlock)
	}
	_ = slotTime // slotTime correctness is implicitly covered by slot being right
}

// TestFindNextSubmissionTarget_AlreadySubmittedForBlock verifies that when the
// derived targetBlockNumber <= lastSubmissionBlock the function returns
// valid=false (no new submission needed).
func TestFindNextSubmissionTarget_AlreadySubmittedForBlock(t *testing.T) {
	cfg := makeEth2Config(12, 32)

	referenceTimestamp := testGenesisTime.Unix()
	intervalSeconds := int64(86400)

	oneDaySlots := uint64(86400 / 12)
	finalizedEpoch := oneDaySlots / cfg.SlotsPerEpoch

	// targetSlot is the slot FindLastBlockWithExecutionPayload will land on.
	targetSlot := uint64(oneDaySlots / 2)
	// targetELBlock is what that slot maps to — deliberately lower than
	// lastSubmissionBlock so valid=false is returned.
	targetELBlock := uint64(199)
	lastSubmissionBlock := uint64(200) // already ahead of targetELBlock

	// The EL header for lastSubmissionBlock — ParentBeaconRoot must be non-nil
	// so the function can look up the parent beacon block.
	parentBeaconRoot := makeParentBeaconRoot(targetSlot - 1)
	lastSubmissionHeader := &types.Header{
		Number:           big.NewInt(int64(lastSubmissionBlock)),
		ParentBeaconRoot: &parentBeaconRoot,
	}

	bc := &stubBeaconClient{
		beaconHead: beacon.BeaconHead{FinalizedEpoch: finalizedEpoch},
		// blocksByID: keyed by hex hash for the ParentBeaconRoot lookup
		blocksByID: map[string]beacon.BeaconBlock{
			parentBeaconRoot.Hex(): {Slot: targetSlot - 1},
		},
		// blocksBySlot: keyed by slot for FindLastBlockWithExecutionPayload
		blocksBySlot: map[uint64]beacon.BeaconBlock{
			targetSlot: {
				Slot:                 targetSlot,
				HasExecutionPayload:  true,
				ExecutionBlockNumber: targetELBlock,
			},
		},
	}

	ec := &stubExecutionClient{
		headers: map[uint64]*types.Header{
			// rp.Client uses this to fetch the lastSubmissionBlock header
			lastSubmissionBlock: lastSubmissionHeader,
			// ec uses this to fetch the target block header
			targetELBlock: {Number: big.NewInt(int64(targetELBlock))},
		},
	}

	// rp only needs Client populated — all other fields are unused by this function.
	rp := &rocketpool.RocketPool{Client: ec}

	_, _, _, valid, err := utils.FindNextSubmissionTarget(
		rp,
		cfg,
		bc,
		ec,
		lastSubmissionBlock,
		referenceTimestamp,
		intervalSeconds,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected valid=false when targetBlockNumber <= lastSubmissionBlock")
	}
}

// ============================================================
// hasSubmittedBlockBalances / hasSubmittedSpecificBlockBalances
// ============================================================

// stubStorage is an in-memory storageGetter used by tests.
type stubStorage struct {
	values map[[32]byte]bool
	err    error
}

func (s *stubStorage) GetBool(_ *bind.CallOpts, key [32]byte) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.values[key], nil
}

func TestHasSubmittedBlockBalances_NotSubmitted(t *testing.T) {
	storage := &stubStorage{values: map[[32]byte]bool{}}
	task := &submitNetworkBalances{storage: storage}

	got, err := task.hasSubmittedBlockBalances(common.HexToAddress("0x1234"), 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false when the block has not been submitted")
	}
}

func TestHasSubmittedBlockBalances_Submitted(t *testing.T) {
	nodeAddr := common.HexToAddress("0x1234")
	blockNumber := uint64(12345)
	key := blockBalancesKey(nodeAddr, blockNumber)

	storage := &stubStorage{values: map[[32]byte]bool{key: true}}
	task := &submitNetworkBalances{storage: storage}

	got, err := task.hasSubmittedBlockBalances(nodeAddr, blockNumber)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected true when the block has been submitted")
	}
}

func TestHasSubmittedBlockBalances_DifferentNodeOrBlock(t *testing.T) {
	nodeAddr := common.HexToAddress("0xAAAA")
	blockNumber := uint64(500)
	key := blockBalancesKey(nodeAddr, blockNumber)

	storage := &stubStorage{values: map[[32]byte]bool{key: true}}
	task := &submitNetworkBalances{storage: storage}

	// Same block, different node → should not match
	gotOtherNode, err := task.hasSubmittedBlockBalances(common.HexToAddress("0xBBBB"), blockNumber)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotOtherNode {
		t.Error("expected false for a different node address")
	}

	// Same node, different block → should not match
	gotOtherBlock, err := task.hasSubmittedBlockBalances(nodeAddr, blockNumber+1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotOtherBlock {
		t.Error("expected false for a different block number")
	}
}

func TestHasSubmittedBlockBalances_PropagatesError(t *testing.T) {
	storage := &stubStorage{err: fmt.Errorf("storage unavailable")}
	task := &submitNetworkBalances{storage: storage}

	_, err := task.hasSubmittedBlockBalances(common.HexToAddress("0x1"), 1)
	if err == nil {
		t.Error("expected error to be propagated from storage")
	}
}

func TestHasSubmittedSpecificBlockBalances_NotSubmitted(t *testing.T) {
	storage := &stubStorage{values: map[[32]byte]bool{}}
	task := &submitNetworkBalances{storage: storage}

	b := newNetworkBalances()
	b.ClampedTotalBalanceWei = ethToWei(500)
	b.TotalStaking = ethToWei(300)
	b.SlotTimestamp = 1_234_567_890

	got, err := task.hasSubmittedSpecificBlockBalances(common.HexToAddress("0xabcd"), 99, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false when specific balances have not been submitted")
	}
}

func TestHasSubmittedSpecificBlockBalances_Submitted(t *testing.T) {
	nodeAddr := common.HexToAddress("0xabcd")
	blockNumber := uint64(99)

	b := newNetworkBalances()
	b.ClampedTotalBalanceWei = ethToWei(500)
	b.TotalStaking = ethToWei(300)
	b.SlotTimestamp = 1_234_567_890

	key := specificBlockBalancesKey(nodeAddr, blockNumber, b)
	storage := &stubStorage{values: map[[32]byte]bool{key: true}}
	task := &submitNetworkBalances{storage: storage}

	got, err := task.hasSubmittedSpecificBlockBalances(nodeAddr, blockNumber, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected true when the exact same balances were previously submitted")
	}
}

func TestHasSubmittedSpecificBlockBalances_DifferentValues(t *testing.T) {
	nodeAddr := common.HexToAddress("0xabcd")
	blockNumber := uint64(99)

	submitted := newNetworkBalances()
	submitted.ClampedTotalBalanceWei = ethToWei(500)
	submitted.TotalStaking = ethToWei(300)
	submitted.RETHSupply = ethToWei(400)
	submitted.SlotTimestamp = 1_000

	// Store the specific key for the original submitted values.
	submittedKey := specificBlockBalancesKey(nodeAddr, blockNumber, submitted)
	storage := &stubStorage{values: map[[32]byte]bool{submittedKey: true}}
	task := &submitNetworkBalances{storage: storage}

	// Confirm the original values return true (sanity check).
	gotOriginal, err := task.hasSubmittedSpecificBlockBalances(nodeAddr, blockNumber, submitted)
	if err != nil {
		t.Fatalf("unexpected error on original: %v", err)
	}
	if !gotOriginal {
		t.Fatal("sanity check failed: expected true for the originally submitted values")
	}

	// Change TotalStaking → key no longer matches.
	altered := submitted
	altered.TotalStaking = new(big.Int).Set(ethToWei(301))

	got, err := task.hasSubmittedSpecificBlockBalances(nodeAddr, blockNumber, altered)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false when TotalStaking differs from what was submitted")
	}
}

// TestHasSubmittedSpecificVsBlock verifies that the two functions use distinct
// storage keys: setting the block-level key does not satisfy the specific check.
func TestHasSubmittedSpecificVsBlock(t *testing.T) {
	nodeAddr := common.HexToAddress("0x5678")
	blockNumber := uint64(42)

	b := newNetworkBalances()
	b.ClampedTotalBalanceWei = ethToWei(100)
	b.TotalStaking = ethToWei(50)
	b.SlotTimestamp = 9999

	// Only set the block-level key.
	blockKey := blockBalancesKey(nodeAddr, blockNumber)
	storage := &stubStorage{values: map[[32]byte]bool{blockKey: true}}
	task := &submitNetworkBalances{storage: storage}

	hasBlock, err := task.hasSubmittedBlockBalances(nodeAddr, blockNumber)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hasSpecific, err := task.hasSubmittedSpecificBlockBalances(nodeAddr, blockNumber, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !hasBlock {
		t.Error("expected hasSubmittedBlockBalances=true")
	}
	if hasSpecific {
		t.Error("expected hasSubmittedSpecificBlockBalances=false (key is different from block-level key)")
	}
}

func TestHasSubmittedSpecificBlockBalances_PropagatesError(t *testing.T) {
	storage := &stubStorage{err: fmt.Errorf("storage unavailable")}
	task := &submitNetworkBalances{storage: storage}

	b := newNetworkBalances()
	b.ClampedTotalBalanceWei = big.NewInt(1)
	b.TotalStaking = big.NewInt(0)

	_, err := task.hasSubmittedSpecificBlockBalances(common.HexToAddress("0x1"), 1, b)
	if err == nil {
		t.Error("expected error to be propagated from storage")
	}
}
