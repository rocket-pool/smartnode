package rewards

// This file contains treegen tests which use mock history.
// These mocks are faster to process than real history, and are useful for
// testing new features and refactoring.

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/rewards/test"
	"github.com/rocket-pool/smartnode/shared/services/rewards/test/assets"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

func TestMockIntervalDefaultsTreegenv11(tt *testing.T) {

	history := test.NewDefaultMockHistory(true)
	// Add a node which is earning some bonus commission
	node := history.GetNewDefaultMockNode(&test.NewMockNodeParams{
		SmoothingPool:     true,
		EightEthMinipools: 1,
		CollateralRpl:     5,
	})
	node.Minipools[0].NodeFee, _ = big.NewInt(0).SetString("50000000000000000", 10)
	history.Nodes = append(history.Nodes, node)
	state := history.GetEndNetworkState()
	state.IsSaturnDeployed = true

	t := newRewardsTest(tt, state.NetworkDetails.RewardIndex)

	t.bc.SetState(state)

	consensusStartBlock := history.GetConsensusStartBlock()
	executionStartBlock := history.GetExecutionStartBlock()
	consensusEndBlock := history.GetConsensusEndBlock()
	executionEndBlock := history.GetExecutionEndBlock()

	logger := log.NewColorLogger(color.Faint)

	t.rp.SetRewardSnapshotEvent(history.GetPreviousRewardSnapshotEvent())
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock-1), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock - 1})
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock})
	t.rp.SetHeaderByNumber(big.NewInt(int64(executionStartBlock)), &types.Header{Time: uint64(history.GetStartTime().Unix())})

	for _, validator := range state.MinipoolValidatorDetails {
		t.bc.SetMinipoolPerformance(validator.Index, make([]uint64, 0))
	}
	for _, validator := range state.MegapoolValidatorGlobalIndex {
		pubkey := rptypes.BytesToValidatorPubkey(validator.Pubkey)
		details := state.MegapoolValidatorDetails[pubkey]
		t.bc.SetMinipoolPerformance(details.Index, make([]uint64, 0))
	}

	// Set some custom balances for the validators that opt in and out of smoothing pool
	nodeSummary := history.GetNodeSummary()
	customBalanceNodes := nodeSummary.MustGetClass(tt, "single_eight_eth_opted_in_quarter")
	for _, node := range customBalanceNodes {
		node.Minipools[0].SPWithdrawals = eth.EthToWei(0.75)
	}
	customBalanceNodes = nodeSummary.MustGetClass(tt, "single_eight_eth_opted_out_three_quarters")
	for _, node := range customBalanceNodes {
		node.Minipools[0].SPWithdrawals = eth.EthToWei(0.75)
	}
	customBalanceNodes = nodeSummary.MustGetClass(tt, "single_bond_reduction")
	for _, node := range customBalanceNodes {
		node.Minipools[0].SPWithdrawals = eth.EthToWei(0.5)
	}

	history.SetWithdrawals(t.bc)

	generatorv11 := newTreeGeneratorImpl_v11(
		&logger,
		t.Name()+"-stateless",
		state.NetworkDetails.RewardIndex,
		&SnapshotEnd{
			Slot:           consensusEndBlock,
			ConsensusBlock: consensusEndBlock,
			ExecutionBlock: executionEndBlock,
		},
		&types.Header{
			Number: big.NewInt(int64(history.GetExecutionEndBlock())),
			Time:   assets.Mainnet20ELHeaderTime,
		},
		/* intervalsPassed= */ 1,
		state,
		true,
	)

	v11Artifacts, err := generatorv11.generateTree(
		t.rp,
		"mainnet",
		make([]common.Address, 0),
		t.bc,
	)
	t.failIf(err)

	if testing.Verbose() {
		t.saveArtifacts("v11", v11Artifacts)
	}

	// Validate individual node details in the rewards file
	rewardsFile := v11Artifacts.RewardsFile
	minipoolPerformanceFile := v11Artifacts.MinipoolPerformanceFile

	singleEightEthNodes := nodeSummary.MustGetClass(tt, "single_eight_eth")
	singleSixteenEthNodes := nodeSummary.MustGetClass(tt, "single_sixteen_eth")
	for _, node := range append(singleEightEthNodes, singleSixteenEthNodes...) {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)

		expectedRewardsAmount, _ := big.NewInt(0).SetString("911258527391109533960", 10)

		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got 0 ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Sign() != 0 {
			t.Logf("Node %+v", node)
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), "0")
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}
	singleEightEthNodesSP := nodeSummary.MustGetClass(tt, "single_eight_eth_sp")
	singleSixteenEthNodesSP := nodeSummary.MustGetClass(tt, "single_sixteen_eth_sp")
	for _, node := range append(singleEightEthNodesSP, singleSixteenEthNodesSP...) {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)

		expectedRewardsAmount, _ := big.NewInt(0).SetString("911258527391109533960", 10)

		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		expectedEthAmount := big.NewInt(0)
		if node.SmoothingPoolRegistrationState {
			if node.Class == "single_eight_eth_sp" {
				expectedEthAmount.SetString("631795904665995300", 10)
				// There should be a bonus for these nodes' minipools
				if len(node.Minipools) != 1 {
					t.Fatalf("Expected 1 minipool for node %s, got %d", node.Notes, len(node.Minipools))
				}
				minipoolPerf, _ := minipoolPerformanceFile.GetMinipoolPerformance(node.Minipools[0].Address)
				// 8 eth minipools with 10% collateral earn 14% commission overall.
				// They earned 10% on 24/32 of the 1 eth of consensus rewards already, which is 0.075 eth.
				// Their bonus is therefore 4/10 of 0.075 eth, which is 0.03 eth.
				expectedBonusEthEarned, _ := big.NewInt(0).SetString("30000000000000000", 10)
				if minipoolPerf.GetBonusEthEarned().Cmp(expectedBonusEthEarned) != 0 {
					t.Fatalf("Minipool %s bonus does not match expected value: %s != %s", node.Minipools[0].Address.Hex(), minipoolPerf.GetBonusEthEarned().String(), expectedBonusEthEarned.String())
				}
				expectedAttestationScore := big.NewInt(0).Sub(oneEth, big.NewInt(14e16))
				expectedAttestationScore.Mul(expectedAttestationScore, eightEth)
				expectedAttestationScore.Div(expectedAttestationScore, thirtyTwoEth)
				expectedAttestationScore.Add(expectedAttestationScore, big.NewInt(14e16))
				expectedAttestationScore.Mul(expectedAttestationScore, big.NewInt(101)) // there are 101 epochs in the interval
				if minipoolPerf.GetAttestationScore().Cmp(expectedAttestationScore) != 0 {
					t.Fatalf("Minipool %s attestation score does not match expected value: %s != %s", node.Minipools[0].Address.Hex(), minipoolPerf.GetAttestationScore().String(), expectedAttestationScore.String())
				}
			} else {
				// 16-eth minipools earn more eth! A bit less than double.
				expectedEthAmount.SetString("932359852299429338", 10)
				if len(node.Minipools) != 1 {
					t.Fatalf("Expected 1 minipool for node %s, got %d", node.Notes, len(node.Minipools))
				}
				minipoolPerf, _ := minipoolPerformanceFile.GetMinipoolPerformance(node.Minipools[0].Address)
				// The 16 eth minipools earn 10% on 24/32.
				expectedAttestationScore := big.NewInt(0).Sub(oneEth, big.NewInt(1e17))
				expectedAttestationScore.Mul(expectedAttestationScore, sixteenEth)
				expectedAttestationScore.Div(expectedAttestationScore, thirtyTwoEth)
				expectedAttestationScore.Add(expectedAttestationScore, big.NewInt(1e17))
				expectedAttestationScore.Mul(expectedAttestationScore, big.NewInt(101)) // there are 101 epochs in the interval
				if minipoolPerf.GetAttestationScore().Cmp(expectedAttestationScore) != 0 {
					t.Fatalf("Minipool %s attestation score does not match expected value: %s != %s", node.Minipools[0].Address.Hex(), minipoolPerf.GetAttestationScore().String(), expectedAttestationScore.String())
				}
				// 16 eth minipools earn no bonus.
				if minipoolPerf.GetBonusEthEarned().Sign() != 0 {
					t.Fatalf("Minipool %s bonus does not match expected value: %s != 0", node.Minipools[0].Address.Hex(), minipoolPerf.GetBonusEthEarned().String())
				}
			}
		}
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}

	optingInNodesSP := append(
		nodeSummary.MustGetClass(tt, "single_eight_eth_opted_in_quarter"),
		nodeSummary.MustGetClass(tt, "single_sixteen_eth_opted_in_quarter")...,
	)
	for _, node := range optingInNodesSP {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)

		mp := node.Minipools[0]
		perf, _ := minipoolPerformanceFile.GetMinipoolPerformance(mp.Address)

		// Node has 20 RPL and only 1 8-eth minpool which puts it above the linear curve
		expectedRewardsAmount := big.NewInt(0)
		if node.Class == "single_eight_eth_opted_in_quarter" {
			expectedRewardsAmount.SetString("1595205464807146635862", 10)
		} else {
			// 16-eth minipools earn less for the same RPL stake, due to RPIP-30
			expectedRewardsAmount.SetString("1171278656914890979185", 10)
			if perf.GetBonusEthEarned().Sign() != 0 {
				// 16 eth minipools should not get bonus commission
				t.Fatalf("Minipool %s shouldn't have earned bonus eth and did", mp.Address.Hex())
			}
		}

		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		expectedEthAmount := big.NewInt(0)
		if node.Class == "single_eight_eth_opted_in_quarter" {
			// About 3/4 what the full nodes got
			expectedEthAmount.SetString("475336522322927156", 10)
			// Earns 3/4 the bonus of a node that was in for the whole interval
			expectedBonusEthEarned, _ := big.NewInt(0).SetString("22500000000000000", 10)
			if perf.GetBonusEthEarned().Cmp(expectedBonusEthEarned) != 0 {
				t.Fatalf("Minipool %s bonus does not match expected value: %s != %s", mp.Address.Hex(), perf.GetBonusEthEarned().String(), expectedBonusEthEarned.String())
			}
		} else {
			// 16-eth minipools earn more eth! A bit less than double.
			expectedEthAmount.SetString("701577710641154749", 10)
		}
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}

	optingOutNodesSP := append(
		nodeSummary.MustGetClass(tt, "single_eight_eth_opted_out_three_quarters"),
		nodeSummary.MustGetClass(tt, "single_sixteen_eth_opted_out_three_quarters")...,
	)
	for _, node := range optingOutNodesSP {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)

		mp := node.Minipools[0]
		perf, _ := minipoolPerformanceFile.GetMinipoolPerformance(mp.Address)

		// Node has 20 RPL and only 1 8-eth minpool which puts it above the linear curve
		expectedRewardsAmount := big.NewInt(0)
		if node.Class == "single_eight_eth_opted_out_three_quarters" {
			expectedRewardsAmount.SetString("1595205464807146635862", 10)
		} else {
			// 16-eth minipools earn less for the same RPL stake, due to RPIP-30
			expectedRewardsAmount.SetString("1171278656914890979185", 10)
		}

		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		expectedEthAmount := big.NewInt(0)
		if node.Class == "single_eight_eth_opted_out_three_quarters" {
			// About 3/4 what the full nodes got
			expectedEthAmount.SetString("469378147029204431", 10)
			// Earns 3/4 the bonus of a node that was in for the whole interval
			expectedBonusEthEarned, _ := big.NewInt(0).SetString("22500000000000000", 10)
			if perf.GetBonusEthEarned().Cmp(expectedBonusEthEarned) != 0 {
				t.Fatalf("Minipool %s bonus does not match expected value: %s != %s", mp.Address.Hex(), perf.GetBonusEthEarned().String(), expectedBonusEthEarned.String())
			}
		} else {
			// 16-eth minipools earn more eth! A bit less than double.
			expectedEthAmount.SetString("692346424974823766", 10)
			if perf.GetBonusEthEarned().Sign() != 0 {
				// 16 eth minipools should not get bonus commission
				t.Fatalf("Minipool %s shouldn't have earned bonus eth and did", mp.Address.Hex())
			}
		}
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}

	bondReductionNode := nodeSummary.MustGetClass(tt, "single_bond_reduction")
	for _, node := range bondReductionNode {

		mp := node.Minipools[0]
		perf, _ := minipoolPerformanceFile.GetMinipoolPerformance(mp.Address)

		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)

		// Nodes that bond reduce are treated as having their new bond for the full interval,
		// when it comes to RPL rewards.
		expectedRewardsAmount, _ := big.NewInt(0).SetString("911258527391109533960", 10)

		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got reduced ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		expectedEthAmount, _ := big.NewInt(0).SetString("822401812688821752", 10)
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// And a reduced bonus
		expectedBonusEthEarned, _ := big.NewInt(0).SetString("15000000000000000", 10)
		if perf.GetBonusEthEarned().Cmp(expectedBonusEthEarned) != 0 {
			t.Fatalf("Minipool %s bonus does not match expected value: %s != %s", mp.Address.Hex(), perf.GetBonusEthEarned().String(), expectedBonusEthEarned.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}

	noMinipoolsNodes := nodeSummary.MustGetClass(tt, "no_minipools")
	for _, node := range noMinipoolsNodes {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Sign() != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), "0")
		}

		// Make sure it got ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Sign() != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), "0")
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}

	megapoolNoRplNoSpNodes := nodeSummary.MustGetClass(tt, "megapool_no_rpl_no_sp")
	for _, node := range megapoolNoRplNoSpNodes {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Sign() != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), "0")
		}

		// Make sure it didn't get any ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Sign() != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), "0")
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}

	megapoolNoRplNoSpMinipoolsNodes := nodeSummary.MustGetClass(tt, "megapool_no_rpl_no_sp_minipools")
	for _, node := range megapoolNoRplNoSpMinipoolsNodes {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Sign() != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), "0")
		}

		// Make sure it didn't get any ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Sign() != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), "0")
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}

	megapoolNoRplNoSpMinipoolsCollateralNodes := nodeSummary.MustGetClass(tt, "megapool_no_rpl_no_sp_minipools_collateral")
	for _, node := range megapoolNoRplNoSpMinipoolsCollateralNodes {
		// Check the rewards amount in the rewards file
		expectedRewardsAmount, _ := big.NewInt(0).SetString("911258527391109533960", 10)
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Logf("Node %+v", node)
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it didn't get any ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Sign() != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), "0")
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}

	megapoolNoRplSpNodes := nodeSummary.MustGetClass(tt, "megapool_no_rpl_sp")
	for _, node := range megapoolNoRplSpNodes {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Sign() != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), "0")
		}

		// Make sure it got ETH
		expectedEthAmount, _ := big.NewInt(0).SetString("858194864048338368", 10)
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}

	megapoolNoRplSpMinipoolsNodes := nodeSummary.MustGetClass(tt, "megapool_no_rpl_sp_minipools")
	for _, node := range megapoolNoRplSpMinipoolsNodes {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Sign() != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), "0")
		}

		// Make sure it got ETH
		expectedEthAmount, _ := big.NewInt(0).SetString("3824794394091977170", 10)
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}

	megapoolNoRplSpMinipoolsCollateralNodes := nodeSummary.MustGetClass(tt, "megapool_no_rpl_sp_minipools_collateral")
	for _, node := range megapoolNoRplSpMinipoolsCollateralNodes {
		// Check the rewards amount in the rewards file
		expectedRewardsAmount, _ := big.NewInt(0).SetString("911258527391109533960", 10)
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got ETH
		expectedEthAmount, _ := big.NewInt(0).SetString("3873307989258140312", 10)
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got 0 voter share ETH
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Sign() != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), "0")
		}
	}

	megapoolStakedRplNoSpNodes := nodeSummary.MustGetClass(tt, "megapool_staked_rpl_no_sp")
	for _, node := range megapoolStakedRplNoSpNodes {
		validatorCount := node.MegapoolValidators
		// Check the rewards amount in the rewards file
		expectedRewardsAmount, _ := big.NewInt(0).SetString("91125852739110953396", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		expectedRewardsAmount.Mul(expectedRewardsAmount, big.NewInt(int64(validatorCount)))
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got no ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Sign() != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), "0")
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got voter share ETH
		exepectedVoterShareEthAmount, _ := big.NewInt(0).SetString("944580815709969788", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		exepectedVoterShareEthAmount.Mul(exepectedVoterShareEthAmount, big.NewInt(int64(validatorCount)))
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		diff := new(big.Int).Sub(voterShareEthAmount, exepectedVoterShareEthAmount)
		diff.Abs(diff)
		if diff.Cmp(big.NewInt(1)) > 0 { // allow off by one wei
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), exepectedVoterShareEthAmount.String())
		}
	}

	megapoolStakedRplNoSpMinipoolsNodes := nodeSummary.MustGetClass(tt, "megapool_staked_rpl_no_sp_minipools")
	for _, node := range megapoolStakedRplNoSpMinipoolsNodes {
		validatorCount := node.MegapoolValidators
		// Check the rewards amount in the rewards file
		expectedRewardsAmount, _ := big.NewInt(0).SetString("91125852739110953396", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		expectedRewardsAmount.Mul(expectedRewardsAmount, big.NewInt(int64(validatorCount)))
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got no ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Sign() != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), "0")
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got voter share ETH
		exepectedVoterShareEthAmount, _ := big.NewInt(0).SetString("944580815709969788", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		exepectedVoterShareEthAmount.Mul(exepectedVoterShareEthAmount, big.NewInt(int64(validatorCount)))
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Cmp(exepectedVoterShareEthAmount) != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), exepectedVoterShareEthAmount.String())
		}
	}

	megapoolStakedRplNoSpMinipoolsCollateralNodes := nodeSummary.MustGetClass(tt, "megapool_staked_rpl_no_sp_minipools_collateral")
	for _, node := range megapoolStakedRplNoSpMinipoolsCollateralNodes {
		validatorCount := node.MegapoolValidators
		// Check the rewards amount in the rewards file
		expectedRewardsAmount, _ := big.NewInt(0).SetString("91125852739110953396", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		expectedRewardsAmount.Mul(expectedRewardsAmount, big.NewInt(int64(validatorCount)))
		// Add a constant amount for minipool rewards
		minipoolRewardsAmount, _ := big.NewInt(0).SetString("911258527391109533960", 10)
		expectedRewardsAmount.Add(expectedRewardsAmount, minipoolRewardsAmount)
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got no ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Sign() != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), "0")
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got voter share ETH
		exepectedVoterShareEthAmount, _ := big.NewInt(0).SetString("944580815709969788", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		exepectedVoterShareEthAmount.Mul(exepectedVoterShareEthAmount, big.NewInt(int64(validatorCount)))
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		diff := new(big.Int).Sub(voterShareEthAmount, exepectedVoterShareEthAmount)
		diff.Abs(diff)
		if diff.Cmp(big.NewInt(1)) > 0 { // allow off by one wei
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), exepectedVoterShareEthAmount.String())
		}
	}

	megapoolStakedRplSpNodes := nodeSummary.MustGetClass(tt, "megapool_staked_rpl_sp")
	for _, node := range megapoolStakedRplSpNodes {
		validatorCount := node.MegapoolValidators
		// Check the rewards amount in the rewards file
		expectedRewardsAmount, _ := big.NewInt(0).SetString("91125852739110953396", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		expectedRewardsAmount.Mul(expectedRewardsAmount, big.NewInt(int64(validatorCount)))
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got ETH
		expectedEthAmount, _ := big.NewInt(0).SetString("286064954682779456", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		expectedEthAmount.Mul(expectedEthAmount, big.NewInt(int64(validatorCount)))
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got voter share ETH
		exepectedVoterShareEthAmount, _ := big.NewInt(0).SetString("944580815709969788", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		exepectedVoterShareEthAmount.Mul(exepectedVoterShareEthAmount, big.NewInt(int64(validatorCount)))
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		diff := new(big.Int).Sub(voterShareEthAmount, exepectedVoterShareEthAmount)
		diff.Abs(diff)
		if diff.Cmp(big.NewInt(1)) > 0 { // allow off by one wei
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), exepectedVoterShareEthAmount.String())
		}
	}

	megapoolStakedRplSpMinipoolsNodes := nodeSummary.MustGetClass(tt, "megapool_staked_rpl_sp_minipools")
	for _, node := range megapoolStakedRplSpMinipoolsNodes {
		validatorCount := node.MegapoolValidators
		// Check the rewards amount in the rewards file
		expectedRewardsAmount, _ := big.NewInt(0).SetString("91125852739110953396", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		expectedRewardsAmount.Mul(expectedRewardsAmount, big.NewInt(int64(validatorCount)))
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got ETH
		expectedEthAmount, _ := big.NewInt(0).SetString("3252664484726418258", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		expectedEthAmount.Mul(expectedEthAmount, big.NewInt(int64(validatorCount)))
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got voter share ETH
		exepectedVoterShareEthAmount, _ := big.NewInt(0).SetString("944580815709969788", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		exepectedVoterShareEthAmount.Mul(exepectedVoterShareEthAmount, big.NewInt(int64(validatorCount)))
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		if voterShareEthAmount.Cmp(exepectedVoterShareEthAmount) != 0 {
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), exepectedVoterShareEthAmount.String())
		}
	}

	megapoolStakedRplSpMinipoolsCollateralNodes := nodeSummary.MustGetClass(tt, "megapool_staked_rpl_sp_minipools_collateral")
	for _, node := range megapoolStakedRplSpMinipoolsCollateralNodes {
		validatorCount := node.MegapoolValidators
		// Check the rewards amount in the rewards file
		expectedRewardsAmount, _ := big.NewInt(0).SetString("91125852739110953396", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		expectedRewardsAmount.Mul(expectedRewardsAmount, big.NewInt(int64(validatorCount)))
		// Add a constant amount for minipool rewards
		minipoolRewardsAmount, _ := big.NewInt(0).SetString("911258527391109533960", 10)
		expectedRewardsAmount.Add(expectedRewardsAmount, minipoolRewardsAmount)
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)
		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got ETH
		minipoolEthAmount, _ := big.NewInt(0).SetString("2698353474320241690", 10)
		expectedEthAmount, _ := big.NewInt(0).SetString("444444780127559583", 10)
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		// Multiply by i+1 since the number of validators scales with i+1
		expectedEthAmount.Mul(expectedEthAmount, big.NewInt(int64(validatorCount)))
		expectedEthAmount.Add(expectedEthAmount, minipoolEthAmount)
		diff := new(big.Int).Sub(ethAmount, expectedEthAmount)
		diff.Abs(diff)
		if diff.Cmp(big.NewInt(1)) > 0 { // allow off by one wei
			fmt.Printf("Node: %+v\n", node)
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}

		// Make sure it got voter share ETH
		exepectedVoterShareEthAmount, _ := big.NewInt(0).SetString("944580815709969789", 10)
		// Multiply by i+1 since the number of validators scales with i+1
		exepectedVoterShareEthAmount.Mul(exepectedVoterShareEthAmount, big.NewInt(int64(validatorCount)))
		voterShareEthAmount := rewardsFile.GetNodeVoterShareEth(node.Address)
		diff = new(big.Int).Sub(voterShareEthAmount, exepectedVoterShareEthAmount)
		diff.Abs(diff)
		if diff.Cmp(big.NewInt(1)) > 0 { // allow off by one wei
			t.Fatalf("Voter share ETH amount does not match expected value for node %s: %s != %s", node.Notes, voterShareEthAmount.String(), exepectedVoterShareEthAmount.String())
		}
	}
	// Validate merkle root
	v11MerkleRoot := v11Artifacts.RewardsFile.GetMerkleRoot()

	// Expected merkle root:
	// 0x45c2300d804f26dd032623948da6696c27ba4c02d3a52c4b44b64c87d68a3ac7
	//
	// If this does not match, it implies either you updated the set of default mock nodes,
	// or you introduced a regression in treegen.
	// DO NOT update this value unless you know what you are doing.
	expectedMerkleRoot := "0x45c2300d804f26dd032623948da6696c27ba4c02d3a52c4b44b64c87d68a3ac7"
	if !strings.EqualFold(v11MerkleRoot, expectedMerkleRoot) {
		t.Fatalf("Merkle root does not match expected value %s != %s", v11MerkleRoot, expectedMerkleRoot)
	} else {
		t.Logf("Merkle root matches expected value %s", expectedMerkleRoot)
	}
}

func TestInsufficientEthForBonusesesV11(tt *testing.T) {

	history := test.NewDefaultMockHistoryNoNodes()
	// Add two nodes which are earning some bonus commission
	nodeOne := history.GetNewDefaultMockNode(&test.NewMockNodeParams{
		SmoothingPool:     true,
		EightEthMinipools: 1,
		CollateralRpl:     5,
	})
	nodeOne.Minipools[0].NodeFee, _ = big.NewInt(0).SetString("50000000000000000", 10)
	history.Nodes = append(history.Nodes, nodeOne)
	nodeTwo := history.GetNewDefaultMockNode(&test.NewMockNodeParams{
		SmoothingPool:     true,
		EightEthMinipools: 1,
		CollateralRpl:     20,
	})
	history.Nodes = append(history.Nodes, nodeTwo)

	// Add oDAO nodes
	odaoNodes := history.GetDefaultMockODAONodes()
	history.Nodes = append(history.Nodes, odaoNodes...)

	// Ovewrite the SP balance to a value under the bonus commission
	history.NetworkDetails.SmoothingPoolBalance = big.NewInt(1100)
	// Set the SP voter share to 0
	history.NetworkDetails.SmoothingPoolPendingVoterShare = big.NewInt(100)
	// Set the pdao share to 0
	state := history.GetEndNetworkState()
	state.IsSaturnDeployed = true

	t := newRewardsTest(tt, state.NetworkDetails.RewardIndex)

	t.bc.SetState(state)
	history.SetWithdrawals(t.bc)

	consensusStartBlock := history.GetConsensusStartBlock()
	executionStartBlock := history.GetExecutionStartBlock()
	consensusEndBlock := history.GetConsensusEndBlock()
	executionEndBlock := history.GetExecutionEndBlock()

	logger := log.NewColorLogger(color.Faint)

	t.rp.SetRewardSnapshotEvent(history.GetPreviousRewardSnapshotEvent())
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock-1), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock - 1})
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock})
	t.rp.SetHeaderByNumber(big.NewInt(int64(executionStartBlock)), &types.Header{Time: uint64(history.GetStartTime().Unix())})

	for _, validator := range state.MinipoolValidatorDetails {
		t.bc.SetMinipoolPerformance(validator.Index, make([]uint64, 0))
	}

	generatorv11 := newTreeGeneratorImpl_v11(
		&logger,
		t.Name(),
		state.NetworkDetails.RewardIndex,
		&SnapshotEnd{
			Slot:           consensusEndBlock,
			ConsensusBlock: consensusEndBlock,
			ExecutionBlock: executionEndBlock,
		},
		&types.Header{
			Number: big.NewInt(int64(history.GetExecutionEndBlock())),
			Time:   assets.Mainnet20ELHeaderTime,
		},
		/* intervalsPassed= */ 1,
		state,
		true,
	)

	v11Artifacts, err := generatorv11.generateTree(
		t.rp,
		"mainnet",
		make([]common.Address, 0),
		t.bc,
	)
	t.failIf(err)

	if testing.Verbose() {
		t.saveArtifacts("v11", v11Artifacts)
	}

	// Check the rewards file
	rewardsFile := v11Artifacts.RewardsFile
	ethOne := rewardsFile.GetNodeSmoothingPoolEth(nodeOne.Address)
	if ethOne.Uint64() != 579 {
		t.Fatalf("Node one ETH amount does not match expected value: %s != %d", ethOne.String(), 169+416)
	}
	ethTwo := rewardsFile.GetNodeSmoothingPoolEth(nodeTwo.Address)
	if ethTwo.Uint64() != 420 {
		t.Fatalf("Node two ETH amount does not match expected value: %s != %d", ethTwo.String(), 177+237)
	}

	// Check the minipool performance file
	minipoolPerformanceFile := v11Artifacts.MinipoolPerformanceFile
	perfOne, ok := minipoolPerformanceFile.GetMinipoolPerformance(nodeOne.Minipools[0].Address)
	if !ok {
		t.Fatalf("Node one minipool performance not found")
	}
	if perfOne.GetBonusEthEarned().Uint64() != 393 {
		t.Fatalf("Node one bonus does not match expected value: %s != %d", perfOne.GetBonusEthEarned().String(), 416)
	}
	perfTwo, ok := minipoolPerformanceFile.GetMinipoolPerformance(nodeTwo.Minipools[0].Address)
	if !ok {
		t.Fatalf("Node two minipool performance not found")
	}
	if perfTwo.GetBonusEthEarned().Uint64() != 225 {
		t.Fatalf("Node two bonus does not match expected value: %s != %d", perfTwo.GetBonusEthEarned().String(), 237)
	}
}

