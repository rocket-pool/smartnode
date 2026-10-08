package performancewatcher

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/go-bitfield"
)

type Requester interface {
	WatcherRequest(context.Context, string) ([]byte, error)
}
type uintString uint64

func (u *uintString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	n, err := strconv.ParseUint(s, 10, 64)
	*u = uintString(n)
	return err
}

type metadata struct {
	Finalized  *bool `json:"finalized"`
	Optimistic *bool `json:"execution_optimistic"`
}

func (m metadata) valid() error {
	if m.Finalized == nil || !*m.Finalized || m.Optimistic == nil || *m.Optimistic {
		return fmt.Errorf("response is not explicitly finalized and execution-verified")
	}
	return nil
}

type header struct {
	Slot                uint64
	Root, Parent, State string
}
type headerResponse struct {
	metadata
	Data struct {
		Root      common.Hash `json:"root"`
		Canonical bool        `json:"canonical"`
		Header    struct {
			Message struct {
				Slot   *uintString `json:"slot"`
				Parent common.Hash `json:"parent_root"`
				State  common.Hash `json:"state_root"`
			} `json:"message"`
		} `json:"header"`
	} `json:"data"`
}
type attestation struct {
	Bits       string `json:"aggregation_bits"`
	Committees string `json:"committee_bits"`
	Data       struct {
		Slot   uintString `json:"slot"`
		Index  uintString `json:"index"`
		Target struct {
			Epoch uintString  `json:"epoch"`
			Root  common.Hash `json:"root"`
		} `json:"target"`
	} `json:"data"`
}
type attestationsResponse struct {
	metadata
	Version string        `json:"version"`
	Data    []attestation `json:"data"`
}

