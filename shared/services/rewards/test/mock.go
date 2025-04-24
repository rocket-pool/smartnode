package test

import (
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	rprewards "github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/rewards/fees"
	"github.com/rocket-pool/smartnode/shared/services/state"
)

const FarFutureEpoch uint64 = 0xffffffffffffffff

// This file contains structs useful for quickly creating mock histories for testing.

func (h *MockHistory) GetValidatorIndex() string {
	u, err := strconv.ParseUint(h.lastValidatorIndex, 10, 64)
	if err != nil {
		panic(err)
	}
	h.lastValidatorIndex = strconv.FormatUint(u+1, 10)
	return h.lastValidatorIndex
}

func (h *MockHistory) GetValidatorPubkey() types.ValidatorPubkey {
	next := big.NewInt(0).Add(big.NewInt(0).SetBytes(h.lastValidatorPubkey.Bytes()), big.NewInt(1))
	h.lastValidatorPubkey = types.BytesToValidatorPubkey(next.Bytes())
	return h.lastValidatorPubkey
}

func (h *MockHistory) GetMinipoolAddress() common.Address {
	next := big.NewInt(0).Add(big.NewInt(0).SetBytes(h.lastMinipoolAddress.Bytes()), big.NewInt(1))
	h.lastMinipoolAddress = common.BigToAddress(next)
	return h.lastMinipoolAddress
}

func (h *MockHistory) GetNodeAddress() common.Address {
	next := big.NewInt(0).Add(big.NewInt(0).SetBytes(h.lastNodeAddress.Bytes()), big.NewInt(1))
	h.lastNodeAddress = common.BigToAddress(next)
	return h.lastNodeAddress
}

var oneEth = big.NewInt(1000000000000000000)
var thirtyTwoEth = big.NewInt(0).Mul(oneEth, big.NewInt(32))

func (h *MockHistory) GetMinipoolAttestationScoreAndCount(address common.Address, state *state.NetworkState) (*big.Int, uint64) {
	out := big.NewInt(0)
	mpi := state.MinipoolDetailsByAddress[address]
	nodeDetails := state.NodeDetailsByAddress[mpi.NodeAddress]

	// Check every slot in the history
	count := uint64(0)
	for slot := h.GetConsensusStartBlock(); slot <= h.GetConsensusEndBlock(); slot++ {
		// Get the time at the slot
		blockTime := h.BeaconConfig.GetSlotTime(slot)
		// Check the status of the minipool at this time
		if mpi.Status != types.Staking {
			continue
		}
		if mpi.Finalised {
			continue
		}
		// Check if the minipool was opted in at this time
		if !nodeDetails.WasOptedInAt(blockTime) {
			continue
		}
		pubkey := mpi.Pubkey
		validator := state.MinipoolValidatorDetails[pubkey]
		// Check if the validator was exited before this slot
		if validator.ExitEpoch <= h.BeaconConfig.SlotToEpoch(slot) {
			continue
		}
		index := validator.Index
		indexInt, _ := strconv.ParseUint(index, 10, 64)
		// Count the attestation if index%32 == slot%32
		if indexInt%32 == uint64(slot%32) {
			count++

			bond, fee := mpi.GetMinipoolBondAndNodeFee(blockTime)
			// Give the minipool a score according to its fee
			eligibleBorrowedEth := state.GetEligibleBorrowedEth(nodeDetails)
			_, percentOfBorrowedEth := state.GetStakedRplValueInEthAndPercentOfBorrowedEth(eligibleBorrowedEth, nodeDetails.RplStake)
			fee = fees.GetMinipoolFeeWithBonus(bond, fee, percentOfBorrowedEth)
			minipoolScore := big.NewInt(0).Sub(oneEth, fee) // 1 - fee
			minipoolScore.Mul(minipoolScore, bond)          // Multiply by bond
			minipoolScore.Div(minipoolScore, thirtyTwoEth)  // Divide by 32 to get the bond as a fraction of a total validator
			minipoolScore.Add(minipoolScore, fee)           // Total = fee + (bond/32)(1 - fee)
			out.Add(out, minipoolScore)
		}
	}
	return out, count
}

