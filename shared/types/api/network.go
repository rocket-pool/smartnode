package api

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	sharedtypes "github.com/rocket-pool/smartnode/v2/shared/types"
)

type NetworkNodeFeeData struct {
	NodeFee       *big.Int `json:"nodeFee"`
	MinNodeFee    *big.Int `json:"minNodeFee"`
	TargetNodeFee *big.Int `json:"targetNodeFee"`
	MaxNodeFee    *big.Int `json:"maxNodeFee"`
}

type NetworkRplPriceData struct {
	RplPrice                    *big.Int `json:"rplPrice"`
	RplPriceBlock               uint64   `json:"rplPriceBlock"`
	MinPer8EthMinipoolRplStake  *big.Int `json:"minPer8EthMinipoolRplStake"`
	MinPer16EthMinipoolRplStake *big.Int `json:"minPer16EthMinipoolRplStake"`
}

type NetworkStatsData struct {
	TotalValueLocked          *big.Int       `json:"totalValueLocked"`
	DepositPoolBalance        *big.Int       `json:"depositPoolBalance"`
	MinipoolCapacity          *big.Int       `json:"minipoolCapacity"`
	StakerUtilization         *big.Int       `json:"stakerUtilization"`
	NodeFee                   *big.Int       `json:"nodeFee"`
	NodeCount                 uint64         `json:"nodeCount"`
	InitializedMinipoolCount  uint64         `json:"initializedMinipoolCount"`
	PrelaunchMinipoolCount    uint64         `json:"prelaunchMinipoolCount"`
	StakingMinipoolCount      uint64         `json:"stakingMinipoolCount"`
	WithdrawableMinipoolCount uint64         `json:"withdrawableMinipoolCount"`
	DissolvedMinipoolCount    uint64         `json:"dissolvedMinipoolCount"`
	FinalizedMinipoolCount    uint64         `json:"finalizedMinipoolCount"`
	RplPrice                  *big.Int       `json:"rplPrice"`
	TotalRplStaked            *big.Int       `json:"totalRplStaked"`
	EffectiveRplStaked        *big.Int       `json:"effectiveRplStaked"`
	RethPrice                 *big.Int       `json:"rethPrice"`
	SmoothingPoolNodes        uint64         `json:"smoothingPoolNodes"`
	SmoothingPoolAddress      common.Address `json:"SmoothingPoolAddress"`
	SmoothingPoolBalance      *big.Int       `json:"smoothingPoolBalance"`
}

type NetworkTimezonesData struct {
	TimezoneCounts map[string]uint64 `json:"timezoneCounts"`
	TimezoneTotal  uint64            `json:"timezoneTotal"`
	NodeTotal      uint64            `json:"nodeTotal"`
}

type NetworkRewardsFileData struct {
	CurrentIndex   uint64 `json:"currentIndex"`
	TreeFileExists bool   `json:"treeFileExists"`
}

type NetworkDaoProposalsData struct {
	AccountAddress          common.Address                  `json:"accountAddress"`
	VotingDelegate          common.Address                  `json:"votingDelegate"`
	ActiveSnapshotProposals []*sharedtypes.SnapshotProposal `json:"activeSnapshotProposals"`
}

type NetworkLatestDelegateData struct {
	Address common.Address `json:"address"`
}

type NetworkDepositContractInfoData struct {
	Status                string         `json:"status"`
	Error                 string         `json:"error"`
	RPDepositContract     common.Address `json:"rpDepositContract"`
	RPNetwork             uint64         `json:"rpNetwork"`
	BeaconDepositContract common.Address `json:"beaconDepositContract"`
	BeaconNetwork         uint64         `json:"beaconNetwork"`
	SufficientSync        bool           `json:"sufficientSync"`
}

func (ndcid *NetworkDepositContractInfoData) Mismatched() bool {
	return ndcid.RPNetwork != ndcid.BeaconNetwork ||
		ndcid.RPDepositContract != ndcid.BeaconDepositContract
}

func (ndcid *NetworkDepositContractInfoData) PrintMismatch() bool {
	if !ndcid.Mismatched() {
		fmt.Println("Your Beacon Node is on the correct network.")
		fmt.Println()
		return false
	}
	fmt.Printf("%s***ALERT***\n", terminal.ColorRed)
	fmt.Println("YOUR ETH2 CLIENT IS NOT CONNECTED TO THE SAME NETWORK THAT ROCKET POOL IS USING!")
	fmt.Println("This is likely because your ETH2 client is using the wrong configuration.")
	fmt.Println("For the safety of your funds, Rocket Pool will not let you deposit your ETH until this is resolved.")
	fmt.Println()
	fmt.Println("To fix it if you are in Docker mode:")
	fmt.Println("\t1. Run 'rocketpool service install -d' to get the latest configuration")
	fmt.Println("\t2. Run 'rocketpool service stop' and 'rocketpool service start' to apply the configuration.")
	fmt.Println("If you are using Hybrid or Native mode, please correct the network flags in your ETH2 launch script.")
	fmt.Println()
	fmt.Println("Details:")
	fmt.Printf("\tRocket Pool expects deposit contract %s on chain %d.\n", ndcid.RPDepositContract.Hex(), ndcid.RPNetwork)
	fmt.Printf("\tYour Beacon client is using deposit contract %s on chain %d.%s\n", ndcid.BeaconDepositContract.Hex(), ndcid.BeaconNetwork, terminal.ColorReset)
	return true
}

type NetworkHotfixDeployedData struct {
	IsHoustonHotfixDeployed bool `json:"isHoustonHotfixDeployed"`
}