func TestMockNoRPLRewardsV11(tt *testing.T) {

	history := test.NewDefaultMockHistoryNoNodes()
	// Add two nodes which are earning some bonus commission
	nodeOne := history.GetNewDefaultMockNode(&test.NewMockNodeParams{
		SmoothingPool:     false,
		EightEthMinipools: 1,
		CollateralRpl:     0,
	})
	nodeOne.Minipools[0].NodeFee, _ = big.NewInt(0).SetString("50000000000000000", 10)
	history.Nodes = append(history.Nodes, nodeOne)
	nodeTwo := history.GetNewDefaultMockNode(&test.NewMockNodeParams{
		SmoothingPool:     true,
		EightEthMinipools: 2,
		CollateralRpl:     0,
	})
	nodeTwo.Minipools[0].NodeFee, _ = big.NewInt(0).SetString("50000000000000000", 10)
	nodeTwo.Minipools[1].NodeFee, _ = big.NewInt(0).SetString("50000000000000000", 10)
	history.Nodes = append(history.Nodes, nodeTwo)

	// Add oDAO nodes
	odaoNodes := history.GetDefaultMockODAONodes()
	history.Nodes = append(history.Nodes, odaoNodes...)

	state := history.GetEndNetworkState()
	state.IsSaturnDeployed = true

	t := newRewardsTest(tt, state.NetworkDetails.RewardIndex)

	t.bc.SetState(state)

	// Give all three minipools 1 ETH of consensus income
	t.bc.AddWithdrawal(history.BeaconConfig.FirstSlotOfEpoch(history.StartEpoch+1), nodeOne.Minipools[0].ValidatorIndex, big.NewInt(1e18))
	t.bc.AddWithdrawal(history.BeaconConfig.FirstSlotOfEpoch(history.StartEpoch+1), nodeTwo.Minipools[0].ValidatorIndex, big.NewInt(1e18))
	t.bc.AddWithdrawal(history.BeaconConfig.FirstSlotOfEpoch(history.StartEpoch+1), nodeTwo.Minipools[1].ValidatorIndex, big.NewInt(1e18))

	consensusStartBlock := history.GetConsensusStartBlock()
	executionStartBlock := history.GetExecutionStartBlock()
	consensusEndBlock := history.GetConsensusEndBlock()
	executionEndBlock := history.GetExecutionEndBlock()

	logger := log.NewColorLogger(color.Faint)

	t.rp.SetRewardSnapshotEvent(history.GetPreviousRewardSnapshotEvent())
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock-1), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock - 1})
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock})
	t.rp.SetHeaderByNumber(big.NewInt(int64(executionStartBlock)), &types.Header{Time: uint64(history.GetStartTime().Unix())})

	for _, validator := range state.MinipoolValidatorDetails {
		t.bc.SetMinipoolPerformance(validator.Index, make([]uint64, 0))
	}

	generatorv11 := newTreeGeneratorImpl_v11(
		&logger,
		t.Name(),
		state.NetworkDetails.RewardIndex,
		&SnapshotEnd{
			Slot:           consensusEndBlock,
			ConsensusBlock: consensusEndBlock,
			ExecutionBlock: executionEndBlock,
		},
		&types.Header{
			Number: big.NewInt(int64(history.GetExecutionEndBlock())),
			Time:   assets.Mainnet20ELHeaderTime,
		},
		/* intervalsPassed= */ 1,
		state,
		true,
	)

	v11Artifacts, err := generatorv11.generateTree(
		t.rp,
		"mainnet",
		make([]common.Address, 0),
		t.bc,
	)
	t.failIf(err)

	if testing.Verbose() {
		t.saveArtifacts("v11", v11Artifacts)
	}

	// Check the rewards file
	rewardsFile := v11Artifacts.RewardsFile
	ethOne := rewardsFile.GetNodeSmoothingPoolEth(nodeOne.Address)
	// Node one is not a SP, so it should have 0 ETH
	if ethOne.Uint64() != 0 {
		t.Fatalf("Node one ETH amount does not match expected value: %s != %d", ethOne.String(), 0)
	}
	ethTwo := rewardsFile.GetNodeSmoothingPoolEth(nodeTwo.Address)
	expectedEthTwo, _ := big.NewInt(0).SetString("32575000000000000000", 10)
	if ethTwo.Cmp(expectedEthTwo) != 0 {
		t.Fatalf("Node two ETH amount does not match expected value: %s != %s", ethTwo.String(), expectedEthTwo.String())
	}

	// Check the minipool performance file
	minipoolPerformanceFile := v11Artifacts.MinipoolPerformanceFile
	_, ok := minipoolPerformanceFile.GetMinipoolPerformance(nodeOne.Minipools[0].Address)
	if ok {
		t.Fatalf("Node one minipool performance should not be found")
	}
	perfTwo, ok := minipoolPerformanceFile.GetMinipoolPerformance(nodeTwo.Minipools[0].Address)
	if !ok {
		t.Fatalf("Node two minipool one performance not found")
	}
	if perfTwo.GetBonusEthEarned().Uint64() != 37500000000000000 {
		t.Fatalf("Node two minipool one bonus does not match expected value: %s != %d", perfTwo.GetBonusEthEarned().String(), 37500000000000000)
	}
	// Node two is in the SP and starts with 5% commission. It has no RPL staked, so it earns an extra 5% on top of that.
	if perfTwo.GetEffectiveCommission().Uint64() != 100000000000000000 {
		t.Fatalf("Node two minipool one effective commission does not match expected value: %s != %d", perfTwo.GetEffectiveCommission().String(), 100000000000000000)
	}
	perfThree, ok := minipoolPerformanceFile.GetMinipoolPerformance(nodeTwo.Minipools[1].Address)
	if !ok {
		t.Fatalf("Node two minipool two performance not found")
	}
	if perfThree.GetBonusEthEarned().Uint64() != 37500000000000000 {
		t.Fatalf("Node two minipool two bonus does not match expected value: %s != %d", perfThree.GetBonusEthEarned().String(), 37500000000000000)
	}
	// Node two is in the SP and starts with 5% commission. It has no RPL staked, so it earns an extra 5% on top of that.
	if perfThree.GetEffectiveCommission().Uint64() != 100000000000000000 {
		t.Fatalf("Node two minipool two effective commission does not match expected value: %s != %d", perfThree.GetEffectiveCommission().String(), 100000000000000000)
	}
}