type MockMinipool struct {
	Address            common.Address
	Pubkey             types.ValidatorPubkey
	Status             types.MinipoolStatus
	StatusBlock        *big.Int
	StatusTime         time.Time
	Finalised          bool
	NodeFee            *big.Int
	NodeDepositBalance *big.Int
	NodeAddress        common.Address

	LastBondReductionTime        time.Time
	LastBondReductionPrevValue   *big.Int
	LastBondReductionPrevNodeFee *big.Int

	ValidatorIndex string

	// Withdrawal amount to add to the minipool during its SP period
	SPWithdrawals *big.Int

	// Withdrawal amount to add to the minipool during its regular period
	OptedOutWithdrawals *big.Int

	// The epoch after which the minipool is withdrawable.
	// Defaults to FAR_FUTURE_EPOCH.
	WithdrawableEpoch uint64

	Notes []string
}

type BondSize *big.Int

var (
	BondSizeEightEth      = BondSize(eth.EthToWei(8))
	BondSizeSixteenEth    = BondSize(eth.EthToWei(16))
	_bondSizeThirtyTwoEth = BondSize(eth.EthToWei(32))
)

func (h *MockHistory) GetNewDefaultMockMinipool(bondSize BondSize) *MockMinipool {
	if (*big.Int)(_bondSizeThirtyTwoEth).Cmp(bondSize) <= 0 {
		panic("Bond size must be less than 32 ether")
	}

	out := &MockMinipool{
		Address: h.GetMinipoolAddress(),
		Pubkey:  h.GetValidatorPubkey(),
		// By default, staked since always
		Status:      types.Staking,
		StatusBlock: big.NewInt(0),
		StatusTime:  time.Unix(DefaultMockHistoryGenesis, 0),
		// Default to 10% to make math simpler. Aka 0.1 ether
		NodeFee:            big.NewInt(100000000000000000),
		NodeDepositBalance: big.NewInt(0).Set(bondSize),
		ValidatorIndex:     h.GetValidatorIndex(),
		// Default to 1 ETH of SP withdrawals
		SPWithdrawals: big.NewInt(1e18),
	}

	return out
}

type MockNode struct {
	Address                          common.Address
	RegistrationTime                 time.Time
	RplStake                         *big.Int
	SmoothingPoolRegistrationState   bool
	SmoothingPoolRegistrationChanged time.Time

	IsOdao       bool
	JoinedOdaoAt time.Time

	bondedEth   *big.Int
	borrowedEth *big.Int
	Minipools   []*MockMinipool

	Notes string
	Class string
}

func (n *MockNode) AddMinipool(minipool *MockMinipool) {
	minipool.NodeAddress = n.Address
	n.bondedEth.Add(n.bondedEth, minipool.NodeDepositBalance)
	borrowedEth := big.NewInt(0).Sub((*big.Int)(_bondSizeThirtyTwoEth), minipool.NodeDepositBalance)
	n.borrowedEth.Add(n.borrowedEth, borrowedEth)

	n.Minipools = append(n.Minipools, minipool)
}

func (h *MockHistory) SetWithdrawals(mockBeaconClient *MockBeaconClient) {
	for _, node := range h.Nodes {
		var slotWhileIn uint64
		// Get a slot inside the node's SP period
		if node.SmoothingPoolRegistrationState {
			// Use the last slot of the SP period
			slotWhileIn = h.BeaconConfig.LastSlotOfEpoch(h.EndEpoch)
		} else {
			// Get the opt-out time and ensure the node was opted in for at least 1 slot
			optedOut := node.SmoothingPoolRegistrationChanged
			if optedOut.Unix() != 0 {
				slotWhileIn = h.BeaconConfig.FirstSlotAtLeast(optedOut.Unix()) - 1
				if slotWhileIn < h.BeaconConfig.FirstSlotOfEpoch(h.StartEpoch) {
					slotWhileIn = 0
				}
			}
		}

		var slotWhileOut uint64
		if !node.SmoothingPoolRegistrationState {
			slotWhileOut = h.BeaconConfig.LastSlotOfEpoch(h.EndEpoch)
		} else {
			// Get the opt-in time and ensure the node was opted out for at least 1 slot
			optedIn := node.SmoothingPoolRegistrationChanged
			if optedIn.Unix() != 0 {
				slotWhileOut = h.BeaconConfig.FirstSlotAtLeast(optedIn.Unix()) - 1
				if slotWhileOut < h.BeaconConfig.FirstSlotOfEpoch(h.StartEpoch) {
					slotWhileOut = 0
				}
			}
		}

		for _, minipool := range node.Minipools {
			if minipool.SPWithdrawals != nil && minipool.SPWithdrawals.Sign() > 0 {
				if slotWhileIn == 0 {
					panic("minipool has sp withdrawals but node was never in the sp")
				}
				mockBeaconClient.AddWithdrawal(slotWhileIn, minipool.ValidatorIndex, minipool.SPWithdrawals)
			}
			if minipool.OptedOutWithdrawals != nil && minipool.OptedOutWithdrawals.Sign() > 0 {
				if slotWhileOut == 0 {
					panic("minipool has opted out withdrawals but node was never opted out of the sp")
				}
				mockBeaconClient.AddWithdrawal(slotWhileOut, minipool.ValidatorIndex, minipool.OptedOutWithdrawals)
			}
		}
	}
}

