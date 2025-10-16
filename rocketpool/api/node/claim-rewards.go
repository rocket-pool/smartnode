package node

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	node131 "github.com/rocket-pool/smartnode/bindings/legacy/v1.3.1/node"
	rewards131 "github.com/rocket-pool/smartnode/bindings/legacy/v1.3.1/rewards"

	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rewards"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	updateCheck "github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

func getRewardsInfo(c *cli.Context) (*api.NodeGetRewardsInfoResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Check if Saturn is already deployed
	saturnDeployed, err := updateCheck.IsSaturnDeployed(rp, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeGetRewardsInfoResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	response.Registered, err = node.GetNodeExists(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	if !response.Registered {
		return &response, nil
	}

	// Get the claimed and unclaimed intervals
	unclaimed, claimed, err := rprewards.GetClaimStatus(rp, nodeAccount.Address)
	if err != nil {
		return nil, err
	}
	response.ClaimedIntervals = claimed

	// Get the info for each unclaimed interval
	for _, unclaimedInterval := range unclaimed {
		intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAccount.Address, unclaimedInterval, nil)
		if err != nil {
			return nil, err
		}
		if !intervalInfo.TreeFileExists || !intervalInfo.MerkleRootValid {
			response.InvalidIntervals = append(response.InvalidIntervals, intervalInfo)
			continue
		}
		if intervalInfo.NodeExists {
			response.UnclaimedIntervals = append(response.UnclaimedIntervals, intervalInfo)
		}
	}

	// Get collateral info for restaking
	var totalMinipools int
	var finalizedMinipools int
	details, err := getNodeMinipoolCountDetails(rp, nodeAccount.Address)
	if err == nil {
		totalMinipools = len(details)
		for _, mpDetails := range details {
			if mpDetails.Finalised {
				finalizedMinipools++
			}
		}
	}
	activeMinipools := totalMinipools - finalizedMinipools
	response.ActiveMinipools = activeMinipools

	// Sync
	var wg errgroup.Group

	if saturnDeployed {
		wg.Go(func() error {
			var err error
			response.RplStake, err = node.GetNodeStakedRPL(rp, nodeAccount.Address, nil)
			return err
		})

	} else {
		wg.Go(func() error {
			var err error
			response.RplStake, err = node131.GetNodeRPLStake(rp, nodeAccount.Address, nil)
			return err
		})
	}

	wg.Go(func() error {
		var err error
		response.RplPrice, err = network.GetRPLPrice(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	if activeMinipools > 0 {
		var wg2 errgroup.Group
		wg2.Go(func() error {
			var err error
			response.EthBorrowed, response.EthBorrowLimit, response.PendingBorrowAmount, err = rputils.CheckCollateral(saturnDeployed, rp, nodeAccount.Address, nil)
			return err
		})

		// Wait for data
		if err := wg2.Wait(); err != nil {
			return nil, err
		}

		response.BondedCollateralRatio = eth.WeiToEth(response.RplPrice) * eth.WeiToEth(response.RplStake) / (float64(activeMinipools)*32.0 - eth.WeiToEth(response.EthBorrowed) - eth.WeiToEth(response.PendingBorrowAmount))
		response.BorrowedCollateralRatio = eth.WeiToEth(response.RplPrice) * eth.WeiToEth(response.RplStake) / (eth.WeiToEth(response.EthBorrowed) + eth.WeiToEth(response.PendingBorrowAmount))
	} else {
		response.BorrowedCollateralRatio = -1
	}

	return &response, nil
}

func canClaimRewards(c *cli.Context, indicesString string) (*api.CanNodeClaimRewardsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanNodeClaimRewardsResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get the rewards
	claims, err := getRewardsForIntervals(rp, cfg, nodeAccount.Address, indicesString)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := rewards.EstimateClaimGas(rp, nodeAccount.Address, claims, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo
	return &response, nil
}

func claimRewards(c *cli.Context, indicesString string) (*api.NodeClaimRewardsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Check if Saturn is already deployed
	isSaturnDeployed, err := updateCheck.IsSaturnDeployed(rp, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeClaimRewardsResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	if !isSaturnDeployed {
		// Get the rewards
		indices, amountRPL, amountETH, merkleProofs, err := getRewardsForIntervalsHouston(rp, cfg, nodeAccount.Address, indicesString)
		if err != nil {
			return nil, err
		}
		hash, err := rewards131.Claim(rp, nodeAccount.Address, indices, amountRPL, amountETH, merkleProofs, opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash
	} else {
		// Get the rewards
		claims, err := getRewardsForIntervals(rp, cfg, nodeAccount.Address, indicesString)
		if err != nil {
			return nil, err
		}
		hash, err := rewards.Claim(rp, nodeAccount.Address, claims, opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash
	}

	// Return response
	return &response, nil

}

func canClaimAndStakeRewards(c *cli.Context, indicesString string, stakeAmount *big.Int) (*api.CanNodeClaimAndStakeRewardsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Check if Saturn is already deployed
	isSaturnDeployed, err := updateCheck.IsSaturnDeployed(rp, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanNodeClaimAndStakeRewardsResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the rewards
	if !isSaturnDeployed {
		indices, amountRPL, amountETH, merkleProofs, err := getRewardsForIntervalsHouston(rp, cfg, nodeAccount.Address, indicesString)
		if err != nil {
			return nil, err
		}
		gasInfo, err := rewards131.EstimateClaimAndStakeGas(rp, nodeAccount.Address, indices, amountRPL, amountETH, merkleProofs, stakeAmount, opts)
		if err != nil {
			return nil, err
		}
		response.GasInfo = gasInfo
	} else {
		claims, err := getRewardsForIntervals(rp, cfg, nodeAccount.Address, indicesString)
		if err != nil {
			return nil, err
		}
		gasInfo, err := rewards.EstimateClaimAndStakeGas(rp, nodeAccount.Address, claims, stakeAmount, opts)
		if err != nil {
			return nil, err
		}
		response.GasInfo = gasInfo
	}

	return &response, nil

}

func claimAndStakeRewards(c *cli.Context, indicesString string, stakeAmount *big.Int) (*api.NodeClaimAndStakeRewardsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Check if Saturn is already deployed
	isSaturnDeployed, err := updateCheck.IsSaturnDeployed(rp, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeClaimAndStakeRewardsResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Get the rewards
	if !isSaturnDeployed {
		indices, amountRPL, amountETH, merkleProofs, err := getRewardsForIntervalsHouston(rp, cfg, nodeAccount.Address, indicesString)
		if err != nil {
			return nil, err
		}
		hash, err := rewards131.ClaimAndStake(rp, nodeAccount.Address, indices, amountRPL, amountETH, merkleProofs, stakeAmount, opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash
	} else {
		claims, err := getRewardsForIntervals(rp, cfg, nodeAccount.Address, indicesString)
		if err != nil {
			return nil, err
		}
		hash, err := rewards.ClaimAndStake(rp, nodeAccount.Address, claims, stakeAmount, opts)
		if err != nil {
			return nil, err
		}
		response.TxHash = hash
	}

	// Return response
	return &response, nil

}

// Get the rewards for the provided interval indices returning the parameters needed for Houston (Separate function so it's easier to remove after the upgrade)
func getRewardsForIntervalsHouston(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, nodeAddress common.Address, indicesString string) ([]*big.Int, []*big.Int, []*big.Int, [][]common.Hash, error) {

	// Get the indices
	seenIndices := map[uint64]bool{}
	elements := strings.Split(indicesString, ",")
	indices := []*big.Int{}
	for _, element := range elements {
		index, err := strconv.ParseUint(element, 0, 64)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("cannot convert index %s to a number: %w", element, err)
		}
		// Ignore duplicates
		_, exists := seenIndices[index]
		if !exists {
			indices = append(indices, big.NewInt(0).SetUint64(index))
			seenIndices[index] = true
		}
	}

	// Read the tree files to get the details
	amountRPL := []*big.Int{}
	amountETH := []*big.Int{}
	merkleProofs := [][]common.Hash{}

	// Populate the interval info for each one
	for _, index := range indices {

		intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAddress, index.Uint64(), nil)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		// Validate
		if !intervalInfo.TreeFileExists {
			return nil, nil, nil, nil, fmt.Errorf("rewards tree file '%s' doesn't exist", intervalInfo.TreeFilePath)
		}
		if !intervalInfo.MerkleRootValid {
			return nil, nil, nil, nil, fmt.Errorf("merkle root for rewards tree file '%s' doesn't match the canonical merkle root for interval %d", intervalInfo.TreeFilePath, index.Uint64())
		}

		// Get the rewards from it
		if intervalInfo.NodeExists {
			rplForInterval := big.NewInt(0)
			rplForInterval.Add(rplForInterval, &intervalInfo.CollateralRplAmount.Int)
			rplForInterval.Add(rplForInterval, &intervalInfo.ODaoRplAmount.Int)

			ethForInterval := big.NewInt(0)
			ethForInterval.Add(ethForInterval, &intervalInfo.SmoothingPoolEthAmount.Int)

			amountRPL = append(amountRPL, rplForInterval)
			amountETH = append(amountETH, ethForInterval)
			merkleProofs = append(merkleProofs, intervalInfo.MerkleProof)
		}
	}

	// Return
	return indices, amountRPL, amountETH, merkleProofs, nil

}

// Get the rewards for the provided interval indices
func getRewardsForIntervals(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, nodeAddress common.Address, indicesString string) (types.Claims, error) {

	// Get the indices
	seenIndices := map[uint64]bool{}
	elements := strings.Split(indicesString, ",")
	indices := []*big.Int{}
	for _, element := range elements {
		index, err := strconv.ParseUint(element, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert index %s to a number: %w", element, err)
		}
		// Ignore duplicates
		_, exists := seenIndices[index]
		if !exists {
			indices = append(indices, big.NewInt(0).SetUint64(index))
			seenIndices[index] = true
		}
	}

	// Read the tree files to get the details
	claims := types.Claims{}

	// Populate the interval info for each one
	for _, index := range indices {

		intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAddress, index.Uint64(), nil)
		if err != nil {
			return nil, err
		}

		// Validate
		if !intervalInfo.TreeFileExists {
			return nil, fmt.Errorf("rewards tree file '%s' doesn't exist", intervalInfo.TreeFilePath)
		}
		if !intervalInfo.MerkleRootValid {
			return nil, fmt.Errorf("merkle root for rewards tree file '%s' doesn't match the canonical merkle root for interval %d", intervalInfo.TreeFilePath, index.Uint64())
		}

		// Get the rewards from it
		if intervalInfo.NodeExists {
			rplForInterval := big.NewInt(0)
			rplForInterval.Add(rplForInterval, &intervalInfo.CollateralRplAmount.Int)
			rplForInterval.Add(rplForInterval, &intervalInfo.ODaoRplAmount.Int)

			smoothingEthForInterval := big.NewInt(0)
			smoothingEthForInterval.Add(smoothingEthForInterval, &intervalInfo.SmoothingPoolEthAmount.Int)

			voterShareEthForInterval := big.NewInt(0)
			voterShareEthForInterval.Add(voterShareEthForInterval, &intervalInfo.VoterShareEth.Int)

			claims = append(claims, types.Claim{
				Index:               index,
				AmountRPL:           rplForInterval,
				AmountSmoothingETH:  smoothingEthForInterval,
				AmountVoterShareETH: voterShareEthForInterval,
				Proof:               intervalInfo.MerkleProof,
			})
		}
	}

	// Return
	return claims, nil

}