func (a *attestation) UnmarshalJSON(raw []byte) error {
	// Missing numeric fields must not silently become epoch/slot/index zero.
	var fields struct {
		Bits *string `json:"aggregation_bits"`
		Data *struct {
			Slot   *uintString `json:"slot"`
			Index  *uintString `json:"index"`
			Target *struct {
				Epoch *uintString  `json:"epoch"`
				Root  *common.Hash `json:"root"`
			} `json:"target"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if fields.Bits == nil || *fields.Bits == "" || fields.Data == nil || fields.Data.Slot == nil || fields.Data.Index == nil || fields.Data.Target == nil || fields.Data.Target.Epoch == nil || fields.Data.Target.Root == nil {
		return fmt.Errorf("incomplete attestation response")
	}
	type plain attestation
	return json.Unmarshal(raw, (*plain)(a))
}

// Beacon is owned by one worker. Raw caches contain at most a few epochs.
type Beacon struct {
	clients                                                []Requester
	verified                                               []bool
	GenesisRoot                                            string
	ExpectedChainID                                        uint64
	GenesisTime, SlotsPerEpoch, SecondsPerSlot, DenebEpoch uint64
	headers                                                map[string]header
	attestations                                           map[string]attestationsResponse
	Failures                                               atomic.Uint64
	Requests                                               atomic.Uint64
}

func NewBeacon(clients []Requester) *Beacon {
	return &Beacon{clients: clients, verified: make([]bool, len(clients)), headers: map[string]header{}, attestations: map[string]attestationsResponse{}}
}

func (b *Beacon) read(ctx context.Context, path string, parse func([]byte) error) error {
	var errs []error
	for i, c := range b.clients {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !b.verified[i] && b.GenesisRoot != "" {
			b.Requests.Add(1)
			raw, err := c.WatcherRequest(ctx, "/eth/v1/beacon/genesis")
			var g struct {
				Data struct {
					Root common.Hash `json:"genesis_validators_root"`
				} `json:"data"`
			}
			if err == nil {
				err = json.Unmarshal(raw, &g)
			}
			if err == nil && g.Data.Root.Hex() != b.GenesisRoot {
				return fmt.Errorf("configured beacon clients have different genesis roots")
			}
			if err != nil {
				b.Failures.Add(1)
				errs = append(errs, err)
				continue
			}
			b.verified[i] = true
		}
		b.Requests.Add(1)
		raw, err := c.WatcherRequest(ctx, path)
		if err == nil {
			err = parse(raw)
		}
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrFinalityConflict) {
			return err
		}
		b.Failures.Add(1)
		errs = append(errs, err)
	}
	if len(errs) == 0 {
		return fmt.Errorf("no beacon clients configured")
	}
	return errors.Join(errs...)
}

func (b *Beacon) Init(ctx context.Context) error {
	err := b.read(ctx, "/eth/v1/beacon/genesis", func(raw []byte) error {
		var g struct {
			Data struct {
				Root common.Hash `json:"genesis_validators_root"`
				Time uintString  `json:"genesis_time"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &g); err != nil {
			return err
		}
		if g.Data.Root == (common.Hash{}) {
			return fmt.Errorf("missing genesis root")
		}
		b.GenesisRoot = g.Data.Root.Hex()
		b.GenesisTime = uint64(g.Data.Time)
		return nil
	})
	if err != nil {
		return err
	}
	if b.ExpectedChainID != 0 {
		err = b.read(ctx, "/eth/v1/config/deposit_contract", func(raw []byte) error {
			var r struct {
				Data struct {
					ChainID uintString `json:"chain_id"`
				} `json:"data"`
			}
			if err := json.Unmarshal(raw, &r); err != nil {
				return err
			}
			if uint64(r.Data.ChainID) != b.ExpectedChainID {
				return fmt.Errorf("beacon execution chain ID does not match configured network")
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return b.read(ctx, "/eth/v1/config/spec", func(raw []byte) error {
		var s struct {
			Data struct {
				SlotsPerEpoch  *uintString `json:"SLOTS_PER_EPOCH"`
				SecondsPerSlot *uintString `json:"SECONDS_PER_SLOT"`
				SlotDurationMS *uintString `json:"SLOT_DURATION_MS"`
				DenebEpoch     *uintString `json:"DENEB_FORK_EPOCH"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &s); err != nil {
			return err
		}
		// The spec also contains structured values, such as BLOB_SCHEDULE.
		// Decode only the timing fields required by this collector.
		if s.Data.SlotsPerEpoch == nil || s.Data.DenebEpoch == nil {
			return fmt.Errorf("missing beacon timing fields")
		}
		b.SlotsPerEpoch = uint64(*s.Data.SlotsPerEpoch)
		b.DenebEpoch = uint64(*s.Data.DenebEpoch)
		if s.Data.SecondsPerSlot != nil {
			b.SecondsPerSlot = uint64(*s.Data.SecondsPerSlot)
		} else {
			if s.Data.SlotDurationMS == nil || *s.Data.SlotDurationMS%1000 != 0 {
				return fmt.Errorf("unsupported slot duration")
			}
			b.SecondsPerSlot = uint64(*s.Data.SlotDurationMS) / 1000
		}
		if b.SlotsPerEpoch == 0 || b.SecondsPerSlot == 0 || b.SlotsPerEpoch > math.MaxUint64/b.SecondsPerSlot {
			return fmt.Errorf("invalid beacon timing")
		}
		return nil
	})
}

func (b *Beacon) Checkpoint(ctx context.Context) (Checkpoint, error) {
	err := b.read(ctx, "/eth/v1/node/syncing", func(raw []byte) error {
		var r struct {
			Data struct {
				HeadSlot   *uintString `json:"head_slot"`
				Syncing    bool        `json:"is_syncing"`
				Optimistic bool        `json:"is_optimistic"`
				Offline    bool        `json:"el_offline"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &r); err != nil {
			return err
		}
		if r.Data.Syncing || r.Data.Optimistic || r.Data.Offline {
			return ErrSyncing
		}
		if r.Data.HeadSlot == nil {
			return fmt.Errorf("missing beacon head slot")
		}
		// Some clients briefly report is_syncing=false even after their head
		// has stopped advancing. Avoid adding historical replay load then.
		now := uint64(time.Now().Unix())
		if now >= b.GenesisTime {
			current := (now - b.GenesisTime) / b.SecondsPerSlot
			head := uint64(*r.Data.HeadSlot)
			if current > head && current-head > 2 {
				return fmt.Errorf("beacon head is %d slots behind: %w", current-head, ErrSyncing)
			}
		}
		return nil
	})
	if err != nil {
		return Checkpoint{}, err
	}
	var cp Checkpoint
	err = b.read(ctx, "/eth/v1/beacon/states/head/finality_checkpoints", func(raw []byte) error {
		var r struct {
			Optimistic *bool `json:"execution_optimistic"`
			Data       struct {
				Finalized struct {
					Epoch uintString  `json:"epoch"`
					Root  common.Hash `json:"root"`
				} `json:"finalized"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &r); err != nil {
			return err
		}
		if r.Optimistic == nil || *r.Optimistic || r.Data.Finalized.Root == (common.Hash{}) {
			return fmt.Errorf("unavailable finalized checkpoint")
		}
		cp.Epoch = uint64(r.Data.Finalized.Epoch)
		cp.Root = r.Data.Finalized.Root.Hex()
		return nil
	})
	if err != nil {
		return cp, err
	}
	h, err := b.header(ctx, cp.Root)
	if err != nil {
		return cp, err
	}
	cp.Slot = h.Slot
	cp.StateRoot = h.State
	if cp.Epoch > math.MaxUint64/b.SlotsPerEpoch || cp.Slot > cp.Epoch*b.SlotsPerEpoch {
		return cp, fmt.Errorf("invalid finalized checkpoint slot")
	}
	return cp, nil
}

func (b *Beacon) header(ctx context.Context, id string) (header, error) {
	if h, ok := b.headers[id]; ok {
		return h, nil
	}
	var h header
	err := b.read(ctx, "/eth/v1/beacon/headers/"+id, func(raw []byte) error {
		var r headerResponse
		if err := json.Unmarshal(raw, &r); err != nil {
			return err
		}
		if err := r.valid(); err != nil {
			return err
		}
		if !r.Data.Canonical || r.Data.Root == (common.Hash{}) || r.Data.Header.Message.State == (common.Hash{}) {
			return fmt.Errorf("invalid canonical header")
		}
		if r.Data.Header.Message.Slot == nil {
			return fmt.Errorf("missing header slot")
		}
		if strings.HasPrefix(id, "0x") && r.Data.Root.Hex() != id {
			return ErrFinalityConflict
		}
		if slot, err := strconv.ParseUint(id, 10, 64); err == nil && uint64(*r.Data.Header.Message.Slot) != slot {
			return fmt.Errorf("header slot does not match requested slot")
		}
		h = header{uint64(*r.Data.Header.Message.Slot), r.Data.Root.Hex(), r.Data.Header.Message.Parent.Hex(), r.Data.Header.Message.State.Hex()}
		return nil
	})
	if err == nil {
		b.headers[h.Root] = h
	}
	return h, err
}

// The first header at/after end proves gaps via its parent, including fully
// skipped epochs. Failed slot lookups are only search hints, never evidence.
func (b *Beacon) anchor(ctx context.Context, end uint64, cp Checkpoint) (header, error) {
	best := header{Slot: cp.Slot, Root: cp.Root, State: cp.StateRoot}
	if cp.Slot < end {
		return b.header(ctx, cp.Root)
	}
	for _, h := range b.headers {
		if h.Slot >= end && h.Slot <= best.Slot {
			best = h
		}
	}
	if best.Slot-end <= b.SlotsPerEpoch*2 {
		return b.header(ctx, best.Root)
	}
	for slot := end; slot <= cp.Slot && slot-end < 4096; slot++ {
		h, err := b.header(ctx, strconv.FormatUint(slot, 10))
		if err == nil {
			return h, nil
		}
		if ctx.Err() != nil {
			return header{}, ctx.Err()
		}
	}
	return header{}, fmt.Errorf("no available finalized anchor near slot %d", end)
}

func (b *Beacon) collectAttestations(ctx context.Context, h header) (attestationsResponse, error) {
	if r, ok := b.attestations[h.Root]; ok {
		return r, nil
	}
	var r attestationsResponse
	err := b.read(ctx, "/eth/v2/beacon/blocks/"+h.Root+"/attestations", func(raw []byte) error {
		r = attestationsResponse{}
		if err := json.Unmarshal(raw, &r); err != nil {
			return err
		}
		if err := r.valid(); err != nil {
			return err
		}
		if r.Data == nil {
			return fmt.Errorf("missing attestation array")
		}
		switch r.Version {
		case "altair", "bellatrix", "capella", "deneb", "electra", "fulu", "gloas":
			return nil
		default:
			return fmt.Errorf("unsupported attestation version %q", r.Version)
		}
	})
	if err == nil {
		b.attestations[h.Root] = r
	}
	return r, err
}

func (b *Beacon) Collect(ctx context.Context, epoch uint64, cp Checkpoint, validators []Validator) (EpochRecord, error) {
	r := NewRecord(epoch, len(validators))
	r.FinalizedRoot = cp.Root
	if cp.Epoch < 2 || epoch > cp.Epoch-2 || epoch > math.MaxUint64/b.SlotsPerEpoch-2 {
		return r, fmt.Errorf("epoch inclusion window is not finalized")
	}
	start, end := epoch*b.SlotsPerEpoch, (epoch+2)*b.SlotsPerEpoch
	// Evict before fetching to keep raw data bounded even during backfill jumps.
	for root, h := range b.headers {
		if h.Slot+b.SlotsPerEpoch < start || h.Slot > end+b.SlotsPerEpoch {
			delete(b.headers, root)
			delete(b.attestations, root)
		}
	}
	h, err := b.anchor(ctx, end, cp)
	if err != nil {
		return r, err
	}
	var blocks []header
	for {
		if h.Slot <= start {
			r.TargetRoot = h.Root
			if h.Slot == start {
				blocks = append(blocks, h)
			}
			break
		}
		if h.Slot < end {
			blocks = append(blocks, h)
		}
		parent, err := b.header(ctx, h.Parent)
		if err != nil {
			return r, err
		}
		if parent.Slot >= h.Slot {
			return r, fmt.Errorf("non-descending finalized ancestry")
		}
		h = parent
	}
	// Pin committees to a state in the duty epoch, or its canonical boundary
	// ancestor when all early slots were skipped. No head fallback is allowed.
	state := h.State
	for _, block := range blocks {
		if block.Slot/b.SlotsPerEpoch == epoch {
			state = block.State
			break
		}
	}
	type committee struct {
		Index      *uintString  `json:"index"`
		Slot       *uintString  `json:"slot"`
		Validators []uintString `json:"validators"`
	}
	var committees struct {
		metadata
		Data []committee `json:"data"`
	}
	parseCommittees := func(raw []byte) error {
		committees.Data = nil
		committees.metadata = metadata{}
		if err := json.Unmarshal(raw, &committees); err != nil {
			return err
		}
		if err := committees.valid(); err != nil {
			return err
		}
		if committees.Data == nil {
			return fmt.Errorf("missing committees")
		}
		for _, c := range committees.Data {
			if c.Index == nil || c.Slot == nil {
				return fmt.Errorf("incomplete committee response")
			}
		}
		return nil
	}
	err = b.read(ctx, fmt.Sprintf("/eth/v1/beacon/states/%s/committees?epoch=%d", state, epoch), parseCommittees)
	if err != nil && !errors.Is(err, ErrFinalityConflict) && ctx.Err() == nil {
		// A fully skipped epoch may require the processed empty-slot state,
		// rather than its older boundary ancestor. The slot is finalized and
		// response metadata must confirm that; this never falls back to head.
		err = b.read(ctx, fmt.Sprintf("/eth/v1/beacon/states/%d/committees?epoch=%d", start, epoch), parseCommittees)
	}
	if err != nil {
		return r, err
	}
	positions := make(map[uint64]int, len(validators))
	for i, v := range validators {
		if v.Eligible && epoch >= v.Activation && epoch < v.Exit {
			positions[v.Index] = i
		} else {
			r.Set(i, Inactive)
		}
	}
	type key struct{ slot, index uint64 }
	duties := map[key][]int{}
	assigned := make([]bool, len(validators))
	for _, c := range committees.Data {
		k := key{uint64(*c.Slot), uint64(*c.Index)}
		if k.slot/b.SlotsPerEpoch != epoch || k.index >= 64 || len(c.Validators) == 0 {
			return r, fmt.Errorf("invalid committee")
		}
		if _, ok := duties[k]; ok {
			return r, fmt.Errorf("duplicate committee")
		}
		d := make([]int, len(c.Validators))
		for j, index := range c.Validators {
			d[j] = -1
			if i, ok := positions[uint64(index)]; ok {
				if assigned[i] {
					return r, fmt.Errorf("duplicate validator duty")
				}
				assigned[i] = true
				d[j] = i
				r.Set(i, Missed)
			}
		}
		duties[k] = d
	}
	for _, i := range positions {
		if !assigned[i] {
			return r, fmt.Errorf("missing duty for validator %d", validators[i].Index)
		}
	}
	for _, block := range blocks {
		attestations, err := b.collectAttestations(ctx, block)
		if err != nil {
			return r, err
		}
		for _, a := range attestations.Data {
			slot := uint64(a.Data.Slot)
			if uint64(a.Data.Target.Epoch) != epoch || a.Data.Target.Root.Hex() != r.TargetRoot {
				continue
			}
			if slot/b.SlotsPerEpoch != epoch || slot >= block.Slot {
				return r, fmt.Errorf("invalid included attestation slot")
			}
			if block.Slot/b.SlotsPerEpoch < b.DenebEpoch && block.Slot-slot > b.SlotsPerEpoch {
				continue
			}
			bits, err := hex.DecodeString(strings.TrimPrefix(a.Bits, "0x"))
			if err != nil || len(bits) == 0 || bits[len(bits)-1] == 0 {
				return r, fmt.Errorf("invalid aggregation bitlist")
			}
			var indices []uint64
			electra := attestations.Version == "electra" || attestations.Version == "fulu" || attestations.Version == "gloas"
			if electra {
				cb, e := hex.DecodeString(strings.TrimPrefix(a.Committees, "0x"))
				if e != nil || len(cb) != 8 {
					return r, fmt.Errorf("invalid committee bits")
				}
				// Gloas uses index for payload availability (0 or 1), which
				// affects the head flag, not timely-target participation.
				if (attestations.Version == "gloas" && a.Data.Index > 1) || (attestations.Version != "gloas" && a.Data.Index != 0) {
					return r, fmt.Errorf("invalid attestation index")
				}
				for j := uint64(0); j < 64; j++ {
					if cb[j/8]&(1<<(j%8)) != 0 {
						indices = append(indices, j)
					}
				}
			} else {
				indices = []uint64{uint64(a.Data.Index)}
			}
			if len(indices) == 0 {
				return r, fmt.Errorf("empty attestation committees")
			}
			total := 0
			for _, ci := range indices {
				d, ok := duties[key{slot, ci}]
				if !ok {
					return r, fmt.Errorf("attestation references unavailable committee")
				}
				total += len(d)
			}
			bl := bitfield.Bitlist(bits)
			if bl.Len() != uint64(total) {
				return r, fmt.Errorf("aggregation bitlist length does not match committees")
			}
			offset := 0
			for _, ci := range indices {
				for j, i := range duties[key{slot, ci}] {
					if i >= 0 && bl.BitAt(uint64(offset+j)) {
						r.Set(i, Timely)
					}
				}
				offset += len(duties[key{slot, ci}])
			}
		}
	}
	return r, nil
}

func (b *Beacon) ResolveValidators(ctx context.Context, cp Checkpoint, in []Validator) ([]Validator, error) {
	var out []Validator
	// Keep GET URLs below common reverse-proxy limits.
	for start := 0; start < len(in); start += 50 {
		end := min(start+50, len(in))
		ids := make([]string, end-start)
		byPubkey := map[string]Validator{}
		for i, v := range in[start:end] {
			ids[i] = v.Pubkey
			byPubkey[strings.ToLower(v.Pubkey)] = v
		}
		err := b.read(ctx, "/eth/v1/beacon/states/"+cp.StateRoot+"/validators?id="+strings.Join(ids, ","), func(raw []byte) error {
			var r struct {
				metadata
				Data []struct {
					Index     *uintString `json:"index"`
					Validator struct {
						Pubkey     string      `json:"pubkey"`
						Activation *uintString `json:"activation_epoch"`
						Exit       *uintString `json:"exit_epoch"`
					} `json:"validator"`
				} `json:"data"`
			}
			if err := json.Unmarshal(raw, &r); err != nil {
				return err
			}
			if err := r.valid(); err != nil {
				return err
			}
			if r.Data == nil {
				return fmt.Errorf("missing validators")
			}
			var resolved []Validator
			seen := map[uint64]bool{}
			for _, bv := range r.Data {
				if bv.Index == nil || bv.Validator.Activation == nil || bv.Validator.Exit == nil {
					return fmt.Errorf("incomplete validator response")
				}
				v, ok := byPubkey[strings.ToLower(bv.Validator.Pubkey)]
				if !ok {
					return fmt.Errorf("unexpected validator identity")
				}
				v.Index = uint64(*bv.Index)
				v.Activation = uint64(*bv.Validator.Activation)
				v.Exit = uint64(*bv.Validator.Exit)
				if seen[v.Index] {
					return fmt.Errorf("duplicate validator index")
				}
				seen[v.Index] = true
				resolved = append(resolved, v)
			}
			out = append(out, resolved...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
