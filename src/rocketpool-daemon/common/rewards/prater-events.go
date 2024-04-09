package rewards

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/v2/rewards"
)

// This file contains hardcoded events for Prater for the intervals before the contracts emitted rewards snapshot events,
// so the Smartnode has something to "look up" during events collection.
var praterPrehistoryIntervalEvents []rewards.RewardsEvent = []rewards.RewardsEvent{
	// TX: 0xe90cb89e2f83fb6577a9cb27421ec802df090fea1b156fea8444d57e0c41040a
	{
		Index:             big.NewInt(0),
		ExecutionBlock:    big.NewInt(7279809),
		ConsensusBlock:    big.NewInt(3509855),
		MerkleRoot:        common.HexToHash("0xae2f4ca79cb5176e7262201486739c614fd9197a9d95b1a28da23498a25000c4"),
		MerkleTreeCID:     "bafybeifcixenfbmzsbaar7zowpiyn3rh42px2sgci3zhc3oecr5ixfm754",
		IntervalsPassed:   big.NewInt(2),
		TreasuryRPL:       parseBigInt("3389231943988237492628"),
		TrustedNodeRPL:    []*big.Int{parseBigInt("3389231943988237492032")},
		NodeRPL:           []*big.Int{parseBigInt("15816415738611774962240")},
		NodeETH:           []*big.Int{big.NewInt(0)},
		UserETH:           big.NewInt(0),
		IntervalStartTime: time.Unix(1658107596, 0),
		IntervalEndTime:   time.Unix(1658625996, 0),
		SubmissionTime:    time.Unix(1658740664, 0),
	},

	// TX: 0x054bf4ca2c8649199b70f87ee6dec96eabb3184b77d8d396e85cb46c693acc19
	{
		Index:             big.NewInt(1),
		ExecutionBlock:    big.NewInt(7296892),
		ConsensusBlock:    big.NewInt(3531455),
		MerkleRoot:        common.HexToHash("0x76a5d0de3c7f92dab4d606c7a66246c07d4dc4003e330fc645541e828e8d11eb"),
		MerkleTreeCID:     "bafybeifhelxls7xsdeh4j7gokmqfherpzoyzo576s2j5gi42wl7zyhukeq",
		IntervalsPassed:   big.NewInt(1),
		TreasuryRPL:       parseBigInt("1134827945857665760894"),
		TrustedNodeRPL:    []*big.Int{parseBigInt("1134827945857665760312")},
		NodeRPL:           []*big.Int{parseBigInt("5295863747335773547586")},
		NodeETH:           []*big.Int{big.NewInt(565216198390830158)},
		UserETH:           big.NewInt(0),
		IntervalStartTime: time.Unix(1658625996, 0),
		IntervalEndTime:   time.Unix(1658885196, 0),
		SubmissionTime:    time.Unix(1658887477, 0),
	},

	// TX: 0xb70877c8cb8f63f59f67a2460f261b317d122349d869fea0796b7c692751eba7
	{
		Index:             big.NewInt(2),
		ExecutionBlock:    big.NewInt(7314096),
		ConsensusBlock:    big.NewInt(3553055),
		MerkleRoot:        common.HexToHash("0x2ef194dbd254c797c263089f9aa49132843a4b2881a2e8aa9c7456a06a59e6a4"),
		MerkleTreeCID:     "bafybeihsny3shsgsgiaft3u5oceur2wwj2quyspws7wceqpo6rlinriu3q",
		IntervalsPassed:   big.NewInt(1),
		TreasuryRPL:       parseBigInt("1135283120200590076960"),
		TrustedNodeRPL:    []*big.Int{parseBigInt("1135283120200590076448")},
		NodeRPL:           []*big.Int{parseBigInt("5297987894269420356281")},
		NodeETH:           []*big.Int{big.NewInt(245851786378530887)},
		UserETH:           big.NewInt(0),
		IntervalStartTime: time.Unix(1658885196, 0),
		IntervalEndTime:   time.Unix(1659144396, 0),
		SubmissionTime:    time.Unix(1659146709, 0),
	},

	// TX: 0x36baad3ecb07d9c1751b3f76f7378c23c0844296a333eafaa42d3fced5be0e93
	{
		Index:             big.NewInt(3),
		ExecutionBlock:    big.NewInt(7331289),
		ConsensusBlock:    big.NewInt(3574655),
		MerkleRoot:        common.HexToHash("0x64a1cba078633b82c12e32f5e6d94117c85bb41cfa1e237b60757d7e45d92ecb"),
		MerkleTreeCID:     "bafybeigbmzcdw2dbil5qkqiztqf3httceyd2fitwtk3aaf5x533x67dk2i",
		IntervalsPassed:   big.NewInt(1),
		TreasuryRPL:       parseBigInt("1135738477111879239196"),
		TrustedNodeRPL:    []*big.Int{parseBigInt("1135738477111879238408")},
		NodeRPL:           []*big.Int{parseBigInt("5300112893188769778496")},
		NodeETH:           []*big.Int{big.NewInt(106846466469715898)},
		UserETH:           big.NewInt(0),
		IntervalStartTime: time.Unix(1659144396, 0),
		IntervalEndTime:   time.Unix(1659403596, 0),
		SubmissionTime:    time.Unix(1659406463, 0),
	},

	// TX: 0xbd46261e656c6f446a127ca58484e1c9c6a68e735e2cb32d90a290062cb4efd9
	{
		Index:             big.NewInt(4),
		ExecutionBlock:    big.NewInt(7382763),
		ConsensusBlock:    big.NewInt(3639455),
		MerkleRoot:        common.HexToHash("0x7ff29433dead5b3f4f289b7ecbede03e0b8a2377748191584471283035e93fed"),
		MerkleTreeCID:     "bafybeiezkea32s4npgwv3nlkjfn6er4ohhiamnkl2if5ugf7cqmd5dnjne",
		IntervalsPassed:   big.NewInt(3),
		TreasuryRPL:       parseBigInt("3409949399585607533077"),
		TrustedNodeRPL:    []*big.Int{parseBigInt("3409949399585607532568")},
		NodeRPL:           []*big.Int{parseBigInt("15913097198066168484824")},
		NodeETH:           []*big.Int{parseBigInt("23844561349772065836")},
		UserETH:           big.NewInt(0),
		IntervalStartTime: time.Unix(1659403596, 0),
		IntervalEndTime:   time.Unix(1660181196, 0),
		SubmissionTime:    time.Unix(1660252236, 0),
	},

	// TX: 0xab81c7882d1646c31acab36145739ed87f76693e23eba50171a1dffd2d8aa4e0
	{
		Index:             big.NewInt(5),
		ExecutionBlock:    big.NewInt(7400317),
		ConsensusBlock:    big.NewInt(3661055),
		MerkleRoot:        common.HexToHash("0x9b6b9689f5638b3a86b356b3c5acdc44167c25415a3e3209c5c78a5f92567bb8"),
		MerkleTreeCID:     "bafybeiaukxxdzni6twjvebjamomtdkic4knccqdgdjwiunqfzfzohwtgq4",
		IntervalsPassed:   big.NewInt(1),
		TreasuryRPL:       parseBigInt("1137561731905671847934"),
		TrustedNodeRPL:    []*big.Int{parseBigInt("1137561731905671847312")},
		NodeRPL:           []*big.Int{parseBigInt("5308621415559801953508")},
		NodeETH:           []*big.Int{parseBigInt("10484778253257561700")},
		UserETH:           big.NewInt(0),
		IntervalStartTime: time.Unix(1660181196, 0),
		IntervalEndTime:   time.Unix(1660440396, 0),
		SubmissionTime:    time.Unix(1660609908, 0),
	},
}

// Parses a string into a BigInt, panicking if it's not legal
func parseBigInt(number string) *big.Int {
	result, success := big.NewInt(0).SetString(number, 10)
	if !success {
		panic(fmt.Sprintf("Error parsing Prater precompiled value %s", number))
	}

	return result
}
