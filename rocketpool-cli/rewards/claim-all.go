package claims

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	rocketpoolapi "github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/urfave/cli"
)

const (
	colorReset  string = "\033[0m"
	colorRed    string = "\033[31m"
	colorGreen  string = "\033[32m"
	colorYellow string = "\033[33m"
	colorBlue   string = "\033[36m"
)

// pendingClaim represents a single category of rewards that can be claimed.
type pendingClaim struct {
	id      int
	name    string
	gasInfo rocketpoolapi.GasInfo
	execute func() error
}

func claimAll(c *cli.Context, statusOnly bool) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	autoConfirm := c.Bool("yes")

	// Check Saturn deployment
	isSaturn := false
	saturnResp, err := rp.IsSaturnDeployed()
	if err == nil && saturnResp.IsSaturnDeployed {
		isSaturn = true
	}

	// Track totals
	totalEthWei := new(big.Int)
	totalRplWei := new(big.Int)
	sectionID := 0

	// Collect claims for the execution phase
	var claims []pendingClaim

	// Periodic rewards restake tracking (resolved after claim selection)
	var periodicRestakeAmount *big.Int
	var periodicClaimRpl *big.Int
	var periodicIntervalIndices []uint64
	periodicRestakeResolved := false

	fmt.Printf("%s============================================================%s\n", colorGreen, colorReset)
	fmt.Printf("%s              Available Rewards Summary                      %s\n", colorGreen, colorReset)
	fmt.Printf("%s============================================================%s\n\n", colorGreen, colorReset)

	// ================================================================
	// 1. Megapool EL Rewards (distribute)
	// ================================================================
	if isSaturn {
		sectionID++
		id := sectionID
		fmt.Printf("%s--- [%d] Megapool Execution Layer Rewards ---%s\n", colorGreen, id, colorReset)

		canDistribute, err := rp.CanDistributeMegapool()
		if err != nil {
			fmt.Printf("  %sCould not check megapool: %s%s\n\n", colorYellow, err, colorReset)
		} else if !canDistribute.CanDistribute {
			if canDistribute.MegapoolNotDeployed {
				fmt.Printf("  No megapool deployed.\n\n")
			} else if canDistribute.LastDistributionTime == 0 {
				fmt.Printf("  No staking validators in the megapool.\n\n")
			} else {
				reasons := []string{}
				if canDistribute.ExitingValidatorCount > 0 {
					reasons = append(reasons, fmt.Sprintf("%d validator(s) exiting", canDistribute.ExitingValidatorCount))
				}
				if canDistribute.LockedValidatorCount > 0 {
					reasons = append(reasons, fmt.Sprintf("%d validator(s) locked", canDistribute.LockedValidatorCount))
				}
				if len(reasons) > 0 {
					fmt.Printf("  Cannot distribute: %s\n\n", strings.Join(reasons, ", "))
				} else {
					fmt.Printf("  Cannot distribute at this time.\n\n")
				}
			}
		} else {
			// Get the pending rewards breakdown
			pendingRewards, err := rp.CalculatePendingRewards()
			if err != nil {
				fmt.Printf("  %sCould not calculate pending rewards: %s%s\n\n", colorYellow, err, colorReset)
			} else {
				megapoolTotal := new(big.Int).Add(pendingRewards.RewardSplit.NodeRewards, pendingRewards.RefundValue)
				if megapoolTotal.Cmp(big.NewInt(0)) > 0 {
					fmt.Printf("  Node share:    %.6f ETH\n", math.RoundDown(eth.WeiToEth(pendingRewards.RewardSplit.NodeRewards), 6))
					if pendingRewards.RefundValue.Cmp(big.NewInt(0)) > 0 {
						fmt.Printf("  Refund value:  %.6f ETH\n", math.RoundDown(eth.WeiToEth(pendingRewards.RefundValue), 6))
					}
					fmt.Printf("  Total:         %.6f ETH\n\n", math.RoundDown(eth.WeiToEth(megapoolTotal), 6))
					totalEthWei.Add(totalEthWei, megapoolTotal)

					gasInfo := canDistribute.GasInfo
					claims = append(claims, pendingClaim{
						id:      id,
						name:    "Megapool EL Rewards (distribute)",
						gasInfo: gasInfo,
						execute: func() error {
							fmt.Printf("  Submitting transaction...\n")
							response, err := rp.DistributeMegapool()
							if err != nil {
								return fmt.Errorf("transaction could not be submitted: %w", err)
							}
							fmt.Printf("  Distributing megapool rewards...\n")
							cliutils.PrintTransactionHash(rp, response.TxHash)
							if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
								return fmt.Errorf("transaction was submitted but failed onchain: %w", err)
							}
							fmt.Printf("  %sSuccessfully distributed megapool rewards.%s\n", colorGreen, colorReset)
							return nil
						},
					})
				} else {
					fmt.Printf("  No pending rewards to distribute.\n\n")
				}
			}
		}
	}

	// ================================================================
	// 2. Fee Distributor (distribute)
	// ================================================================
	sectionID++
	feeDistID := sectionID
	fmt.Printf("%s--- [%d] Fee Distributor ---%s\n", colorGreen, feeDistID, colorReset)

	isInitResponse, err := rp.IsFeeDistributorInitialized()
	if err != nil {
		fmt.Printf("  %sCould not check fee distributor: %s%s\n\n", colorYellow, err, colorReset)
	} else if !isInitResponse.IsInitialized {
		fmt.Printf("  Fee distributor not initialized. Run 'rocketpool node initialize-fee-distributor' first.\n\n")
	} else {
		canDistResp, err := rp.CanDistribute()
		if err != nil {
			fmt.Printf("  %sCould not check fee distributor balance: %s%s\n\n", colorYellow, err, colorReset)
		} else {
			balance := eth.WeiToEth(canDistResp.Balance)
			if balance == 0 {
				fmt.Printf("  No balance in fee distributor.\n\n")
			} else {
				rEthShare := balance - canDistResp.NodeShare
				fmt.Printf("  Distributor balance: %.6f ETH\n", math.RoundDown(balance, 6))
				fmt.Printf("  Your share:          %.6f ETH\n", math.RoundDown(canDistResp.NodeShare, 6))
				fmt.Printf("  rETH stakers share:  %.6f ETH\n\n", math.RoundDown(rEthShare, 6))

				nodeShareWei := eth.EthToWei(canDistResp.NodeShare)
				totalEthWei.Add(totalEthWei, nodeShareWei)

				gasInfo := canDistResp.GasInfo
				claims = append(claims, pendingClaim{
					id:      feeDistID,
					name:    "Fee Distributor (distribute)",
					gasInfo: gasInfo,
					execute: func() error {
						fmt.Printf("  Submitting transaction...\n")
						response, err := rp.Distribute()
						if err != nil {
							return fmt.Errorf("transaction could not be submitted: %w", err)
						}
						fmt.Printf("  Distributing fee distributor balance...\n")
						cliutils.PrintTransactionHash(rp, response.TxHash)
						if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
							return fmt.Errorf("transaction was submitted but failed on-chain: %w", err)
						}
						fmt.Printf("  %sSuccessfully distributed fee distributor balance.%s\n", colorGreen, colorReset)
						return nil
					},
				})
			}
		}
	}

	// ================================================================
	// 3. Minipool Balance Distribution
	// ================================================================
	sectionID++
	minipoolID := sectionID
	fmt.Printf("%s--- [%d] Minipool Balance Distribution ---%s\n", colorGreen, minipoolID, colorReset)

	minipoolDetails, err := rp.GetDistributeBalanceDetails()
	if err != nil {
		fmt.Printf("  %sCould not check minipool balances: %s%s\n\n", colorYellow, err, colorReset)
	} else {
		eligibleMinipools := []api.MinipoolBalanceDistributionDetails{}
		for _, mp := range minipoolDetails.Details {
			if mp.CanDistribute {
				eligibleMinipools = append(eligibleMinipools, mp)
			}
		}

		if len(eligibleMinipools) == 0 {
			fmt.Printf("  No minipools eligible for balance distribution.\n\n")
		} else {
			// Sort by balance (highest first)
			sort.Slice(eligibleMinipools, func(i, j int) bool {
				first := eligibleMinipools[i]
				second := eligibleMinipools[j]
				var firstAmt, secondAmt float64
				if first.Status == types.Dissolved {
					firstAmt = eth.WeiToEth(first.Balance)
				} else {
					firstAmt = eth.WeiToEth(first.NodeShareOfBalance) + eth.WeiToEth(first.Refund)
				}
				if second.Status == types.Dissolved {
					secondAmt = eth.WeiToEth(second.Balance)
				} else {
					secondAmt = eth.WeiToEth(second.NodeShareOfBalance) + eth.WeiToEth(second.Refund)
				}
				return firstAmt > secondAmt
			})

			mpTotalEth := new(big.Int)
			for _, mp := range eligibleMinipools {
				if mp.Status == types.Dissolved {
					fmt.Printf("  %s: %.6f ETH (dissolved, all to you)\n", mp.Address.Hex(), math.RoundDown(eth.WeiToEth(mp.Balance), 6))
					mpTotalEth.Add(mpTotalEth, mp.Balance)
				} else {
					nodeAmount := new(big.Int).Add(mp.NodeShareOfBalance, mp.Refund)
					fmt.Printf("  %s: %.6f ETH (your share) + %.6f ETH (refund)\n",
						mp.Address.Hex(),
						math.RoundDown(eth.WeiToEth(mp.NodeShareOfBalance), 6),
						math.RoundDown(eth.WeiToEth(mp.Refund), 6))
					mpTotalEth.Add(mpTotalEth, nodeAmount)
				}
			}
			fmt.Printf("  Total from %d minipool(s): %.6f ETH\n\n", len(eligibleMinipools), math.RoundDown(eth.WeiToEth(mpTotalEth), 6))
			totalEthWei.Add(totalEthWei, mpTotalEth)

			// Accumulate gas
			var totalGasEst, totalGasSafe uint64
			var mpGasInfo rocketpoolapi.GasInfo
			for _, mp := range eligibleMinipools {
				mpGasInfo = mp.GasInfo
				totalGasEst += mp.GasInfo.EstGasLimit
				totalGasSafe += mp.GasInfo.SafeGasLimit
			}
			mpGasInfo.EstGasLimit = totalGasEst
			mpGasInfo.SafeGasLimit = totalGasSafe

			// Capture for closure
			mps := eligibleMinipools
			claims = append(claims, pendingClaim{
				id:      minipoolID,
				name:    fmt.Sprintf("Minipool Balance Distribution (%d minipool(s))", len(mps)),
				gasInfo: mpGasInfo,
				execute: func() error {
					failCount := 0
					for _, mp := range mps {
						fmt.Printf("  Submitting transaction for minipool %s...\n", mp.Address.Hex())
						response, err := rp.DistributeBalance(mp.Address)
						if err != nil {
							fmt.Printf("  %sFailed to distribute minipool %s: %s%s\n", colorRed, mp.Address.Hex(), err, colorReset)
							failCount++
							continue
						}
						fmt.Printf("  Distributing balance of minipool %s...\n", mp.Address.Hex())
						cliutils.PrintTransactionHash(rp, response.TxHash)
						if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
							fmt.Printf("  %sTransaction failed for minipool %s: %s%s\n", colorRed, mp.Address.Hex(), err, colorReset)
							failCount++
						} else {
							fmt.Printf("  %sSuccessfully distributed balance of minipool %s.%s\n", colorGreen, mp.Address.Hex(), colorReset)
						}
					}
					if failCount > 0 {
						return fmt.Errorf("%d of %d minipool distribution(s) failed", failCount, len(mps))
					}
					return nil
				},
			})
		}
	}

	// ================================================================
	// 4. Periodic Rewards (RPL + ETH)
	// ================================================================
	sectionID++
	periodicID := sectionID
	fmt.Printf("%s--- [%d] Periodic Rewards (RPL + ETH) ---%s\n", colorGreen, periodicID, colorReset)

	rewardsInfo, err := rp.GetRewardsInfo()
	if err != nil {
		fmt.Printf("  %sCould not check periodic rewards: %s%s\n\n", colorYellow, err, colorReset)
	} else if !rewardsInfo.Registered {
		fmt.Printf("  Node is not registered.\n\n")
	} else {
		// Handle missing/invalid merkle trees
		missingIntervals := []int{}
		for _, interval := range rewardsInfo.InvalidIntervals {
			if !interval.TreeFileExists || !interval.MerkleRootValid {
				missingIntervals = append(missingIntervals, int(interval.Index))
			}
		}
		if len(missingIntervals) > 0 && !statusOnly {
			fmt.Printf("  %sMissing or invalid Merkle tree files for intervals: %v%s\n", colorYellow, missingIntervals, colorReset)
			if autoConfirm || prompt.Confirm("  Would you like to download the missing rewards tree files?") {
				cfg, _, err := rp.LoadConfig()
				if err != nil {
					fmt.Printf("  %sCould not load config for tree download: %s%s\n", colorYellow, err, colorReset)
				} else {
					for _, interval := range rewardsInfo.InvalidIntervals {
						if !interval.TreeFileExists || !interval.MerkleRootValid {
							fmt.Printf("  Downloading interval %d file... ", interval.Index)
							err := interval.DownloadRewardsFile(cfg, false)
							if err != nil {
								fmt.Printf("error: %s\n", err)
							} else {
								fmt.Println("done!")
							}
						}
					}
					// Reload rewards info
					rewardsInfo, err = rp.GetRewardsInfo()
					if err != nil {
						fmt.Printf("  %sCould not reload rewards info: %s%s\n\n", colorYellow, err, colorReset)
					}
				}
			}
		}

		if err == nil && len(rewardsInfo.UnclaimedIntervals) == 0 {
			fmt.Printf("  No unclaimed reward intervals.\n\n")
		} else if err == nil {
			prTotalRpl := new(big.Int)
			prTotalEth := new(big.Int)
			var intervalIndices []uint64
			for _, interval := range rewardsInfo.UnclaimedIntervals {
				intervalIndices = append(intervalIndices, interval.Index)
				prTotalRpl.Add(prTotalRpl, &interval.CollateralRplAmount.Int)
				prTotalRpl.Add(prTotalRpl, &interval.ODaoRplAmount.Int)
				prTotalEth.Add(prTotalEth, &interval.SmoothingPoolEthAmount.Int)
				prTotalEth.Add(prTotalEth, &interval.VoterShareEth.Int)
			}

			fmt.Printf("  Unclaimed intervals: %d\n", len(rewardsInfo.UnclaimedIntervals))
			for _, interval := range rewardsInfo.UnclaimedIntervals {
				rpl := new(big.Int).Add(&interval.CollateralRplAmount.Int, &interval.ODaoRplAmount.Int)
				ethAmt := new(big.Int).Add(&interval.SmoothingPoolEthAmount.Int, &interval.VoterShareEth.Int)
				fmt.Printf("    Interval %d: %.6f RPL, %.6f ETH\n", interval.Index,
					math.RoundDown(eth.WeiToEth(rpl), 6),
					math.RoundDown(eth.WeiToEth(ethAmt), 6))
			}
			fmt.Printf("  Total: %.6f RPL + %.6f ETH\n\n",
				math.RoundDown(eth.WeiToEth(prTotalRpl), 6),
				math.RoundDown(eth.WeiToEth(prTotalEth), 6))

			totalRplWei.Add(totalRplWei, prTotalRpl)
			totalEthWei.Add(totalEthWei, prTotalEth)

			// Parse restake flag (interactive prompt deferred until after claim selection)
			periodicClaimRpl = prTotalRpl
			periodicIntervalIndices = intervalIndices
			restakeAmountFlag := c.String("restake-amount")
			if restakeAmountFlag == "all" {
				periodicRestakeAmount = prTotalRpl
				periodicRestakeResolved = true
			} else if restakeAmountFlag != "" {
				stakeAmt, parseErr := strconv.ParseFloat(restakeAmountFlag, 64)
				if parseErr == nil && stakeAmt > 0 {
					periodicRestakeAmount = eth.EthToWei(stakeAmt)
					if periodicRestakeAmount.Cmp(prTotalRpl) > 0 {
						periodicRestakeAmount = prTotalRpl
					}
				}
				periodicRestakeResolved = true
			} else if autoConfirm {
				// Ignore restaking if -y is specified but restake-amount isn't
				periodicRestakeAmount = nil
				periodicRestakeResolved = true
			}

			// Get preliminary gas estimate (restake prompt deferred, so use claim-only estimate)
			var gasInfo rocketpoolapi.GasInfo
			canClaim, canErr := rp.CanNodeClaimRewards(intervalIndices)
			if canErr != nil {
				fmt.Printf("  %sWarning: could not estimate gas for periodic rewards: %s%s\n", colorYellow, canErr, colorReset)
			} else {
				gasInfo = canClaim.GasInfo
			}

			claims = append(claims, pendingClaim{
				id:      periodicID,
				name:    "Periodic Rewards (RPL + ETH)",
				gasInfo: gasInfo,
				execute: func() error {
					fmt.Printf("  Submitting transaction...\n")
					var txHash common.Hash
					if periodicRestakeAmount == nil {
						response, err := rp.NodeClaimRewards(periodicIntervalIndices)
						if err != nil {
							return fmt.Errorf("transaction could not be submitted: %w", err)
						}
						txHash = response.TxHash
					} else {
						response, err := rp.NodeClaimAndStakeRewards(periodicIntervalIndices, periodicRestakeAmount)
						if err != nil {
							return fmt.Errorf("transaction could not be submitted: %w", err)
						}
						txHash = response.TxHash
					}
					fmt.Printf("  Claiming periodic rewards...\n")
					cliutils.PrintTransactionHash(rp, txHash)
					if _, err := rp.WaitForTransaction(txHash); err != nil {
						return fmt.Errorf("transaction was submitted but failed on-chain: %w", err)
					}
					if periodicRestakeAmount != nil {
						fmt.Printf("  %sSuccessfully claimed rewards and restaked %.6f RPL.%s\n", colorGreen, eth.WeiToEth(periodicRestakeAmount), colorReset)
					} else {
						fmt.Printf("  %sSuccessfully claimed periodic rewards.%s\n", colorGreen, colorReset)
					}
					return nil
				},
			})
		}
	}

	// ================================================================
	// 5. Megapool Refund - ETH refunded to the owner after a dissolution or from distributed rewards
	// ================================================================
	if isSaturn {
		sectionID++
		megaRefundID := sectionID
		fmt.Printf("%s--- [%d] Megapool Refund ---%s\n", colorGreen, megaRefundID, colorReset)

		megapoolStatus, err := rp.MegapoolStatus(false)
		if err != nil {
			fmt.Printf("  %sCould not check megapool status: %s%s\n\n", colorYellow, err, colorReset)
		} else if megapoolStatus.Megapool.RefundValue == nil || megapoolStatus.Megapool.RefundValue.Cmp(big.NewInt(0)) <= 0 {
			fmt.Printf("  No megapool refund available.\n\n")
		} else {
			refundVal := megapoolStatus.Megapool.RefundValue
			fmt.Printf("  Refund value: %.6f ETH\n", math.RoundDown(eth.WeiToEth(refundVal), 6))
			if megapoolStatus.Megapool.NodeDebt != nil && megapoolStatus.Megapool.NodeDebt.Cmp(big.NewInt(0)) > 0 {
				fmt.Printf("  Node debt (deducted): %.6f ETH\n", math.RoundDown(eth.WeiToEth(megapoolStatus.Megapool.NodeDebt), 6))
			}
			fmt.Println()
			totalEthWei.Add(totalEthWei, refundVal)

			canClaim, canErr := rp.CanClaimMegapoolRefund()
			var gasInfo rocketpoolapi.GasInfo
			canClaimOk := false
			if canErr != nil {
				fmt.Printf("  %sWarning: could not estimate gas: %s%s\n", colorYellow, canErr, colorReset)
			} else if !canClaim.CanClaim {
				fmt.Printf("  %sCannot claim megapool refund at this time.%s\n", colorYellow, colorReset)
			} else {
				gasInfo = canClaim.GasInfo
				canClaimOk = true
			}

			if canClaimOk {
				claims = append(claims, pendingClaim{
					id:      megaRefundID,
					name:    "Megapool Refund (claim)",
					gasInfo: gasInfo,
					execute: func() error {
						fmt.Printf("  Submitting transaction...\n")
						response, err := rp.ClaimMegapoolRefund()
						if err != nil {
							return fmt.Errorf("transaction could not be submitted: %w", err)
						}
						fmt.Printf("  Claiming megapool refund...\n")
						cliutils.PrintTransactionHash(rp, response.TxHash)
						if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
							return fmt.Errorf("transaction was submitted but failed on-chain: %w", err)
						}
						fmt.Printf("  %sSuccessfully claimed megapool refund.%s\n", colorGreen, colorReset)
						return nil
					},
				})
			}
		}
	}

	// ================================================================
	// 6. Unclaimed Rewards - available when the withdrawal address was unable to receive ETH
	// ================================================================
	if isSaturn {
		sectionID++
		unclaimedID := sectionID
		fmt.Printf("%s--- [%d] Unclaimed Rewards ---%s\n", colorGreen, unclaimedID, colorReset)

		nodeStatus, err := rp.NodeStatus()
		if err != nil {
			fmt.Printf("  %sCould not check node status: %s%s\n\n", colorYellow, err, colorReset)
		} else if nodeStatus.UnclaimedRewards == nil || nodeStatus.UnclaimedRewards.Cmp(big.NewInt(0)) <= 0 {
			fmt.Printf("  No unclaimed rewards.\n\n")
		} else {
			fmt.Printf("  Unclaimed rewards: %.6f ETH\n", math.RoundDown(eth.WeiToEth(nodeStatus.UnclaimedRewards), 6))
			fmt.Printf("  (Rewards distributed previously but not yet sent to withdrawal address)\n\n")
			totalEthWei.Add(totalEthWei, nodeStatus.UnclaimedRewards)

			nodeAddr := nodeStatus.AccountAddress
			canClaim, canErr := rp.CanClaimUnclaimedRewards(nodeAddr)
			var gasInfo rocketpoolapi.GasInfo
			canClaimOk := false
			if canErr != nil {
				fmt.Printf("  %sWarning: could not estimate gas: %s%s\n", colorYellow, canErr, colorReset)
			} else if !canClaim.CanClaim {
				fmt.Printf("  %sCannot claim unclaimed rewards at this time.%s\n", colorYellow, colorReset)
			} else {
				gasInfo = canClaim.GasInfo
				canClaimOk = true
			}

			if canClaimOk {
				claims = append(claims, pendingClaim{
					id:      unclaimedID,
					name:    "Unclaimed Rewards (claim)",
					gasInfo: gasInfo,
					execute: func() error {
						fmt.Printf("  Submitting transaction...\n")
						response, err := rp.ClaimUnclaimedRewards(nodeAddr)
						if err != nil {
							return fmt.Errorf("transaction could not be submitted: %w", err)
						}
						fmt.Printf("  Claiming unclaimed rewards...\n")
						cliutils.PrintTransactionHash(rp, response.TxHash)
						if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
							return fmt.Errorf("transaction was submitted but failed on-chain: %w", err)
						}
						fmt.Printf("  %sSuccessfully claimed unclaimed rewards.%s\n", colorGreen, colorReset)
						return nil
					},
				})
			}
		}
	}

	// ================================================================
	// 7. PDAO Bond Claims (RPL)
	// ================================================================
	sectionID++
	pdaoID := sectionID
	fmt.Printf("%s--- [%d] PDAO Bond Claims ---%s\n", colorGreen, pdaoID, colorReset)

	bondsResponse, err := rp.PDAOGetClaimableBonds()
	if err != nil {
		fmt.Printf("  %sCould not check PDAO bonds: %s%s\n\n", colorYellow, err, colorReset)
	} else if len(bondsResponse.ClaimableBonds) == 0 {
		fmt.Printf("  No claimable bonds or rewards.\n\n")
	} else {
		pdaoRplTotal := new(big.Int)
		for _, bond := range bondsResponse.ClaimableBonds {
			bondTotal := new(big.Int).Add(bond.UnlockAmount, bond.RewardAmount)
			pdaoRplTotal.Add(pdaoRplTotal, bondTotal)
			fmt.Printf("  Proposal %d: %.6f RPL (unlock) + %.6f RPL (reward)\n",
				bond.ProposalID,
				math.RoundDown(eth.WeiToEth(bond.UnlockAmount), 6),
				math.RoundDown(eth.WeiToEth(bond.RewardAmount), 6))
		}
		fmt.Printf("  Total: %.6f RPL from %d proposal(s)\n\n",
			math.RoundDown(eth.WeiToEth(pdaoRplTotal), 6), len(bondsResponse.ClaimableBonds))
		totalRplWei.Add(totalRplWei, pdaoRplTotal)

		// Accumulate gas
		var totalGasEst, totalGasSafe uint64
		var bondGasInfo rocketpoolapi.GasInfo
		allCanClaim := true
		for _, bond := range bondsResponse.ClaimableBonds {
			indices := getClaimIndicesForBond(bond)
			canResponse, canErr := rp.PDAOCanClaimBonds(bond.ProposalID, indices)
			if canErr != nil {
				fmt.Printf("  %sWarning: could not estimate gas for proposal %d: %s%s\n", colorYellow, bond.ProposalID, canErr, colorReset)
				allCanClaim = false
				break
			}
			bondGasInfo = canResponse.GasInfo
			totalGasEst += canResponse.GasInfo.EstGasLimit
			totalGasSafe += canResponse.GasInfo.SafeGasLimit
		}

		if allCanClaim {
			bondGasInfo.EstGasLimit = totalGasEst
			bondGasInfo.SafeGasLimit = totalGasSafe
			bonds := bondsResponse.ClaimableBonds
			claims = append(claims, pendingClaim{
				id:      pdaoID,
				name:    fmt.Sprintf("PDAO Bond Claims (%d proposal(s))", len(bonds)),
				gasInfo: bondGasInfo,
				execute: func() error {
					failCount := 0
					for _, bond := range bonds {
						indices := getClaimIndicesForBond(bond)
						fmt.Printf("  Submitting transaction for proposal %d...\n", bond.ProposalID)
						response, err := rp.PDAOClaimBonds(bond.IsProposer, bond.ProposalID, indices)
						if err != nil {
							fmt.Printf("  %sFailed to claim bonds from proposal %d: %s%s\n", colorRed, bond.ProposalID, err, colorReset)
							failCount++
							continue
						}
						fmt.Printf("  Claiming bonds from proposal %d...\n", bond.ProposalID)
						cliutils.PrintTransactionHash(rp, response.TxHash)
						if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
							fmt.Printf("  %sTransaction failed for proposal %d: %s%s\n", colorRed, bond.ProposalID, err, colorReset)
							failCount++
						} else {
							fmt.Printf("  %sSuccessfully claimed bonds from proposal %d.%s\n", colorGreen, bond.ProposalID, colorReset)
						}
					}
					if failCount > 0 {
						return fmt.Errorf("%d of %d PDAO bond claim(s) failed", failCount, len(bonds))
					}
					return nil
				},
			})
		}
	}

	// ================================================================
	// Summary
	// ================================================================
	fmt.Printf("%s============================================================%s\n", colorGreen, colorReset)
	fmt.Printf("%s                       Totals                               %s\n", colorGreen, colorReset)
	fmt.Printf("%s============================================================%s\n", colorGreen, colorReset)
	fmt.Printf("  ETH: %.6f\n", math.RoundDown(eth.WeiToEth(totalEthWei), 6))
	fmt.Printf("  RPL: %.6f\n\n", math.RoundDown(eth.WeiToEth(totalRplWei), 6))

	if statusOnly {
		if len(claims) > 0 {
			fmt.Printf("Run 'rocketpool rewards claim-all' to claim these rewards.\n")
		}
		return nil
	}

	if len(claims) == 0 {
		fmt.Println("No rewards available to claim at this time.")
		return nil
	}

	// List what can be claimed
	fmt.Printf("The following %d claim(s) are available:\n", len(claims))
	for i, claim := range claims {
		fmt.Printf("  %d. %s\n", i+1, claim.name)
	}
	fmt.Println()

	// Select which claims to execute
	var selectedClaims []pendingClaim
	if autoConfirm {
		selectedClaims = claims
	} else {
		indexSelection := prompt.Prompt(
			"Enter the numbers of the claims you want to execute (comma-separated), 'all' to claim everything, or 'none' to cancel:",
			"^(all|none|\\d+(,\\d+)*)$",
			"Invalid selection. Enter 'all', 'none', or comma-separated numbers (e.g. '1,3').",
		)

		if indexSelection == "none" {
			fmt.Println("Cancelled.")
			return nil
		} else if indexSelection == "all" {
			selectedClaims = claims
		} else {
			elements := strings.Split(indexSelection, ",")
			seen := map[int]bool{}
			for _, element := range elements {
				idx, err := strconv.Atoi(strings.TrimSpace(element))
				if err != nil || idx < 1 || idx > len(claims) {
					return fmt.Errorf("invalid selection '%s': must be between 1 and %d", element, len(claims))
				}
				if !seen[idx] {
					selectedClaims = append(selectedClaims, claims[idx-1])
					seen[idx] = true
				}
			}
		}
	}

	if len(selectedClaims) == 0 {
		fmt.Println("No claims selected.")
		return nil
	}

	fmt.Printf("\n%d claim(s) selected:\n", len(selectedClaims))
	for i, claim := range selectedClaims {
		fmt.Printf("  %d. %s\n", i+1, claim.name)
	}
	fmt.Println()

	// If the periodic rewards claim is selected and restake hasn't been resolved yet, prompt now
	if !periodicRestakeResolved && periodicClaimRpl != nil {
		for i := range selectedClaims {
			if selectedClaims[i].id == periodicID {
				availableRpl := eth.WeiToEth(periodicClaimRpl)
				amountOptions := []string{
					"None (do not restake any RPL)",
					fmt.Sprintf("All %.6f RPL", availableRpl),
					"A custom amount",
				}
				selected, _ := prompt.Select("Please choose an amount of RPL to restake:", amountOptions)
				switch selected {
				case 0:
					periodicRestakeAmount = nil
				case 1:
					periodicRestakeAmount = periodicClaimRpl
				case 2:
					for {
						inputAmount := prompt.Prompt("Please enter an amount of RPL to restake:", "^\\d+(\\.\\d+)?$", "Invalid amount")
						stakeAmount, err := strconv.ParseFloat(inputAmount, 64)
						if err != nil {
							fmt.Printf("Invalid amount '%s': %s\n", inputAmount, err.Error())
						} else if stakeAmount < 0 {
							fmt.Println("Amount must be greater than zero.")
						} else if stakeAmount > availableRpl {
							fmt.Println("Amount must be less than or equal to the RPL available to claim.")
						} else {
							periodicRestakeAmount = eth.EthToWei(stakeAmount)
							break
						}
					}
				}
				// Re-estimate gas if restaking was chosen
				if periodicRestakeAmount != nil {
					canClaim, canErr := rp.CanNodeClaimAndStakeRewards(periodicIntervalIndices, periodicRestakeAmount)
					if canErr == nil {
						selectedClaims[i].gasInfo = canClaim.GasInfo
					}
				}
				break
			}
		}
	}

	// Accumulate total gas for fee estimation
	var totalGasEst, totalGasSafe uint64
	var lastGasInfo rocketpoolapi.GasInfo
	for _, claim := range selectedClaims {
		lastGasInfo = claim.gasInfo
		totalGasEst += claim.gasInfo.EstGasLimit
		totalGasSafe += claim.gasInfo.SafeGasLimit
	}
	lastGasInfo.EstGasLimit = totalGasEst
	lastGasInfo.SafeGasLimit = totalGasSafe

	// Get gas fee settings (single prompt for all transactions)
	g, err := gas.GetMaxFeeAndLimit(lastGasInfo, rp, autoConfirm)
	if err != nil {
		return err
	}

	// Execute selected claims
	fmt.Printf("\n%sExecuting %d claim(s)...%s\n", colorBlue, len(selectedClaims), colorReset)
	successCount := 0
	failCount := 0
	skippedCount := 0
	for i, claim := range selectedClaims {
		fmt.Printf("\n%s[%d/%d] %s%s\n", colorBlue, i+1, len(selectedClaims), claim.name, colorReset)
		g.Assign(rp)
		err := claim.execute()
		if err != nil {
			failCount++
			fmt.Printf("\n  %sERROR: %s%s\n", colorRed, err, colorReset)

			// If there are more claims and we're not auto-confirming, ask whether to continue
			remaining := len(selectedClaims) - i - 1
			if remaining > 0 {
				if autoConfirm {
					fmt.Printf("  %sContinuing with remaining %d claim(s)...%s\n", colorYellow, remaining, colorReset)
				} else {
					if !prompt.Confirm(fmt.Sprintf("  The above claim failed. Continue with the remaining %d claim(s)?", remaining)) {
						skippedCount = remaining
						fmt.Println("  Aborting remaining claims.")
						break
					}
				}
			}
		} else {
			successCount++
		}
	}

	// Final summary
	fmt.Println()
	fmt.Printf("============================================================\n")
	if failCount == 0 && skippedCount == 0 {
		fmt.Printf("%sAll %d claim(s) completed successfully.%s\n", colorGreen, successCount, colorReset)
	} else if successCount == 0 {
		fmt.Printf("%sAll %d claim(s) failed.%s\n", colorRed, failCount, colorReset)
		if skippedCount > 0 {
			fmt.Printf("%s%d claim(s) were skipped.%s\n", colorYellow, skippedCount, colorReset)
		}
	} else {
		fmt.Printf("%s%d claim(s) succeeded%s, %s%d failed%s",
			colorGreen, successCount, colorReset,
			colorRed, failCount, colorReset)
		if skippedCount > 0 {
			fmt.Printf(", %s%d skipped%s", colorYellow, skippedCount, colorReset)
		}
		fmt.Println(".")
	}
	fmt.Printf("============================================================\n")

	if failCount > 0 {
		return fmt.Errorf("%d of %d claim(s) failed", failCount, failCount+successCount)
	}
	return nil
}

// getClaimIndicesForBond collects all unique indices from a bond's unlockable and rewardable indices.
func getClaimIndicesForBond(bond api.BondClaimResult) []uint64 {
	indexMap := map[uint64]bool{}
	for _, index := range bond.UnlockableIndices {
		indexMap[index] = true
	}
	for _, index := range bond.RewardableIndices {
		indexMap[index] = true
	}

	indices := make([]uint64, 0, len(indexMap))
	for index := range indexMap {
		indices = append(indices, index)
	}

	sort.SliceStable(indices, func(i, j int) bool {
		return indices[i] < indices[j]
	})

	return indices
}