func TestMockOptedOutAndThenBondReducedV11(tt *testing.T) {

	history := test.NewDefaultMockHistoryNoNodes()
	// Add two nodes which are earning some bonus commission
	nodeOne := history.GetNewDefaultMockNode(&test.NewMockNodeParams{
		SmoothingPool:     false,
		EightEthMinipools: 1,
		CollateralRpl:     0,
	})
	// Opted out 1/4 of the way through the interval
	nodeOne.SmoothingPoolRegistrationChanged = history.BeaconConfig.GetSlotTime(history.BeaconConfig.FirstSlotOfEpoch(history.StartEpoch + (history.EndEpoch-history.StartEpoch)/4))
	nodeOne.Minipools[0].NodeFee, _ = big.NewInt(0).SetString("50000000000000000", 10)
	// Bond reduced 1/2 of the way through the interval
	nodeOne.Minipools[0].LastBondReductionTime = history.BeaconConfig.GetSlotTime(history.BeaconConfig.FirstSlotOfEpoch(history.StartEpoch + (history.EndEpoch-history.StartEpoch)/2))
	nodeOne.Minipools[0].LastBondReductionPrevValue = big.NewInt(0).Set(sixteenEth)
	nodeOne.Minipools[0].LastBondReductionPrevNodeFee, _ = big.NewInt(0).SetString("140000000000000000", 10)
	history.Nodes = append(history.Nodes, nodeOne)
	nodeTwo := history.GetNewDefaultMockNode(&test.NewMockNodeParams{
		SmoothingPool:     true,
		EightEthMinipools: 1,
		CollateralRpl:     0,
	})
	nodeTwo.Minipools[0].NodeFee, _ = big.NewInt(0).SetString("50000000000000000", 10)
	history.Nodes = append(history.Nodes, nodeTwo)

	// Add oDAO nodes
	odaoNodes := history.GetDefaultMockODAONodes()
	history.Nodes = append(history.Nodes, odaoNodes...)

	state := history.GetEndNetworkState()
	state.IsSaturnDeployed = true

	t := newRewardsTest(tt, state.NetworkDetails.RewardIndex)

	t.bc.SetState(state)

	// Add withdrawals to both minipools
	t.bc.AddWithdrawal(history.BeaconConfig.FirstSlotOfEpoch(history.StartEpoch+1), nodeOne.Minipools[0].ValidatorIndex, big.NewInt(1e18))
	t.bc.AddWithdrawal(history.BeaconConfig.FirstSlotOfEpoch(history.StartEpoch+1), nodeTwo.Minipools[0].ValidatorIndex, big.NewInt(1e18))

	consensusStartBlock := history.GetConsensusStartBlock()
	executionStartBlock := history.GetExecutionStartBlock()
	consensusEndBlock := history.GetConsensusEndBlock()
	executionEndBlock := history.GetExecutionEndBlock()

	logger := log.NewColorLogger(color.Faint)

	t.rp.SetRewardSnapshotEvent(history.GetPreviousRewardSnapshotEvent())
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock-1), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock - 1})
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock})
	t.rp.SetHeaderByNumber(big.NewInt(int64(executionStartBlock)), &types.Header{Time: uint64(history.GetStartTime().Unix())})

	for _, validator := range state.MinipoolValidatorDetails {
		t.bc.SetMinipoolPerformance(validator.Index, make([]uint64, 0))
	}

	generatorv11 := newTreeGeneratorImpl_v11(
		&logger,
		t.Name(),
		state.NetworkDetails.RewardIndex,
		&SnapshotEnd{
			Slot:           consensusEndBlock,
			ConsensusBlock: consensusEndBlock,
			ExecutionBlock: executionEndBlock,
		},
		&types.Header{
			Number: big.NewInt(int64(history.GetExecutionEndBlock())),
			Time:   assets.Mainnet20ELHeaderTime,
		},
		/* intervalsPassed= */ 1,
		state,
		true,
	)

	v11Artifacts, err := generatorv11.generateTree(
		t.rp,
		"mainnet",
		make([]common.Address, 0),
		t.bc,
	)
	t.failIf(err)

	if testing.Verbose() {
		t.saveArtifacts("v11", v11Artifacts)
	}

	// Check the rewards file
	rewardsFile := v11Artifacts.RewardsFile
	ethOne := rewardsFile.GetNodeSmoothingPoolEth(nodeOne.Address)
	// Node one was in the SP so it should have some ETH, but no bonuses
	expectedEthOne, _ := big.NewInt(0).SetString("11309523809523809523", 10)
	if ethOne.Cmp(expectedEthOne) != 0 {
		t.Fatalf("Node one ETH amount does not match expected value: %s != %s", ethOne.String(), expectedEthOne.String())
	}
	ethTwo := rewardsFile.GetNodeSmoothingPoolEth(nodeTwo.Address)
	expectedEthTwo, _ := big.NewInt(0).SetString("26089087301587301587", 10)
	if ethTwo.Cmp(expectedEthTwo) != 0 {
		t.Fatalf("Node two ETH amount does not match expected value: %s != %s", ethTwo.String(), expectedEthTwo.String())
	}

	// Check the minipool performance file
	minipoolPerformanceFile := v11Artifacts.MinipoolPerformanceFile
	perfOne, ok := minipoolPerformanceFile.GetMinipoolPerformance(nodeOne.Minipools[0].Address)
	if !ok {
		t.Fatalf("Node one minipool performance should be found")
	}
	if perfOne.GetBonusEthEarned().Uint64() != 0 {
		t.Fatalf("Node one minipool one bonus does not match expected value: %s != %d", perfOne.GetBonusEthEarned().String(), 0)
	}
	perfTwo, ok := minipoolPerformanceFile.GetMinipoolPerformance(nodeTwo.Minipools[0].Address)
	if !ok {
		t.Fatalf("Node two minipool one performance not found")
	}
	if perfTwo.GetBonusEthEarned().Uint64() != 37500000000000000 {
		t.Fatalf("Node two minipool one bonus does not match expected value: %s != %d", perfTwo.GetBonusEthEarned().String(), 37500000000000000)
	}
	// Node two is in the SP and starts with 5% commission. It has no RPL staked, so it earns an extra 5% on top of that.
	if perfTwo.GetEffectiveCommission().Uint64() != 100000000000000000 {
		t.Fatalf("Node two minipool one effective commission does not match expected value: %s != %d", perfTwo.GetEffectiveCommission().String(), 100000000000000000)
	}
}

