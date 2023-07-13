package state

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	oDaoAddressBatchSize int = 1000
	oDaoDetailsBatchSize int = 50
)

type OracleDaoMemberDetails struct {
	Address             common.Address `json:"address"`
	Exists              bool           `json:"exists"`
	ID                  string         `json:"id"`
	Url                 string         `json:"url"`
	JoinedTime          time.Time      `json:"joinedTime"`
	LastProposalTime    time.Time      `json:"lastProposalTime"`
	RPLBondAmount       *big.Int       `json:"rplBondAmount"`
	ReplacementAddress  common.Address `json:"replacementAddress"`
	IsChallenged        bool           `json:"isChallenged"`
	joinedTimeRaw       *big.Int       `json:"-"`
	lastProposalTimeRaw *big.Int       `json:"-"`
}

// Gets the details for an Oracle DAO member using the efficient multicall contract
func GetOracleDaoMemberDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, memberAddress common.Address) (OracleDaoMemberDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	details := OracleDaoMemberDetails{}
	details.Address = memberAddress

	addOracleDaoMemberDetailsCalls(rp, contracts, contracts.Multicaller, &details, opts)

	_, err := contracts.Multicaller.FlexibleCall(true, opts)
	if err != nil {
		return OracleDaoMemberDetails{}, fmt.Errorf("error executing multicall: %w", err)
	}

	fixupOracleDaoMemberDetails(rp, &details, opts)

	return details, nil
}

// Gets all Oracle DAO member details using the efficient multicall contract
func GetAllOracleDaoMemberDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts) ([]OracleDaoMemberDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	// Get the list of all minipool addresses
	addresses, err := getOdaoAddresses(rp, contracts, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting Oracle DAO addresses: %w", err)
	}

	// Get the minipool details
	return getOracleDaoDetails(rp, contracts, addresses, opts)
}

// Get all Oracle DAO addresses
func getOdaoAddresses(rp *rocketpool.RocketPool, contracts *NetworkContracts, opts *bind.CallOpts) ([]common.Address, error) {
	// Get minipool count
	memberCount, err := trustednode.GetMemberCount(rp, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	addresses := make([]common.Address, memberCount)

	// Run the getters in batches
	count := int(memberCount)
	for i := 0; i < count; i += minipoolAddressBatchSize {
		i := i
		max := i + oDaoAddressBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				mc.AddCall(contracts.RocketDAONodeTrusted, &addresses[j], "getMemberAt", big.NewInt(int64(j)))
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting Oracle DAO addresses: %w", err)
	}

	return addresses, nil
}

// Get the details of the Oracle DAO members
func getOracleDaoDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, addresses []common.Address, opts *bind.CallOpts) ([]OracleDaoMemberDetails, error) {
	memberDetails := make([]OracleDaoMemberDetails, len(addresses))

	// Get the details in batches
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	count := len(addresses)
	for i := 0; i < count; i += minipoolBatchSize {
		i := i
		max := i + minipoolBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {

				address := addresses[j]
				details := &memberDetails[j]
				details.Address = address

				addOracleDaoMemberDetailsCalls(rp, contracts, mc, details, opts)
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting Oracle DAO details: %w", err)
	}

	// Postprocessing
	for i := range memberDetails {
		details := &memberDetails[i]
		fixupOracleDaoMemberDetails(rp, details, opts)
	}

	return memberDetails, nil
}

// Add the Oracle DAO details getters to the multicaller
func addOracleDaoMemberDetailsCalls(rp *rocketpool.RocketPool, contracts *NetworkContracts, mc *multicall.MultiCaller, details *OracleDaoMemberDetails, opts *bind.CallOpts) error {
	address := details.Address
	mc.AddCall(contracts.RocketDAONodeTrusted, &details.Exists, "getMemberIsValid", address)
	mc.AddCall(contracts.RocketDAONodeTrusted, &details.ID, "getMemberID", address)
	mc.AddCall(contracts.RocketDAONodeTrusted, &details.Url, "getMemberUrl", address)
	mc.AddCall(contracts.RocketDAONodeTrusted, &details.joinedTimeRaw, "getMemberJoinedTime", address)
	mc.AddCall(contracts.RocketDAONodeTrusted, &details.lastProposalTimeRaw, "getMemberLastProposalTime", address)
	mc.AddCall(contracts.RocketDAONodeTrusted, &details.RPLBondAmount, "getMemberRPLBondAmount", address)
	mc.AddCall(contracts.RocketDAONodeTrusted, &details.ReplacementAddress, "getMemberReplacedAddress", address)
	mc.AddCall(contracts.RocketDAONodeTrusted, &details.IsChallenged, "getMemberIsChallenged", address)
	return nil
}

// Fixes a member details struct with supplemental logic
func fixupOracleDaoMemberDetails(rp *rocketpool.RocketPool, details *OracleDaoMemberDetails, opts *bind.CallOpts) error {
	details.JoinedTime = convertToTime(details.joinedTimeRaw)
	details.LastProposalTime = convertToTime(details.lastProposalTimeRaw)
	return nil
}
