package state

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
)

// Compile-time check that StaticBeaconClient satisfies beacon.Client
var _ beacon.Client = (*StaticBeaconClient)(nil)

// StaticBeaconClient serves consensus-layer reads from a pre-loaded
// NetworkState snapshot. Anything not modelled in the snapshot
// (attestations, committees, sync duties, BLS domain data, arbitrary SSZ
// state or block data) returns ErrStaticMode.
type StaticBeaconClient struct {
	state *NetworkState
}

// NewStaticBeaconClient wires the given NetworkState into a static
// beacon.Client implementation.
func NewStaticBeaconClient(ns *NetworkState) *StaticBeaconClient {
	return &StaticBeaconClient{state: ns}
}

func (c *StaticBeaconClient) GetClientType() (beacon.BeaconClientType, error) {
	return beacon.Unknown, nil
}

// GetSyncStatus always reports "fully synced" since the snapshot is a single
// point in time by definition.
func (c *StaticBeaconClient) GetSyncStatus() (beacon.SyncStatus, error) {
	return beacon.SyncStatus{Syncing: false, Progress: 1}, nil
}

func (c *StaticBeaconClient) GetEth2Config() (beacon.Eth2Config, error) {
	return c.state.BeaconConfig, nil
}

func (c *StaticBeaconClient) GetEth2DepositContract() (beacon.Eth2DepositContract, error) {
	return beacon.Eth2DepositContract{}, ErrStaticMode
}

func (c *StaticBeaconClient) GetAttestations(_ string) ([]beacon.AttestationInfo, bool, error) {
	return nil, false, ErrStaticMode
}

// blockHead synthesizes a BeaconBlock for the snapshot's slot. Only the
// fields derivable from the snapshot are populated.
func (c *StaticBeaconClient) blockHead() beacon.BeaconBlock {
	return beacon.BeaconBlock{
		Slot:                 c.state.BeaconSlotNumber,
		HasExecutionPayload:  true,
		ExecutionBlockNumber: c.state.ElBlockNumber,
	}
}

// GetBeaconBlock honours the symbolic "head" / "finalized" / "genesis"
// identifiers and any numeric slot equal to the snapshot's slot. Anything
// else returns ErrStaticMode.
func (c *StaticBeaconClient) GetBeaconBlock(blockId string) (beacon.BeaconBlock, bool, error) {
	switch blockId {
	case "head", "finalized":
		return c.blockHead(), true, nil
	case "genesis":
		return beacon.BeaconBlock{}, false, ErrStaticMode
	}
	if slot, err := strconv.ParseUint(blockId, 10, 64); err == nil && slot == c.state.BeaconSlotNumber {
		return c.blockHead(), true, nil
	}
	return beacon.BeaconBlock{}, false, ErrStaticMode
}

func (c *StaticBeaconClient) GetBeaconBlockHeader(blockId string) (beacon.BeaconBlockHeader, bool, error) {
	blk, exists, err := c.GetBeaconBlock(blockId)
	if err != nil || !exists {
		return beacon.BeaconBlockHeader{}, exists, err
	}
	return beacon.BeaconBlockHeader{Slot: blk.Slot, ProposerIndex: blk.ProposerIndex}, true, nil
}

// GetBeaconHead is synthesised from the snapshot's slot. The finalized and
// justified epochs are taken to be the epoch of the snapshot slot itself
// since a snapshot, by definition, represents a settled view of the chain.
func (c *StaticBeaconClient) GetBeaconHead() (beacon.BeaconHead, error) {
	cfg := c.state.BeaconConfig
	if cfg.SlotsPerEpoch == 0 {
		return beacon.BeaconHead{}, fmt.Errorf("beacon config has SlotsPerEpoch=0")
	}
	epoch := c.state.BeaconSlotNumber / cfg.SlotsPerEpoch
	return beacon.BeaconHead{
		Epoch:                  epoch,
		FinalizedEpoch:         epoch,
		JustifiedEpoch:         epoch,
		PreviousJustifiedEpoch: epoch,
	}, nil
}

// lookupValidator returns the snapshot's ValidatorStatus for the given
// pubkey, checking both the minipool and megapool validator maps.
func (c *StaticBeaconClient) lookupValidator(pubkey types.ValidatorPubkey) (beacon.ValidatorStatus, bool) {
	if v, ok := c.state.MinipoolValidatorDetails[pubkey]; ok {
		return v, true
	}
	if v, ok := c.state.MegapoolValidatorDetails[pubkey]; ok {
		return v, true
	}
	return beacon.ValidatorStatus{}, false
}

