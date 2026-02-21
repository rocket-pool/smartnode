package state

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	megapoolValidatorsBatchSize int = 200
	megapoolBatchSize           int = 100
)

type NativeMegapoolDetails struct {
	Address                  common.Address `json:"address"`
	DelegateAddress          common.Address `json:"delegate"`
	EffectiveDelegateAddress common.Address `json:"effectiveDelegateAddress"`
	Deployed                 bool           `json:"deployed"`
	ValidatorCount           uint32         `json:"validatorCount"`
	ActiveValidatorCount     uint32         `json:"activeValidatorCount"`
	LockedValidatorCount     uint32         `json:"lockedValidatorCount"`
	NodeDebt                 *big.Int       `json:"nodeDebt"`
	RefundValue              *big.Int       `json:"refundValue"`
	DelegateExpiry           uint64         `json:"delegateExpiry"`
	DelegateExpired          bool           `json:"delegateExpired"`
	NodeExpressTicketCount   uint64         `json:"nodeExpressTicketCount"`
	UseLatestDelegate        bool           `json:"useLatestDelegate"`
	AssignedValue            *big.Int       `json:"assignedValue"`
	NodeBond                 *big.Int       `json:"nodeBond"`
	UserCapital              *big.Int       `json:"userCapital"`
	BondRequirement          *big.Int       `json:"bondRequirement"`
	EthBalance               *big.Int       `json:"ethBalance"`
	LastDistributionTime     uint64         `json:"lastDistributionTime"`
	PendingRewards           *big.Int       `json:"pendingRewards"`
	NodeQueuedBond           *big.Int       `json:"nodeQueuedBond"`
}

// Get the normalized bond per 32 eth validator
// This is used in treegen to calculate attestation scores
func (m *NativeMegapoolDetails) GetMegapoolBondNormalized() *big.Int {
	if m.ActiveValidatorCount == 0 {
		return big.NewInt(0)
	}
	return big.NewInt(0).Div(m.NodeBond, big.NewInt(int64(m.ActiveValidatorCount)))
}