type NewMockNodeParams struct {
	SmoothingPool       bool
	EightEthMinipools   int
	SixteenEthMinipools int
	CollateralRpl       int64
}

func (h *MockHistory) GetNewDefaultMockNode(params *NewMockNodeParams) *MockNode {
	if params == nil {
		// Inefficient, but nice code follows.
		params = &NewMockNodeParams{}
	}

	out := &MockNode{
		Address:                          h.GetNodeAddress(),
		RegistrationTime:                 time.Unix(DefaultMockHistoryGenesis, 0),
		RplStake:                         big.NewInt(0),
		SmoothingPoolRegistrationState:   params.SmoothingPool,
		SmoothingPoolRegistrationChanged: time.Unix(0, 0),

		borrowedEth: big.NewInt(0),
		bondedEth:   big.NewInt(0),
	}

	for i := 0; i < params.EightEthMinipools; i++ {
		out.AddMinipool(h.GetNewDefaultMockMinipool(BondSizeEightEth))
	}

	for i := 0; i < params.SixteenEthMinipools; i++ {
		out.AddMinipool(h.GetNewDefaultMockMinipool(BondSizeSixteenEth))
	}

	out.RplStake = big.NewInt(params.CollateralRpl)
	out.RplStake.Mul(out.RplStake, eth.EthToWei(1))

	// Opt nodes in an epoch before the start of the interval
	if params.SmoothingPool {
		out.SmoothingPoolRegistrationChanged = h.BeaconConfig.GetSlotTime(h.BeaconConfig.FirstSlotOfEpoch(h.StartEpoch - 1))
	}

	return out
}