func TestMockWithdrawableEpochV11(tt *testing.T) {

	history := test.NewDefaultMockHistoryNoNodes()
	// Add two nodes which are earning some bonus commission
	nodeOne := history.GetNewDefaultMockNode(&test.NewMockNodeParams{
		SmoothingPool:     true,
		EightEthMinipools: 1,
		CollateralRpl:     0,
	})
	nodeOne.Minipools[0].NodeFee, _ = big.NewInt(0).SetString("50000000000000000", 10)
	history.Nodes = append(history.Nodes, nodeOne)
	nodeTwo := history.GetNewDefaultMockNode(&test.NewMockNodeParams{
		SmoothingPool:     true,
		EightEthMinipools: 1,
		CollateralRpl:     0,
	})
	nodeTwo.Minipools[0].NodeFee, _ = big.NewInt(0).SetString("50000000000000000", 10)
	// Withdrawable epoch half way through the interval
	nodeTwo.Minipools[0].WithdrawableEpoch = history.StartEpoch + (history.EndEpoch-history.StartEpoch)/2
	history.Nodes = append(history.Nodes, nodeTwo)

	// Add oDAO nodes
	odaoNodes := history.GetDefaultMockODAONodes()
	history.Nodes = append(history.Nodes, odaoNodes...)

	state := history.GetEndNetworkState()
	state.IsSaturnDeployed = true

	t := newRewardsTest(tt, state.NetworkDetails.RewardIndex)

	t.bc.SetState(state)

	// Add withdrawals to both minipools
	t.bc.AddWithdrawal(history.BeaconConfig.FirstSlotOfEpoch(history.StartEpoch+1), nodeOne.Minipools[0].ValidatorIndex, big.NewInt(1e18))
	// Add a withdrawal in the epoch after the interval ends
	t.bc.AddWithdrawal(history.BeaconConfig.FirstSlotOfEpoch(history.EndEpoch+1), nodeOne.Minipools[0].ValidatorIndex, big.NewInt(1e18))
	// Withdraw 0.5 eth at the start of the interval
	t.bc.AddWithdrawal(history.BeaconConfig.FirstSlotOfEpoch(history.StartEpoch+1), nodeTwo.Minipools[0].ValidatorIndex, big.NewInt(5e17))
	// Withdraw 32.5 eth at the end of the interval
	t.bc.AddWithdrawal(history.BeaconConfig.FirstSlotOfEpoch(history.EndEpoch-1), nodeTwo.Minipools[0].ValidatorIndex, big.NewInt(0).Mul(big.NewInt(325), big.NewInt(1e17)))

	consensusStartBlock := history.GetConsensusStartBlock()
	executionStartBlock := history.GetExecutionStartBlock()
	consensusEndBlock := history.GetConsensusEndBlock()
	executionEndBlock := history.GetExecutionEndBlock()

	logger := log.NewColorLogger(color.Faint)

	t.rp.SetRewardSnapshotEvent(history.GetPreviousRewardSnapshotEvent())
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock-1), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock - 1})
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock})
	t.rp.SetHeaderByNumber(big.NewInt(int64(executionStartBlock)), &types.Header{Time: uint64(history.GetStartTime().Unix())})

	for _, validator := range state.MinipoolValidatorDetails {
		t.bc.SetMinipoolPerformance(validator.Index, make([]uint64, 0))
	}

	generatorv11 := newTreeGeneratorImpl_v11(
		&logger,
		t.Name(),
		state.NetworkDetails.RewardIndex,
		&SnapshotEnd{
			Slot:           consensusEndBlock,
			ConsensusBlock: consensusEndBlock,
			ExecutionBlock: executionEndBlock,
		},
		&types.Header{
			Number: big.NewInt(int64(history.GetExecutionEndBlock())),
			Time:   assets.Mainnet20ELHeaderTime,
		},
		/* intervalsPassed= */ 1,
		state,
		true,
	)

	v11Artifacts, err := generatorv11.generateTree(
		t.rp,
		"mainnet",
		make([]common.Address, 0),
		t.bc,
	)
	t.failIf(err)

	if testing.Verbose() {
		t.saveArtifacts("v11", v11Artifacts)
	}

	// Check the rewards file
	rewardsFile := v11Artifacts.RewardsFile
	ethOne := rewardsFile.GetNodeSmoothingPoolEth(nodeOne.Address)
	expectedEthOne, _ := big.NewInt(0).SetString("21920833333333333333", 10)
	if ethOne.Cmp(expectedEthOne) != 0 {
		t.Fatalf("Node one ETH amount does not match expected value: %s != %s", ethOne.String(), expectedEthOne.String())
	}
	ethTwo := rewardsFile.GetNodeSmoothingPoolEth(nodeTwo.Address)
	expectedEthTwo, _ := big.NewInt(0).SetString("10654166666666666666", 10)
	if ethTwo.Cmp(expectedEthTwo) != 0 {
		t.Fatalf("Node two ETH amount does not match expected value: %s != %s", ethTwo.String(), expectedEthTwo.String())
	}

	// Check the minipool performance file
	minipoolPerformanceFile := v11Artifacts.MinipoolPerformanceFile
	perfOne, ok := minipoolPerformanceFile.GetMinipoolPerformance(nodeOne.Minipools[0].Address)
	if !ok {
		t.Fatalf("Node one minipool performance should be found")
	}
	if perfOne.GetBonusEthEarned().Uint64() != 37500000000000000 {
		t.Fatalf("Node one minipool one bonus does not match expected value: %s != %d", perfOne.GetBonusEthEarned().String(), 37500000000000000)
	}
	if perfOne.GetEffectiveCommission().Uint64() != 100000000000000000 {
		t.Fatalf("Node one minipool one effective commission does not match expected value: %s != %d", perfOne.GetEffectiveCommission().String(), 1000000000000000000)
	}
	if perfOne.GetConsensusIncome().Uint64() != 1000000000000000000 {
		t.Fatalf("Node one minipool one consensus income does not match expected value: %s != %d", perfOne.GetConsensusIncome().String(), 1000000000000000000)
	}
	perfTwo, ok := minipoolPerformanceFile.GetMinipoolPerformance(nodeTwo.Minipools[0].Address)
	if !ok {
		t.Fatalf("Node two minipool one performance not found")
	}
	if perfTwo.GetBonusEthEarned().Uint64() != 37500000000000000 {
		t.Fatalf("Node two minipool one bonus does not match expected value: %s != %d", perfTwo.GetBonusEthEarned().String(), 37500000000000000)
	}
	if perfTwo.GetEffectiveCommission().Uint64() != 100000000000000000 {
		t.Fatalf("Node two minipool one effective commission does not match expected value: %s != %d", perfTwo.GetEffectiveCommission().String(), 100000000000000000)
	}
	if perfTwo.GetConsensusIncome().Uint64() != 1000000000000000000 {
		t.Fatalf("Node two minipool one consensus income does not match expected value: %s != %d", perfTwo.GetConsensusIncome().String(), 1000000000000000000)
	}
}