// Get all megapool validators using batched multicalls
func GetAllMegapoolValidators(rp *rocketpool.RocketPool, contracts *NetworkContracts) ([]megapool.ValidatorInfoFromGlobalIndex, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	if contracts.Multicaller == nil {
		return nil, fmt.Errorf("multicaller is nil")
	}
	if contracts.RocketMegapoolManager == nil {
		return nil, fmt.Errorf("RocketMegapoolManager contract is nil")
	}

	// Capture values from contracts before launching goroutines
	multicallerAddress := contracts.Multicaller.ContractAddress
	megapoolManagerContract := contracts.RocketMegapoolManager

	megapoolValidatorsCount, err := megapool.GetValidatorCount(rp, opts)
	if err != nil {
		return nil, err
	}

	count := int(megapoolValidatorsCount)
	validators := make([]megapool.ValidatorInfoFromGlobalIndex, count)

	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	for i := 0; i < count; i += megapoolValidatorsBatchSize {
		i := i
		max := min(i+megapoolValidatorsBatchSize, count)

		wg.Go(func() error {
			mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
			if err != nil {
				return err
			}
			var dummy *big.Int
			for j := i; j < max; j++ {
				mc.AddCall(megapoolManagerContract, &dummy, "getValidatorInfo", big.NewInt(int64(j)))
			}
			responses, err := mc.Execute(true, opts)
			if err != nil {
				return fmt.Errorf("error executing megapool validator multicall: %w", err)
			}
			for idx, response := range responses {
				if !response.Status {
					return fmt.Errorf("megapool validator call failed for global index %d", i+idx)
				}
				validators[i+idx], err = unpackValidatorInfoFromGlobalIndex(megapoolManagerContract, response.ReturnDataRaw)
				if err != nil {
					return fmt.Errorf("error unpacking validator info for global index %d: %w", i+idx, err)
				}
			}
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting megapool validators: %w", err)
	}

	return validators, nil
}

// Manually unpack a getValidatorInfo response (nested structs don't work with UnpackIntoInterface)
func unpackValidatorInfoFromGlobalIndex(contract *rocketpool.Contract, data []byte) (megapool.ValidatorInfoFromGlobalIndex, error) {
	iface, err := contract.ABI.Unpack("getValidatorInfo", data)
	if err != nil {
		return megapool.ValidatorInfoFromGlobalIndex{}, err
	}

	src := iface[1].(struct {
		LastAssignmentTime uint32 `json:"lastAssignmentTime"`
		LastRequestedValue uint32 `json:"lastRequestedValue"`
		LastRequestedBond  uint32 `json:"lastRequestedBond"`
		DepositValue       uint32 `json:"depositValue"`

		Staked      bool `json:"staked"`
		Exited      bool `json:"exited"`
		InQueue     bool `json:"inQueue"`
		InPrestake  bool `json:"inPrestake"`
		ExpressUsed bool `json:"expressUsed"`
		Dissolved   bool `json:"dissolved"`
		Exiting     bool `json:"exiting"`
		Locked      bool `json:"locked"`

		ExitBalance uint64 `json:"exitBalance"`
		LockedTime  uint64 `json:"lockedTime"`
	})

	var validator megapool.ValidatorInfoFromGlobalIndex
	validator.Pubkey = iface[0].([]byte)
	validator.ValidatorInfo.LastAssignmentTime = src.LastAssignmentTime
	validator.ValidatorInfo.LastRequestedValue = src.LastRequestedValue
	validator.ValidatorInfo.LastRequestedBond = src.LastRequestedBond
	validator.ValidatorInfo.DepositValue = src.DepositValue
	validator.ValidatorInfo.Staked = src.Staked
	validator.ValidatorInfo.Exited = src.Exited
	validator.ValidatorInfo.InQueue = src.InQueue
	validator.ValidatorInfo.InPrestake = src.InPrestake
	validator.ValidatorInfo.ExpressUsed = src.ExpressUsed
	validator.ValidatorInfo.Dissolved = src.Dissolved
	validator.ValidatorInfo.Exiting = src.Exiting
	validator.ValidatorInfo.Locked = src.Locked
	validator.ValidatorInfo.ExitBalance = src.ExitBalance
	validator.ValidatorInfo.LockedTime = src.LockedTime
	validator.MegapoolAddress = iface[2].(common.Address)
	validator.ValidatorId = iface[3].(uint32)

	return validator, nil
}

// Get multiple megapool details at once using batched multicalls
func GetBulkMegapoolDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, megapoolAddresses []common.Address) (map[common.Address]NativeMegapoolDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	count := len(megapoolAddresses)
	result := make(map[common.Address]NativeMegapoolDetails, count)
	if count == 0 {
		return result, nil
	}

	if contracts.Multicaller == nil {
		return nil, fmt.Errorf("multicaller is nil")
	}

	// Capture values from contracts before launching goroutines
	multicallerAddress := contracts.Multicaller.ContractAddress
	megapoolFactory := contracts.RocketMegapoolFactory
	nodeDeposit := contracts.RocketNodeDeposit

	megapoolDetails := make([]NativeMegapoolDetails, count)
	lastDistributionTimes := make([]*big.Int, count)
	delegateExpiries := make([]*big.Int, count)

	// Get all ETH balances in a single batch
	balances, err := contracts.BalanceBatcher.GetEthBalances(megapoolAddresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting megapool balances: %w", err)
	}

	// Create megapool contract bindings (ABI is cached after the first call)
	megaContracts := make([]*rocketpool.Contract, count)
	for i, addr := range megapoolAddresses {
		megapoolDetails[i].Address = addr
		megapoolDetails[i].Deployed = true
		megapoolDetails[i].EthBalance = balances[i]

		mega, err := megapool.NewMegaPoolV1(rp, addr, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating megapool contract for %s: %w", addr.Hex(), err)
		}
		megaContracts[i] = mega.GetContract()
	}

	// Round 1: all independent fields from every megapool in batched multicalls
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	for i := 0; i < count; i += megapoolBatchSize {
		i := i
		max := min(i+megapoolBatchSize, count)

		wg.Go(func() error {
			mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				addMegapoolDetailsCalls(mc, megaContracts[j], &megapoolDetails[j], &lastDistributionTimes[j])
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing megapool r1 multicall: %w", err)
			}
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting megapool details r1: %w", err)
	}

	// Convert intermediate big.Int values from round 1
	for i := range megapoolDetails {
		if lastDistributionTimes[i] != nil {
			megapoolDetails[i].LastDistributionTime = lastDistributionTimes[i].Uint64()
		}
	}

	// Round 2: dependent fields (delegate expiry from DelegateAddress, bond requirement from ActiveValidatorCount)
	var wg2 errgroup.Group
	wg2.SetLimit(threadLimit)
	for i := 0; i < count; i += megapoolBatchSize {
		i := i
		max := min(i+megapoolBatchSize, count)

		wg2.Go(func() error {
			mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
			if err != nil {
				return err
			}
			callCount := 0
			for j := i; j < max; j++ {
				if megapoolDetails[j].DelegateExpired {
					continue
				}
				mc.AddCall(megapoolFactory, &delegateExpiries[j], "getDelegateExpiry", megapoolDetails[j].DelegateAddress)
				mc.AddCall(nodeDeposit, &megapoolDetails[j].BondRequirement, "getBondRequirement", big.NewInt(int64(megapoolDetails[j].ActiveValidatorCount)))
				callCount += 2
			}
			if callCount == 0 {
				return nil
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing megapool r2 multicall: %w", err)
			}
			return nil
		})
	}
	if err := wg2.Wait(); err != nil {
		return nil, fmt.Errorf("error getting megapool details r2: %w", err)
	}

	// Convert intermediate values and build result map
	for i := range megapoolDetails {
		if delegateExpiries[i] != nil {
			megapoolDetails[i].DelegateExpiry = delegateExpiries[i].Uint64()
		}
		result[megapoolDetails[i].Address] = megapoolDetails[i]
	}

	return result, nil
}

// Add all independent multicall entries for a single megapool's details
func addMegapoolDetailsCalls(mc *multicall.MultiCaller, megaContract *rocketpool.Contract, details *NativeMegapoolDetails, lastDistributionTime **big.Int) {
	mc.AddCall(megaContract, &details.EffectiveDelegateAddress, "getEffectiveDelegate")
	mc.AddCall(megaContract, &details.DelegateAddress, "getDelegate")
	mc.AddCall(megaContract, &details.DelegateExpired, "getDelegateExpired")
	mc.AddCall(megaContract, lastDistributionTime, "getLastDistributionTime")
	mc.AddCall(megaContract, &details.NodeDebt, "getDebt")
	mc.AddCall(megaContract, &details.PendingRewards, "getPendingRewards")
	mc.AddCall(megaContract, &details.RefundValue, "getRefundValue")
	mc.AddCall(megaContract, &details.ValidatorCount, "getValidatorCount")
	mc.AddCall(megaContract, &details.ActiveValidatorCount, "getActiveValidatorCount")
	mc.AddCall(megaContract, &details.LockedValidatorCount, "getLockedValidatorCount")
	mc.AddCall(megaContract, &details.UseLatestDelegate, "getUseLatestDelegate")
	mc.AddCall(megaContract, &details.AssignedValue, "getAssignedValue")
	mc.AddCall(megaContract, &details.NodeBond, "getNodeBond")
	mc.AddCall(megaContract, &details.UserCapital, "getUserCapital")
	mc.AddCall(megaContract, &details.NodeQueuedBond, "getNodeQueuedBond")
}