// Returns a list of nodes with various attributes-
// some will have active minipools, some will not.
// some will be under and over collateralized.
// some will have opted in or out during the interval
// some will have bond reduced during the interval
func (h *MockHistory) GetDefaultMockNodes() []*MockNode {
	nodes := []*MockNode{}

	// Create 10 nodes with one 8-eth minipool each and 10 RPL staked
	for i := 0; i < 10; i++ {
		node := h.GetNewDefaultMockNode(&NewMockNodeParams{
			EightEthMinipools: 1,
			CollateralRpl:     10,
		})
		node.Notes = "Regular node with one regular 8-eth minipool"
		node.Class = "single_eight_eth"
		node.Minipools[0].SPWithdrawals = nil
		node.Minipools[0].OptedOutWithdrawals = big.NewInt(1e18)
		nodes = append(nodes, node)
	}

	// Create 10 more of the same, but in the SP
	for i := 0; i < 10; i++ {
		node := h.GetNewDefaultMockNode(&NewMockNodeParams{
			EightEthMinipools: 1,
			SmoothingPool:     true,
			CollateralRpl:     10,
		})
		node.Notes = "Smoothing pool node with one regular 8-eth minipool"
		node.Class = "single_eight_eth_sp"
		nodes = append(nodes, node)
	}

	// Create 20 as above, but with 16-eth minipools
	for i := 0; i < 10; i++ {
		node := h.GetNewDefaultMockNode(&NewMockNodeParams{
			SixteenEthMinipools: 1,
			CollateralRpl:       10,
		})
		node.Notes = "Regular node with one regular 16-eth minipool"
		node.Class = "single_sixteen_eth"
		node.Minipools[0].SPWithdrawals = nil
		node.Minipools[0].OptedOutWithdrawals = big.NewInt(1e18)
		nodes = append(nodes, node)
	}

	for i := 0; i < 10; i++ {
		node := h.GetNewDefaultMockNode(&NewMockNodeParams{
			SixteenEthMinipools: 1,
			SmoothingPool:       true,
			CollateralRpl:       10,
		})
		node.Notes = "Smoothing pool node with one regular 16-eth minipool"
		node.Class = "single_sixteen_eth_sp"
		nodes = append(nodes, node)
	}

	// Add a node that opts in a quarter of the way through the interval
	node := h.GetNewDefaultMockNode(&NewMockNodeParams{
		EightEthMinipools: 1,
		SmoothingPool:     true,
		CollateralRpl:     20,
	})
	node.SmoothingPoolRegistrationChanged = h.BeaconConfig.GetSlotTime(h.BeaconConfig.FirstSlotOfEpoch(h.StartEpoch + (h.EndEpoch-h.StartEpoch)/4))
	node.Notes = "Smoothing pool node with one 8-eth that opts in 1/4 of the way through the interval"
	node.Class = "single_eight_eth_opted_in_quarter"
	nodes = append(nodes, node)

	// Add a node that opts in a quarter of the way through the interval
	node = h.GetNewDefaultMockNode(&NewMockNodeParams{
		SixteenEthMinipools: 1,
		SmoothingPool:       true,
		CollateralRpl:       20,
	})
	node.SmoothingPoolRegistrationChanged = h.BeaconConfig.GetSlotTime(h.BeaconConfig.FirstSlotOfEpoch(h.StartEpoch + (h.EndEpoch-h.StartEpoch)/4))
	node.Notes = "Smoothing pool node with one 16-eth that opts in 1/4 of the way through the interval"
	node.Class = "single_sixteen_eth_opted_in_quarter"
	nodes = append(nodes, node)

	// Add a node that opts out a three quarters of the way through the interval
	node = h.GetNewDefaultMockNode(&NewMockNodeParams{
		EightEthMinipools: 1,
		SmoothingPool:     false,
		CollateralRpl:     20,
	})
	node.SmoothingPoolRegistrationChanged = h.BeaconConfig.GetSlotTime(h.BeaconConfig.FirstSlotOfEpoch(h.StartEpoch + 3*(h.EndEpoch-h.StartEpoch)/4))
	node.Notes = "Smoothing pool node with one 8-eth that opts out 3/4 of the way through the interval"
	node.Class = "single_eight_eth_opted_out_three_quarters"
	nodes = append(nodes, node)

	// Add a node that opts out a three quarters of the way through the interval
	node = h.GetNewDefaultMockNode(&NewMockNodeParams{
		SixteenEthMinipools: 1,
		SmoothingPool:       false,
		CollateralRpl:       20,
	})
	node.SmoothingPoolRegistrationChanged = h.BeaconConfig.GetSlotTime(h.BeaconConfig.FirstSlotOfEpoch(h.StartEpoch + 3*(h.EndEpoch-h.StartEpoch)/4))
	node.Notes = "Smoothing pool node with one 16-eth that opts out 3/4 of the way through the interval"
	node.Class = "single_sixteen_eth_opted_out_three_quarters"
	nodes = append(nodes, node)

	// Add a node that does a bond reduction half way through the interval
	node = h.GetNewDefaultMockNode(&NewMockNodeParams{
		EightEthMinipools: 1,
		SmoothingPool:     true,
		CollateralRpl:     10,
	})
	node.Minipools[0].LastBondReductionTime = h.BeaconConfig.GetSlotTime(h.BeaconConfig.FirstSlotOfEpoch(h.StartEpoch + (h.EndEpoch-h.StartEpoch)/2))
	node.Minipools[0].LastBondReductionPrevValue = big.NewInt(0).Mul(big.NewInt(16), eth.EthToWei(1))
	// Say it was 20% for fun
	node.Minipools[0].LastBondReductionPrevNodeFee, _ = big.NewInt(0).SetString("200000000000000000", 10)
	node.Notes = "Node with one 16-eth that does a bond reduction to 8 eth halfway through the interval"
	node.Class = "single_bond_reduction"
	nodes = append(nodes, node)

	// Add a node with no minipools
	node = h.GetNewDefaultMockNode(&NewMockNodeParams{
		// Give it collateral so we can test that it's ignored despite having collateral
		CollateralRpl: 10,
	})
	node.Notes = "Node with no minipools but RPL collateral"
	node.Class = "no_minipools"
	nodes = append(nodes, node)

	// Add a node with a pending minipool
	node = h.GetNewDefaultMockNode(&NewMockNodeParams{
		EightEthMinipools: 1,
		CollateralRpl:     10,
	})
	node.Minipools[0].Status = types.Prelaunch
	node.Minipools[0].SPWithdrawals = nil
	node.Notes = "Node with one 8-eth minipool that is pending"
	node.Class = "single_eight_eth_pending"
	nodes = append(nodes, node)

	// Add a node with a single staking minipool that is finalized
	node = h.GetNewDefaultMockNode(&NewMockNodeParams{
		EightEthMinipools: 1,
		CollateralRpl:     10,
	})
	node.Minipools[0].Finalised = true
	node.Minipools[0].SPWithdrawals = nil
	node.Notes = "Node with one 8-eth minipool that is finalized"
	node.Class = "single_eight_eth_finalized"
	nodes = append(nodes, node)

	// Finally, create two odao nodes to share the juicy odao rewards
	odaoNodes := h.GetDefaultMockODAONodes()
	nodes = append(nodes, odaoNodes...)

	return nodes
}

