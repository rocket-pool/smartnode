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
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/rewards/test"
	"github.com/rocket-pool/smartnode/shared/services/rewards/test/assets"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

func TestMockIntervalDefaultsTreegenv8v9(tt *testing.T) {
	t := newV8Test(tt)

	history := test.NewDefaultMockHistory()
	state := history.GetEndNetworkState()

	t.bc.SetState(state)

	consensusStartBlock := history.GetConsensusStartBlock()
	executionStartBlock := history.GetExecutionStartBlock()
	consensusEndBlock := history.GetConsensusEndBlock()
	executionEndBlock := history.GetExecutionEndBlock()

	logger := log.NewColorLogger(color.Faint)
	generator := newTreeGeneratorImpl_v8(
		&logger,
		t.Name(),
		state.NetworkDetails.RewardIndex,
		history.GetStartTime(),
		history.GetEndTime(),
		consensusEndBlock,
		&types.Header{
			Number: big.NewInt(int64(history.GetExecutionEndBlock())),
			Time:   assets.Mainnet20ELHeaderTime,
		},
		/* intervalsPassed= */ 1,
		state,
	)

	t.rp.SetRewardSnapshotEvent(history.GetPreviousRewardSnapshotEvent())
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock-1), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock - 1})
	t.bc.SetBeaconBlock(fmt.Sprint(consensusStartBlock), beacon.BeaconBlock{ExecutionBlockNumber: executionStartBlock})
	t.rp.SetHeaderByNumber(big.NewInt(int64(executionStartBlock)), &types.Header{Time: uint64(history.GetStartTime().Unix())})

	for _, validator := range state.ValidatorDetails {
		t.bc.SetMinipoolPerformance(validator.Index, make([]uint64, 0))
	}

	v8Artifacts, err := generator.generateTree(
		t.rp,
		"mainnet",
		make([]common.Address, 0),
		t.bc,
	)
	t.failIf(err)

	if testing.Verbose() {
		t.saveArtifacts("v8", v8Artifacts)
	}
	generatorv9 := newTreeGeneratorImpl_v9(
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
	)

	v9Artifacts, err := generatorv9.generateTree(
		t.rp,
		"mainnet",
		make([]common.Address, 0),
		t.bc,
	)
	t.failIf(err)

	if testing.Verbose() {
		t.saveArtifacts("v9", v9Artifacts)
	}

	// Validate individual node details in the rewards file
	rewardsFile := v8Artifacts.RewardsFile
	nodeSummary := history.GetNodeSummary()

	singleEightEthNodes := nodeSummary["single_eight_eth"]
	singleSixteenEthNodes := nodeSummary["single_sixteen_eth"]
	for _, node := range append(singleEightEthNodes, singleSixteenEthNodes...) {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)

		expectedRewardsAmount, _ := big.NewInt(0).SetString("1019308880071990649542", 10)

		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got 0 ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		if ethAmount.Sign() != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), "0")
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}
	}
	singleEightEthNodesSP := nodeSummary["single_eight_eth_sp"]
	singleSixteenEthNodesSP := nodeSummary["single_sixteen_eth_sp"]
	for _, node := range append(singleEightEthNodesSP, singleSixteenEthNodesSP...) {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)

		expectedRewardsAmount, _ := big.NewInt(0).SetString("1019308880071990649542", 10)

		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		expectedEthAmount := big.NewInt(0)
		if node.SmoothingPoolRegistrationState {
			if node.Class == "single_eight_eth_sp" {
				expectedEthAmount.SetString("1354725546842756912", 10)
			} else {
				// 16-eth minipools earn more eth! A bit less than double.
				expectedEthAmount.SetString("2292612463887742467", 10)
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
	}

	optingInNodesSP := append(
		nodeSummary["single_eight_eth_opted_in_quarter"],
		nodeSummary["single_sixteen_eth_opted_in_quarter"]...,
	)
	for _, node := range optingInNodesSP {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)

		// Node has 20 RPL and only 1 8-eth minpool which puts it above the linear curve
		expectedRewardsAmount := big.NewInt(0)
		if node.Class == "single_eight_eth_opted_in_quarter" {
			expectedRewardsAmount.SetString("1784353229014464268647", 10)
		} else {
			// 16-eth minipools earn less for the same RPL stake, due to RPIP-30
			expectedRewardsAmount.SetString("1310160289473732090952", 10)
		}

		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		expectedEthAmount := big.NewInt(0)
		if node.Class == "single_eight_eth_opted_in_quarter" {
			// About 3/4 what the full nodes got
			expectedEthAmount.SetString("1019397441188609162", 10)
		} else {
			// 16-eth minipools earn more eth! A bit less than double.
			expectedEthAmount.SetString("1725134131242261659", 10)
		}
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}
	}

	optingOutNodesSP := append(
		nodeSummary["single_eight_eth_opted_out_three_quarters"],
		nodeSummary["single_sixteen_eth_opted_out_three_quarters"]...,
	)
	for _, node := range optingOutNodesSP {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)

		// Node has 20 RPL and only 1 8-eth minpool which puts it above the linear curve
		expectedRewardsAmount := big.NewInt(0)
		if node.Class == "single_eight_eth_opted_out_three_quarters" {
			expectedRewardsAmount.SetString("1784353229014464268647", 10)
		} else {
			// 16-eth minipools earn less for the same RPL stake, due to RPIP-30
			expectedRewardsAmount.SetString("1310160289473732090952", 10)
		}

		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		expectedEthAmount := big.NewInt(0)
		if node.Class == "single_eight_eth_opted_out_three_quarters" {
			// About 3/4 what the full nodes got
			expectedEthAmount.SetString("1005984316962443252", 10)
		} else {
			// 16-eth minipools earn more eth! A bit less than double.
			expectedEthAmount.SetString("1702434997936442426", 10)
		}
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}
	}

	bondReductionNode := nodeSummary["single_bond_reduction"]
	for _, node := range bondReductionNode {
		// Check the rewards amount in the rewards file
		rewardsAmount := rewardsFile.GetNodeCollateralRpl(node.Address)

		// Nodes that bond reduce are treated as having their new bond for the full interval,
		// when it comes to RPL rewards.
		expectedRewardsAmount, _ := big.NewInt(0).SetString("1019308880071990649542", 10)

		if rewardsAmount.Cmp(expectedRewardsAmount) != 0 {
			t.Fatalf("Rewards amount does not match expected value for node %s: %s != %s", node.Notes, rewardsAmount.String(), expectedRewardsAmount.String())
		}

		// Make sure it got reduced ETH
		ethAmount := rewardsFile.GetNodeSmoothingPoolEth(node.Address)
		expectedEthAmount, _ := big.NewInt(0).SetString("1922203879488237721", 10)
		if ethAmount.Cmp(expectedEthAmount) != 0 {
			t.Fatalf("ETH amount does not match expected value for node %s: %s != %s", node.Notes, ethAmount.String(), expectedEthAmount.String())
		}

		// Make sure it got 0 oDAO rpl
		oDaoRplAmount := rewardsFile.GetNodeOracleDaoRpl(node.Address)
		if oDaoRplAmount.Sign() != 0 {
			t.Fatalf("oDAO rpl amount does not match expected value for node %s: %s != %s", node.Notes, oDaoRplAmount.String(), "0")
		}
	}

	noMinipoolsNodes := nodeSummary["no_minipools"]
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
	}

	// Validate merkle root
	v8MerkleRoot := v8Artifacts.RewardsFile.GetMerkleRoot()
	v9MerkleRoot := v9Artifacts.RewardsFile.GetMerkleRoot()

	if !strings.EqualFold(v8MerkleRoot, v9MerkleRoot) {
		t.Fatalf("Merkle root does not match %s != %s", v8MerkleRoot, v9MerkleRoot)
	} else {
		t.Logf("v8/v9 Merkle root matches %s", v8MerkleRoot)
	}

	// Expected merkle root:
	// 0x9915d949936995f9045d26c3ef919194445377e83f1be2da47d181ee9ce705d8
	//
	// If this does not match, it implies either you updated the set of default mock nodes,
	// or you introduced a regression in treegen.
	// DO NOT update this value unless you know what you are doing.
	expectedMerkleRoot := "0x9915d949936995f9045d26c3ef919194445377e83f1be2da47d181ee9ce705d8"
	if !strings.EqualFold(v8MerkleRoot, expectedMerkleRoot) {
		t.Fatalf("Merkle root does not match expected value %s != %s", v8MerkleRoot, expectedMerkleRoot)
	} else {
		t.Logf("Merkle root matches expected value %s", expectedMerkleRoot)
	}
}