func (c *StaticBeaconClient) GetValidatorStatusByIndex(index string, _ *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	for _, v := range c.state.MinipoolValidatorDetails {
		if v.Index == index {
			return v, nil
		}
	}
	for _, v := range c.state.MegapoolValidatorDetails {
		if v.Index == index {
			return v, nil
		}
	}
	return beacon.ValidatorStatus{Exists: false}, nil
}

func (c *StaticBeaconClient) GetValidatorStatus(pubkey types.ValidatorPubkey, _ *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	if v, ok := c.lookupValidator(pubkey); ok {
		return v, nil
	}
	return beacon.ValidatorStatus{Exists: false}, nil
}

func (c *StaticBeaconClient) GetAllValidators() ([]beacon.ValidatorStatus, error) {
	out := make([]beacon.ValidatorStatus, 0, len(c.state.MinipoolValidatorDetails)+len(c.state.MegapoolValidatorDetails))
	for _, v := range c.state.MinipoolValidatorDetails {
		out = append(out, v)
	}
	for _, v := range c.state.MegapoolValidatorDetails {
		out = append(out, v)
	}
	return out, nil
}

func (c *StaticBeaconClient) GetValidatorStatuses(pubkeys []types.ValidatorPubkey, _ *beacon.ValidatorStatusOptions) (map[types.ValidatorPubkey]beacon.ValidatorStatus, error) {
	out := make(map[types.ValidatorPubkey]beacon.ValidatorStatus, len(pubkeys))
	for _, pk := range pubkeys {
		if v, ok := c.lookupValidator(pk); ok {
			out[pk] = v
		}
	}
	return out, nil
}

func (c *StaticBeaconClient) GetValidatorIndex(pubkey types.ValidatorPubkey) (string, error) {
	if v, ok := c.lookupValidator(pubkey); ok {
		return v.Index, nil
	}
	return "", fmt.Errorf("validator %s not found in snapshot", pubkey.Hex())
}

func (c *StaticBeaconClient) GetValidatorSyncDuties(_ []string, _ uint64) (map[string]bool, error) {
	return nil, ErrStaticMode
}

func (c *StaticBeaconClient) GetValidatorProposerDuties(_ []string, _ uint64) (map[string]uint64, error) {
	return nil, ErrStaticMode
}

func (c *StaticBeaconClient) GetValidatorBalances(indices []string, _ *beacon.ValidatorStatusOptions) (map[string]*big.Int, error) {
	out := make(map[string]*big.Int, len(indices))
	lookup := make(map[string]uint64)
	for _, v := range c.state.MinipoolValidatorDetails {
		lookup[v.Index] = v.Balance
	}
	for _, v := range c.state.MegapoolValidatorDetails {
		lookup[v.Index] = v.Balance
	}
	for _, idx := range indices {
		if bal, ok := lookup[idx]; ok {
			out[idx] = new(big.Int).SetUint64(bal)
		}
	}
	return out, nil
}

func (c *StaticBeaconClient) GetValidatorBalancesSafe(indices []string, opts *beacon.ValidatorStatusOptions) (map[string]*big.Int, error) {
	return c.GetValidatorBalances(indices, opts)
}

func (c *StaticBeaconClient) GetDomainData(_ []byte, _ uint64, _ bool) ([]byte, error) {
	return nil, ErrStaticMode
}

func (c *StaticBeaconClient) ExitValidator(_ string, _ uint64, _ types.ValidatorSignature) error {
	return ErrStaticMode
}

func (c *StaticBeaconClient) Close() error {
	return nil
}

func (c *StaticBeaconClient) GetEth1DataForEth2Block(_ string) (beacon.Eth1Data, bool, error) {
	return beacon.Eth1Data{}, false, ErrStaticMode
}

func (c *StaticBeaconClient) GetCommitteesForEpoch(_ *uint64) (beacon.Committees, error) {
	return nil, ErrStaticMode
}

func (c *StaticBeaconClient) ChangeWithdrawalCredentials(_ string, _ types.ValidatorPubkey, _ common.Address, _ types.ValidatorSignature) error {
	return ErrStaticMode
}

func (c *StaticBeaconClient) GetBeaconStateSSZ(_ uint64) (*beacon.BeaconStateSSZ, error) {
	return nil, ErrStaticMode
}

func (c *StaticBeaconClient) GetBeaconBlockSSZ(_ uint64) (*beacon.BeaconBlockSSZ, bool, error) {
	return nil, false, ErrStaticMode
}