func (h *MockHistory) GetDefaultMockODAONodes() []*MockNode {
	odaoNodes := []*MockNode{
		h.GetNewDefaultMockNode(nil),
		h.GetNewDefaultMockNode(nil),
	}
	for _, node := range odaoNodes {
		node.IsOdao = true
		node.Class = "odao"
	}
	return odaoNodes
}

const DefaultMockHistoryGenesis = 1577836800

type MockHistory struct {
	StartEpoch   uint64
	EndEpoch     uint64
	BlockOffset  uint64
	BeaconConfig beacon.Eth2Config

	// Network details for the final slot
	NetworkDetails *rpstate.NetworkDetails

	Nodes []*MockNode

	// Various offsets to create unique number spaces for each key type
	lastNodeAddress     common.Address
	lastMinipoolAddress common.Address
	lastValidatorPubkey types.ValidatorPubkey
	lastValidatorIndex  string
}

func NewDefaultMockHistoryNoNodes() *MockHistory {
	out := &MockHistory{
		StartEpoch:  100,
		EndEpoch:    200,
		BlockOffset: 100000,
		BeaconConfig: beacon.Eth2Config{
			GenesisEpoch: 0,
			// 2020-01-01 midnight UTC for simplicity
			GenesisTime:     DefaultMockHistoryGenesis,
			SlotsPerEpoch:   32,
			SecondsPerSlot:  12,
			SecondsPerEpoch: 12 * 32,
		},

		NetworkDetails: &rpstate.NetworkDetails{
			// Defaults to 0.24 ether, so 10 RPL is 2.4 ether and a leb8 with 10 RPL is 10% collateralized
			RplPrice: big.NewInt(240000000000000000),
			// Defaults to 10% aka 0.1 ether
			MinCollateralFraction: big.NewInt(100000000000000000),
			// Defaults to 60% to mimic current withdrawal limits
			MaxCollateralFraction: big.NewInt(600000000000000000),
			// Defaults to 100 epochs
			IntervalDuration: 100 * 32 * 12 * time.Second,
			// Defaults to genesis plus 100 epochs
			IntervalStart: time.Unix(DefaultMockHistoryGenesis, 0).Add(100 * 32 * 12 * time.Second),
			// Defaults to 0.7 ether to match mainnet
			NodeOperatorRewardsPercent: big.NewInt(700000000000000000),
			// Defaults to 0.015 ether to match mainnet as of 2024-10-08
			TrustedNodeOperatorRewardsPercent: big.NewInt(15000000000000000),
			// Defaults to 1 - 0.7 - 0.015 ether to round out to 100%
			ProtocolDaoRewardsPercent: big.NewInt(285000000000000000),
			// Defaults to 70,000 ether of RPL to apprixmate 1/13th of 5% of 18m
			PendingRPLRewards: big.NewInt(0).Mul(big.NewInt(70000), big.NewInt(1000000000000000000)),
			// RewardIndex defaults to 40000 to avoid a test tree from being taken seriously
			RewardIndex: 40000,
			// Put 100 ether in the smoothing pool
			SmoothingPoolBalance: big.NewInt(0).Mul(big.NewInt(100), big.NewInt(1000000000000000000)),

			// The rest of the fields seem unimportant and are left empty
		},
		lastNodeAddress:     common.BigToAddress(big.NewInt(2000)),
		lastMinipoolAddress: common.BigToAddress(big.NewInt(30000)),
		lastValidatorPubkey: types.BytesToValidatorPubkey(big.NewInt(600000).Bytes()),
		lastValidatorIndex:  "0",
	}
	return out
}

func NewDefaultMockHistory() *MockHistory {
	out := NewDefaultMockHistoryNoNodes()
	out.Nodes = out.GetDefaultMockNodes()
	return out
}

func (h *MockHistory) GetEndNetworkState() *state.NetworkState {
	out := &state.NetworkState{
		// El block number is the final slot's block, which is the last slot of the last epoch
		// plus the offset
		ElBlockNumber:              h.BlockOffset + h.BeaconConfig.LastSlotOfEpoch(h.EndEpoch),
		BeaconSlotNumber:           h.BeaconConfig.LastSlotOfEpoch(h.EndEpoch),
		BeaconConfig:               h.BeaconConfig,
		NetworkDetails:             h.NetworkDetails,
		NodeDetails:                []rpstate.NativeNodeDetails{},
		NodeDetailsByAddress:       make(map[common.Address]*rpstate.NativeNodeDetails),
		MinipoolDetails:            []rpstate.NativeMinipoolDetails{},
		MinipoolDetailsByAddress:   make(map[common.Address]*rpstate.NativeMinipoolDetails),
		MinipoolDetailsByNode:      make(map[common.Address][]*rpstate.NativeMinipoolDetails),
		MinipoolValidatorDetails:   make(state.ValidatorDetailsMap),
		OracleDaoMemberDetails:     []rpstate.OracleDaoMemberDetails{},
		ProtocolDaoProposalDetails: nil,
	}

	// Add nodes
	for _, node := range h.Nodes {
		// Calculate the node's effective RPL stake
		// If it's below 10% of borrowed eth per the network details, it's 0
		rplStake := node.RplStake
		rplPrice := h.NetworkDetails.RplPrice
		// Calculate the minimum RPL stake according to the network details
		minRplStake := big.NewInt(0).Mul(node.borrowedEth, h.NetworkDetails.MinCollateralFraction)
		// minRplStake is now the minimum RPL stake in eth value measured in wei squared
		// divide by the price to get the minimum RPL stake in RPL
		minRplStake.Div(minRplStake, rplPrice)

		// Same for max
		maxRplStake := big.NewInt(0).Mul(node.borrowedEth, h.NetworkDetails.MaxCollateralFraction)
		maxRplStake.Div(maxRplStake, rplPrice)

		// Eth matching limit is rpl stake times the price divided by the collateral fraction
		ethMatchingLimit := big.NewInt(0).Mul(node.RplStake, rplPrice)
		ethMatchingLimit.Div(ethMatchingLimit, h.NetworkDetails.MinCollateralFraction)
		collateralisationRatio := big.NewInt(0)
		if node.borrowedEth.Sign() > 0 {
			collateralisationRatio.Div(node.bondedEth, big.NewInt(0).Add(big.NewInt(0).Mul(node.bondedEth, eth.EthToWei(1)), node.borrowedEth))
		}

		// Create the node details
		details := rpstate.NativeNodeDetails{
			Exists:            true,
			RegistrationTime:  big.NewInt(node.RegistrationTime.Unix()),
			TimezoneLocation:  "UTC",
			RewardNetwork:     big.NewInt(0),
			RplStake:          node.RplStake,
			EffectiveRPLStake: rplStake,
			MinimumRPLStake:   minRplStake,
			MaximumRPLStake:   maxRplStake,
			EthMatched:        node.borrowedEth,
			EthMatchedLimit:   ethMatchingLimit,
			MinipoolCount:     big.NewInt(int64(len(node.Minipools))),
			// Empty node wallet
			BalanceETH:                       big.NewInt(0),
			BalanceRETH:                      big.NewInt(0),
			BalanceRPL:                       big.NewInt(0),
			BalanceOldRPL:                    big.NewInt(0),
			DepositCreditBalance:             big.NewInt(0),
			DistributorBalance:               big.NewInt(0),
			DistributorBalanceUserETH:        big.NewInt(0),
			DistributorBalanceNodeETH:        big.NewInt(0),
			WithdrawalAddress:                node.Address,
			PendingWithdrawalAddress:         common.Address{},
			SmoothingPoolRegistrationState:   node.SmoothingPoolRegistrationState,
			SmoothingPoolRegistrationChanged: big.NewInt(node.SmoothingPoolRegistrationChanged.Unix()),
			NodeAddress:                      node.Address,

			AverageNodeFee: big.NewInt(0), // Populated by CalculateAverageFeeAndDistributorShares

			// Ratio of bonded to bonded plus borrowed
			CollateralisationRatio: collateralisationRatio,
		}

		out.NodeDetails = append(out.NodeDetails, details)
		ptr := &out.NodeDetails[len(out.NodeDetails)-1]
		out.NodeDetailsByAddress[node.Address] = ptr

		// Add minipools
		for _, minipool := range node.Minipools {
			minipoolDetails := rpstate.NativeMinipoolDetails{
				Exists:                  true,
				MinipoolAddress:         minipool.Address,
				Pubkey:                  minipool.Pubkey,
				StatusRaw:               uint8(minipool.Status),
				StatusBlock:             minipool.StatusBlock,
				StatusTime:              big.NewInt(minipool.StatusTime.Unix()),
				Finalised:               minipool.Finalised,
				NodeFee:                 minipool.NodeFee,
				NodeDepositBalance:      minipool.NodeDepositBalance,
				NodeDepositAssigned:     true,
				UserDepositBalance:      big.NewInt(0).Sub(_bondSizeThirtyTwoEth, minipool.NodeDepositBalance),
				UserDepositAssigned:     true,
				UserDepositAssignedTime: big.NewInt(h.BeaconConfig.GetSlotTime(minipool.StatusBlock.Uint64() - h.BlockOffset).Unix()),
				NodeAddress:             minipool.NodeAddress,
				Balance:                 big.NewInt(0),
				DistributableBalance:    big.NewInt(0),
				NodeShareOfBalance:      big.NewInt(0),
				UserShareOfBalance:      big.NewInt(0),
				NodeRefundBalance:       big.NewInt(0),
				PenaltyCount:            big.NewInt(0),
				PenaltyRate:             big.NewInt(0),
				WithdrawalCredentials:   common.Hash{},
				Status:                  minipool.Status,
				DepositType:             types.Variable,

				LastBondReductionTime:        big.NewInt(minipool.LastBondReductionTime.Unix()),
				LastBondReductionPrevValue:   minipool.LastBondReductionPrevValue,
				LastBondReductionPrevNodeFee: minipool.LastBondReductionPrevNodeFee,
			}
			out.MinipoolDetails = append(out.MinipoolDetails, minipoolDetails)
			minipoolPtr := &out.MinipoolDetails[len(out.MinipoolDetails)-1]
			out.MinipoolDetailsByAddress[minipool.Address] = minipoolPtr
			out.MinipoolDetailsByNode[minipool.NodeAddress] = append(out.MinipoolDetailsByNode[minipool.NodeAddress], minipoolPtr)

			// Finally, populate The ValidatorDetails map
			pubkey := minipool.Pubkey
			withdrawableEpoch := FarFutureEpoch
			exitEpoch := FarFutureEpoch
			if minipool.WithdrawableEpoch != 0 {
				withdrawableEpoch = minipool.WithdrawableEpoch
				exitEpoch = minipool.WithdrawableEpoch - 1
			}
			details := beacon.ValidatorStatus{
				Pubkey:                     minipool.Pubkey,
				Index:                      minipool.ValidatorIndex,
				WithdrawalCredentials:      common.Hash{},
				Balance:                    (*big.Int)(_bondSizeThirtyTwoEth).Uint64(),
				EffectiveBalance:           (*big.Int)(_bondSizeThirtyTwoEth).Uint64(),
				Slashed:                    false,
				ActivationEligibilityEpoch: 0,
				ActivationEpoch:            0,
				ExitEpoch:                  exitEpoch,
				WithdrawableEpoch:          withdrawableEpoch,
				Exists:                     true,
			}
			if minipool.Status == types.Staking {
				details.Status = beacon.ValidatorState_ActiveOngoing
			}
			if minipool.Finalised {
				details.Status = beacon.ValidatorState_WithdrawalDone
			}
			out.MinipoolValidatorDetails[pubkey] = details
		}

		// Calculate the AverageNodeFee and DistributorShares
		ptr.CalculateAverageFeeAndDistributorShares(out.MinipoolDetailsByNode[ptr.NodeAddress])

		// Check if the node is an odao member
		if node.IsOdao {
			details := rpstate.OracleDaoMemberDetails{
				Address:          node.Address,
				Exists:           true,
				ID:               node.Address.Hex(),
				Url:              "https://example.com",
				JoinedTime:       time.Unix(node.RegistrationTime.Unix(), 0),
				LastProposalTime: time.Unix(node.RegistrationTime.Unix(), 0),
				RPLBondAmount:    node.RplStake,
			}
			out.OracleDaoMemberDetails = append(out.OracleDaoMemberDetails, details)
		}
	}

	return out
}

// Boring derived data getters
func (h *MockHistory) GetConsensusStartBlock() uint64 {
	return h.BeaconConfig.FirstSlotOfEpoch(h.StartEpoch)
}

func (h *MockHistory) GetExecutionStartBlock() uint64 {
	return h.GetConsensusStartBlock() + h.BlockOffset
}

func (h *MockHistory) GetConsensusEndBlock() uint64 {
	return h.BeaconConfig.LastSlotOfEpoch(h.EndEpoch)
}

func (h *MockHistory) GetExecutionEndBlock() uint64 {
	return h.GetConsensusEndBlock() + h.BlockOffset
}

func (h *MockHistory) GetStartTime() time.Time {
	return h.BeaconConfig.GetSlotTime(h.GetConsensusStartBlock())
}

func (h *MockHistory) GetEndTime() time.Time {
	return h.BeaconConfig.GetSlotTime(h.GetConsensusEndBlock())
}

func (h *MockHistory) GetPreviousRewardSnapshotEvent() rprewards.RewardsEvent {
	intervalEpochLength := h.EndEpoch - h.StartEpoch + 1
	consensusEndBlock := h.BeaconConfig.LastSlotOfEpoch(h.StartEpoch - 1)
	consensusStartBlock := consensusEndBlock - intervalEpochLength*h.BeaconConfig.SlotsPerEpoch
	return rprewards.RewardsEvent{
		Index:             big.NewInt(int64(h.NetworkDetails.RewardIndex - 1)),
		ExecutionBlock:    big.NewInt(int64(consensusEndBlock + h.BlockOffset)),
		ConsensusBlock:    big.NewInt(int64(consensusEndBlock)),
		MerkleRoot:        common.Hash{},
		MerkleTreeCID:     "",
		IntervalsPassed:   big.NewInt(1),
		TreasuryRPL:       big.NewInt(0),
		TrustedNodeRPL:    []*big.Int{},
		NodeRPL:           []*big.Int{},
		NodeETH:           []*big.Int{},
		UserETH:           big.NewInt(0),
		IntervalStartTime: h.BeaconConfig.GetSlotTime(consensusStartBlock),
		IntervalEndTime:   h.BeaconConfig.GetSlotTime(consensusEndBlock),
		SubmissionTime:    h.BeaconConfig.GetSlotTime(consensusEndBlock),
	}
}

func (h *MockHistory) GetNodeSummary() map[string][]*MockNode {
	out := make(map[string][]*MockNode)
	for _, node := range h.Nodes {
		out[node.Class] = append(out[node.Class], node)
	}
	return out
}
