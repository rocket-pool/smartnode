package rewards

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
)

const (
	rewardsSnapshotSubmittedNodeKey string = "rewards.snapshot.submitted.node.key"
)

// Info for a rewards snapshot event
type RewardsEvent struct {
	Index             *big.Int
	ExecutionBlock    *big.Int
	ConsensusBlock    *big.Int
	MerkleRoot        common.Hash
	MerkleTreeCID     string
	IntervalsPassed   *big.Int
	TreasuryRPL       *big.Int
	TrustedNodeRPL    []*big.Int
	NodeRPL           []*big.Int
	NodeETH           []*big.Int
	UserETH           *big.Int
	IntervalStartTime time.Time
	IntervalEndTime   time.Time
	SubmissionTime    time.Time
}

// Struct for submitting the rewards for a checkpoint
type RewardSubmission struct {
	RewardIndex     *big.Int   `json:"rewardIndex"`
	ExecutionBlock  *big.Int   `json:"executionBlock"`
	ConsensusBlock  *big.Int   `json:"consensusBlock"`
	MerkleRoot      [32]byte   `json:"merkleRoot"`
	MerkleTreeCID   string     `json:"merkleTreeCID"`
	IntervalsPassed *big.Int   `json:"intervalsPassed"`
	TreasuryRPL     *big.Int   `json:"treasuryRPL"`
	TrustedNodeRPL  []*big.Int `json:"trustedNodeRPL"`
	NodeRPL         []*big.Int `json:"nodeRPL"`
	NodeETH         []*big.Int `json:"nodeETH"`
	UserETH         *big.Int   `json:"userETH"`
}

// Internal struct - this is the structure of what gets returned by the RewardSnapshot event
type rewardSnapshot struct {
	RewardIndex       *big.Int         `json:"rewardIndex"`
	Submission        RewardSubmission `json:"submission"`
	IntervalStartTime *big.Int         `json:"intervalStartTime"`
	IntervalEndTime   *big.Int         `json:"intervalEndTime"`
	Time              *big.Int         `json:"time"`
}

// Get the timestamp that the current rewards interval started
func GetClaimIntervalTimeStart(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Time, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return time.Time{}, err
	}
	unixTime := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, unixTime, "getClaimIntervalTimeStart"); err != nil {
		return time.Time{}, fmt.Errorf("error getting claim interval time start: %w", err)
	}
	return time.Unix((*unixTime).Int64(), 0), nil
}

// Get the number of seconds in a claim interval
func GetClaimIntervalTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return 0, err
	}
	unixTime := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, unixTime, "getClaimIntervalTime"); err != nil {
		return 0, fmt.Errorf("error getting claim interval time: %w", err)
	}
	return time.Duration((*unixTime).Int64()) * time.Second, nil
}

// Get the percent of checkpoint rewards that goes to node operators
func GetNodeOperatorRewardsPercent(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return nil, err
	}
	perc := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, perc, "getClaimingContractPerc", "rocketClaimNode"); err != nil {
		return nil, fmt.Errorf("error getting node operator rewards percent: %w", err)
	}
	return *perc, nil
}

// Get the percent of checkpoint rewards that goes to ODAO members
func GetTrustedNodeOperatorRewardsPercent(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return nil, err
	}
	perc := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, perc, "getClaimingContractPerc", "rocketClaimTrustedNode"); err != nil {
		return nil, fmt.Errorf("error getting trusted node operator rewards percent: %w", err)
	}
	return *perc, nil
}

// Get the percent of checkpoint rewards that goes to the PDAO
func GetProtocolDaoRewardsPercent(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return nil, err
	}
	perc := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, perc, "getClaimingContractPerc", "rocketClaimDAO"); err != nil {
		return nil, fmt.Errorf("error getting protocol DAO rewards percent: %w", err)
	}
	return *perc, nil
}

// Get the amount of RPL rewards that will be provided to node operators
func GetPendingRPLRewards(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return nil, err
	}
	rewards := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, rewards, "getPendingRPLRewards"); err != nil {
		return nil, fmt.Errorf("error getting pending RPL rewards: %w", err)
	}
	return *rewards, nil
}

// Get the amount of ETH rewards that will be provided to node operators
func GetPendingETHRewards(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return nil, err
	}
	rewards := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, rewards, "getPendingETHRewards"); err != nil {
		return nil, fmt.Errorf("error getting pending ETH rewards: %w", err)
	}
	return *rewards, nil
}

// Check whether or not the given address has submitted for the given rewards interval
func GetTrustedNodeSubmitted(rp *rocketpool.RocketPool, nodeAddress common.Address, rewardsIndex uint64, opts *bind.CallOpts) (bool, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return false, err
	}

	indexBig := big.NewInt(0).SetUint64(rewardsIndex)
	hasSubmitted := new(bool)
	if err := rocketRewardsPool.Call(opts, hasSubmitted, "getTrustedNodeSubmitted", nodeAddress, indexBig); err != nil {
		return false, fmt.Errorf("error getting trusted node submission status: %w", err)
	}
	return *hasSubmitted, nil
}

// Check whether or not the given address has submitted specific rewards info
func GetTrustedNodeSubmittedSpecificRewards(rp *rocketpool.RocketPool, nodeAddress common.Address, submission RewardSubmission, opts *bind.CallOpts) (bool, error) {
	// NOTE: this doesn't have a view yet so we have to construct it manually, and RLP
	stringTy, _ := abi.NewType("string", "string", nil)
	addressTy, _ := abi.NewType("address", "address", nil)

	submissionTy, _ := abi.NewType("tuple", "struct RewardSubmission", []abi.ArgumentMarshaling{
		{Name: "rewardIndex", Type: "uint256"},
		{Name: "executionBlock", Type: "uint256"},
		{Name: "consensusBlock", Type: "uint256"},
		{Name: "merkleRoot", Type: "bytes32"},
		{Name: "merkleTreeCID", Type: "string"},
		{Name: "intervalsPassed", Type: "uint256"},
		{Name: "treasuryRPL", Type: "uint256"},
		{Name: "trustedNodeRPL", Type: "uint256[]"},
		{Name: "nodeRPL", Type: "uint256[]"},
		{Name: "nodeETH", Type: "uint256[]"},
		{Name: "userETH", Type: "uint256"},
	})

	args := abi.Arguments{
		{Type: stringTy, Name: "key"},
		{Type: addressTy, Name: "trustedNodeAddress"},
		{Type: submissionTy, Name: "submission"},
	}

	bytes, err := args.Pack(rewardsSnapshotSubmittedNodeKey, nodeAddress, &submission)
	if err != nil {
		return false, fmt.Errorf("error encoding submission data into ABI format: %w", err)
	}

	key := crypto.Keccak256Hash(bytes)
	result, err := rp.RocketStorage.GetBool(opts, key)
	if err != nil {
		return false, fmt.Errorf("error checking if trusted node submitted specific rewards: %w", err)
	}
	return result, nil
}

// Estimate the gas for submitting a Merkle Tree-based snapshot for a rewards interval
func EstimateSubmitRewardSnapshotGas(rp *rocketpool.RocketPool, submission RewardSubmission, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketRewardsPool.GetTransactionGasInfo(opts, "submitRewardSnapshot", submission)
}

// Submit a Merkle Tree-based snapshot for a rewards interval
func SubmitRewardSnapshot(rp *rocketpool.RocketPool, submission RewardSubmission, opts *bind.TransactOpts) (common.Hash, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketRewardsPool.Transact(opts, "submitRewardSnapshot", submission)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error submitting rewards snapshot: %w", err)
	}
	return tx.Hash(), nil
}

// Get the event info for a rewards snapshot using the Atlas getter
func GetRewardsEvent(rp *rocketpool.RocketPool, index uint64, rocketRewardsPoolAddresses []common.Address, opts *bind.CallOpts) (bool, RewardsEvent, error) {
	// Check if the client is requesting interval 0 on mainnet, then return the hardcoded RewardsEvent
	data, ok, err := getMainnetIntervalRewardsEvent(rp, index)
	if err != nil {
		return false, RewardsEvent{}, err
	}
	if ok {
		return true, data, nil
	}

	// Get contracts
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return false, RewardsEvent{}, err
	}

	// Get the block that the event was emitted on
	indexBig := big.NewInt(0).SetUint64(index)
	blockWrapper := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, blockWrapper, "getClaimIntervalExecutionBlock", indexBig); err != nil {
		return false, RewardsEvent{}, fmt.Errorf("error getting the event block for interval %d: %w", index, err)
	}
	block := *blockWrapper

	// Create the list of addresses to check
	currentAddress := *rocketRewardsPool.Address
	if rocketRewardsPoolAddresses == nil {
		rocketRewardsPoolAddresses = []common.Address{currentAddress}
	} else {
		found := slices.Contains(rocketRewardsPoolAddresses, currentAddress)
		if !found {
			rocketRewardsPoolAddresses = append(rocketRewardsPoolAddresses, currentAddress)
		}
	}

	// Construct a filter query for relevant logs
	rewardsSnapshotEvent := rocketRewardsPool.ABI.Events["RewardSnapshot"]
	indexBytes := [32]byte{}
	indexBig.FillBytes(indexBytes[:])
	addressFilter := rocketRewardsPoolAddresses
	topicFilter := [][]common.Hash{{rewardsSnapshotEvent.ID}, {indexBytes}}

	// Get the event logs
	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, big.NewInt(1), block, block, nil)
	if err != nil {
		return false, RewardsEvent{}, err
	}
	if len(logs) == 0 {
		return false, RewardsEvent{}, nil
	}

	// Get the log info values
	values, err := rewardsSnapshotEvent.Inputs.Unpack(logs[0].Data)
	if err != nil {
		return false, RewardsEvent{}, fmt.Errorf("error unpacking rewards snapshot event data: %w", err)
	}

	// Convert to a native struct
	var snapshot rewardSnapshot
	err = rewardsSnapshotEvent.Inputs.Copy(&snapshot, values)
	if err != nil {
		return false, RewardsEvent{}, fmt.Errorf("error converting rewards snapshot event data to struct: %w", err)
	}

	// Get the decoded data
	submission := snapshot.Submission
	eventData := RewardsEvent{
		Index:             indexBig,
		ExecutionBlock:    submission.ExecutionBlock,
		ConsensusBlock:    submission.ConsensusBlock,
		IntervalsPassed:   submission.IntervalsPassed,
		TreasuryRPL:       submission.TreasuryRPL,
		TrustedNodeRPL:    submission.TrustedNodeRPL,
		NodeRPL:           submission.NodeRPL,
		NodeETH:           submission.NodeETH,
		UserETH:           submission.UserETH,
		MerkleRoot:        submission.MerkleRoot,
		MerkleTreeCID:     submission.MerkleTreeCID,
		IntervalStartTime: time.Unix(snapshot.IntervalStartTime.Int64(), 0),
		IntervalEndTime:   time.Unix(snapshot.IntervalEndTime.Int64(), 0),
		SubmissionTime:    time.Unix(snapshot.Time.Int64(), 0),
	}

	// Convert v1.1.0-rc1 events to modern ones
	if eventData.UserETH == nil {
		eventData.UserETH = big.NewInt(0)
	}

	return true, eventData, nil
}

// Check if the client is requesting a hardcoded interval on mainnet, then return the RewardsEvent
func getMainnetIntervalRewardsEvent(rp *rocketpool.RocketPool, index uint64) (RewardsEvent, bool, error) {
	// Check if the ec is synced to mainnet
	chainID, err := rp.Client.ChainID(context.Background())
	if err != nil {
		return RewardsEvent{}, false, fmt.Errorf("error getting chainID: %w", err)
	}
	if chainID.Cmp(big.NewInt(1)) != 0 {
		return RewardsEvent{}, false, nil
	}

	var (
		treasuryRPL       *big.Int
		trustedNodeRPL    *big.Int
		nodeRPL           *big.Int
		nodeETH           *big.Int
		executionBlock    *big.Int
		consensusBlock    *big.Int
		intervalsPassed   *big.Int
		merkleRoot        common.Hash
		merkleTreeCID     string
		intervalStartTime time.Time
		intervalEndTime   time.Time
		submissionTime    time.Time
		userETH           *big.Int
	)

	// Hardcoded RewardsEvent for old intervals on mainnet
	switch index {
	case 0:
		treasuryRPL.SetString("10633670478560109530497", 10)
		trustedNodeRPL.SetString("10633670478560109529794", 10)
		nodeRPL.SetString("49623795566613844471758", 10)
		nodeETH = big.NewInt(0)
		executionBlock = big.NewInt(15451078)
		consensusBlock = big.NewInt(4598879)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xb839fa0f5842bf3c8f19091361889fb0f1cb399d64b8da476d372b7de7a93463")
		merkleTreeCID = "bafybeidrck3sz24acv32h56xdb7ruarxq52oci32del7moxqtief3do73y"
		intervalStartTime = time.Unix(1659591339, 0)
		intervalEndTime = time.Unix(1662010539, 0)
		submissionTime = time.Unix(1662011717, 0)
		userETH = big.NewInt(0)
	case 1:
		treasuryRPL.SetString("10550049308997708584809", 10)
		trustedNodeRPL.SetString("10550049308997708584060", 10)
		nodeRPL.SetString("49233563441989306724917", 10)
		nodeETH.SetString("55886528290134709468", 10)
		executionBlock = big.NewInt(15636954)
		consensusBlock = big.NewInt(4800479)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xb060f0964ce14117075608a69835f4e5e3b872936d3fba2dbb17e202b5c2a7d1")
		merkleTreeCID = "bafybeiabhjduq23d6yncrmook3hhw6d2lp6wm4rwav6mkwh7uzmimw4ona"
		intervalStartTime = time.Unix(1662010539, 0)
		intervalEndTime = time.Unix(1664429739, 0)
		submissionTime = time.Unix(1664436887, 0)
		userETH.SetString("41139205675101362849", 10)
	case 2:
		treasuryRPL.SetString("10589610096624449116454", 10)
		trustedNodeRPL.SetString("10589610096624449115610", 10)
		nodeRPL.SetString("49418180450914095872087", 10)
		nodeETH.SetString("120666426232176747457", 10)
		executionBlock = big.NewInt(15837359)
		consensusBlock = big.NewInt(5002079)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x278fd75797e2a9eddc128c0199b448877e30d1196c12306bdc95fb731647c18f")
		merkleTreeCID = "bafybeihi4m4jdj76746gzqvwxfphvocqrpcylqbs2b746kryoajjqrppzu"
		intervalStartTime = time.Unix(1664429739, 0)
		intervalEndTime = time.Unix(1666848939, 0)
		submissionTime = time.Unix(1666874939, 0)
		userETH.SetString("88931841842952009134", 10)
	case 3:
		treasuryRPL.SetString("10629319230090323621908", 10)
		trustedNodeRPL.SetString("10629319230090323621082", 10)
		nodeRPL.SetString("49603489740421510230963", 10)
		nodeETH.SetString("178060769200919879825", 10)
		executionBlock = big.NewInt(16037791)
		consensusBlock = big.NewInt(5203679)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xc0c736dccc4371b8a9d4ded4b50213dd820504430350d157350f52dfd04869ab")
		merkleTreeCID = "bafybeihhngkpe7aoz3uk6aocjujcbssjjbjtzzh6qzau74vixzsovjb5ae"
		intervalStartTime = time.Unix(1666848939, 0)
		intervalEndTime = time.Unix(1669268139, 0)
		submissionTime = time.Unix(1669275119, 0)
		userETH.SetString("131334112536558744486", 10)
	case 4:
		treasuryRPL.SetString("10669177265665550884567", 10)
		trustedNodeRPL.SetString("10669177265665550883677", 10)
		nodeRPL.SetString("49789493906439237456308", 10)
		nodeETH.SetString("90137212013211727003", 10)
		executionBlock = big.NewInt(16238190)
		consensusBlock = big.NewInt(5405279)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x21c047f0035a64ca5b21c42bdad08d329f2121b5f1fe47e51222f9701f373826")
		merkleTreeCID = "bafybeicdqywie7v7a73y4eh6jtdqtronsduvq57e5v6747lqynqkqze4am"
		intervalStartTime = time.Unix(1669268139, 0)
		intervalEndTime = time.Unix(1671687339, 0)
		submissionTime = time.Unix(1671695987, 0)
		userETH.SetString("66619163202260233947", 10)
	case 5:
		treasuryRPL.SetString("10709184761706262970440", 10)
		trustedNodeRPL.SetString("10709184761706262969440", 10)
		nodeRPL.SetString("49976195554629227189775", 10)
		nodeETH.SetString("100959865151123783128", 10)
		executionBlock = big.NewInt(16438804)
		consensusBlock = big.NewInt(5606879)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xde0604812161f69253e6bdaf0f623c63e1d7c49d43a78b5333564b24dae2e450")
		merkleTreeCID = "bafybeia5iqj7yzwfn77xpmf3hqtu7k52h4fpiyan4pkz2enyp6rbmvkotq"
		intervalStartTime = time.Unix(1671687339, 0)
		intervalEndTime = time.Unix(1674106539, 0)
		submissionTime = time.Unix(1674113807, 0)
		userETH.SetString("74626075394052241951", 10)
	case 6:
		treasuryRPL.SetString("10749342278662327028056", 10)
		trustedNodeRPL.SetString("10749342278662327026976", 10)
		nodeRPL.SetString("50163597300424192791516", 10)
		nodeETH.SetString("152862142797510957810", 10)
		executionBlock = big.NewInt(16639245)
		consensusBlock = big.NewInt(5808479)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x804d82caab67b8bcfff14f044cb6745ed3ba59c5fd06f52c27b25c96c95e3290")
		merkleTreeCID = "bafybeid3fzvkb2bihyy4gpx555yagxytotfjkimauehy7dakehkdu5d63y"
		intervalStartTime = time.Unix(1674106539, 0)
		intervalEndTime = time.Unix(1676525739, 0)
		submissionTime = time.Unix(1676533163, 0)
		userETH.SetString("112999364182148607952", 10)
	case 7:
		treasuryRPL.SetString("10789650379085196418454", 10)
		trustedNodeRPL.SetString("10789650379085196417433", 10)
		nodeRPL.SetString("50351701769064249947084", 10)
		nodeETH.SetString("214889774917327718259", 10)
		executionBlock = big.NewInt(16838374)
		consensusBlock = big.NewInt(6010079)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x64692e59bb20ed13f5401b659fb53868c319a9dad410cb5b3c35a99c56db0ba1")
		merkleTreeCID = "bafybeihhsblaladk7e2bsudr455cb2qt22ireq6xp6mvfzxfnu45pguxey"
		intervalStartTime = time.Unix(1676525739, 0)
		intervalEndTime = time.Unix(1678944939, 0)
		submissionTime = time.Unix(1678986371, 0)
		userETH.SetString("158900489644285001733", 10)
	case 8:
		treasuryRPL.SetString("10830109627635791285926", 10)
		trustedNodeRPL.SetString("10830109627635791284908", 10)
		nodeRPL.SetString("50540511595633692661988", 10)
		nodeETH.SetString("124277451394130356705", 10)
		executionBlock = big.NewInt(17036705)
		consensusBlock = big.NewInt(6211679)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x9e53fc60a6e330b8aece2d2475dcc23e17426c5a552c7b4374ddddd46fa6599b")
		merkleTreeCID = "bafybeiagazxkhdlyjqpba2hox4xhtf4jo5x5j4c3alc3kipfrt2g7seage"
		intervalStartTime = time.Unix(1678944939, 0)
		intervalEndTime = time.Unix(1681364139, 0)
		submissionTime = time.Unix(1681371923, 0)
		userETH.SetString("91918564718913412697", 10)
	case 9:
		treasuryRPL.SetString("10870720591092408678629", 10)
		trustedNodeRPL.SetString("10870720591092408677484", 10)
		nodeRPL.SetString("50730029425097907160549", 10)
		nodeETH.SetString("265294502885782199241", 10)
		executionBlock = big.NewInt(17235076)
		consensusBlock = big.NewInt(6413279)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x2238938633ae1f6a3275b52492dc64018f2245912b708f91bd21cc8472fa3f45")
		merkleTreeCID = "bafybeichtj4if3nwxfdlnmt4xcbyzhf4vooa3lgjcy7mtlkds7g2druivq"
		intervalStartTime = time.Unix(1681364139, 0)
		intervalEndTime = time.Unix(1683783339, 0)
		submissionTime = time.Unix(1683791075, 0)
		userETH.SetString("251254537064420777704", 10)
	case 10:
		treasuryRPL.SetString("10911483838358662330962", 10)
		trustedNodeRPL.SetString("10911483838358662329864", 10)
		nodeRPL.SetString("50920257912340424204973", 10)
		nodeETH.SetString("207224314619456280271", 10)
		executionBlock = big.NewInt(17433555)
		consensusBlock = big.NewInt(6614879)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xc16b52575ec0494ef72ec419f7660f65d35abe65a51c277e3a8b4f581988ab25")
		merkleTreeCID = "bafybeiaqpdzdpngpt6py464xjcheoilaggrscvzzlol7c6stvvbhhyg4pu"
		intervalStartTime = time.Unix(1683783339, 0)
		intervalEndTime = time.Unix(1686202539, 0)
		submissionTime = time.Unix(1686209267, 0)
		userETH.SetString("244724588999515442961", 10)
	case 11:
		treasuryRPL.SetString("10952399940471452219796", 10)
		trustedNodeRPL.SetString("10952399940471452218698", 10)
		nodeRPL.SetString("51111199722200110352863", 10)
		nodeETH.SetString("129839125242481699627", 10)
		executionBlock = big.NewInt(17632767)
		consensusBlock = big.NewInt(6816479)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x280677fb18f06701851bd057f52558e59c6a8534f31bf0222e424dd4e2b6d2d2")
		merkleTreeCID = "bafybeihoq5a4fdfagu3cd5zskbeklr7urza5pwg2htahtaf5v46qlzodmi"
		intervalStartTime = time.Unix(1686202539, 0)
		intervalEndTime = time.Unix(1688621739, 0)
		submissionTime = time.Unix(1688629187, 0)
		userETH.SetString("168470911056744138735", 10)
	case 12:
		treasuryRPL.SetString("16123755223559813871223", 10)
		trustedNodeRPL.SetString("5863183717658114134602", 10)
		nodeRPL.SetString("51302857529508498676908", 10)
		nodeETH.SetString("264844913795172053460", 10)
		executionBlock = big.NewInt(17832425)
		consensusBlock = big.NewInt(7018079)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xae4e2bb2dc3efee5a20bedc83be1a7b5450de064867edae2aae59e1df08313fe")
		merkleTreeCID = "bafybeiaiwlnry6uefw5pw2oq2xueeat5nmvsnfhfkcdyksrfx3ml6vydky"
		intervalStartTime = time.Unix(1688621739, 0)
		intervalEndTime = time.Unix(1691040939, 0)
		submissionTime = time.Unix(1691045351, 0)
		userETH.SetString("357819970818590187458", 10)
	case 13:
		treasuryRPL.SetString("16184216406011424405097", 10)
		trustedNodeRPL.SetString("5885169602185972510578", 10)
		nodeRPL.SetString("51495234019127259466761", 10)
		nodeETH.SetString("218680806556832509553", 10)
		executionBlock = big.NewInt(18032505)
		consensusBlock = big.NewInt(7219679)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x8e1a38feed2f3e88782a7bbb3ff24c650f3903beb3169c3d196b358ce70a9080")
		merkleTreeCID = "bafybeihalpqr5o2bmvzx2x4x7her33jh4k6s3ngbvypejjl52nss2oedzi"
		intervalStartTime = time.Unix(1691040939, 0)
		intervalEndTime = time.Unix(1693460139, 0)
		submissionTime = time.Unix(1693462895, 0)
		userETH.SetString("306490363490092102569", 10)
	case 14:
		treasuryRPL.SetString("16983309048252480903029", 10)
		trustedNodeRPL.SetString("5168833188598581144102", 10)
		nodeRPL.SetString("51688331885985811440053", 10)
		nodeETH.SetString("174468095251348093682", 10)
		executionBlock = big.NewInt(18232197)
		consensusBlock = big.NewInt(7421279)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x8b2e52188d616a19f50477d1d2d2e3ddc8a2a056f9486ec7cc25d7140c75217b")
		merkleTreeCID = "bafybeiem7aj4oz7soasl2nvyvl24bab62i33ofak65c5ozpqnngaoa5dma"
		intervalStartTime = time.Unix(1693460139, 0)
		intervalEndTime = time.Unix(1695879339, 0)
		submissionTime = time.Unix(1695887015, 0)
		userETH.SetString("246414308218062973622", 10)
	case 15:
		treasuryRPL.SetString("17046993402967695740540", 10)
		trustedNodeRPL.SetString("5188215383511907398948", 10)
		nodeRPL.SetString("51882153835119073988521", 10)
		nodeETH.SetString("113064713982751905771", 10)
		executionBlock = big.NewInt(18432348)
		consensusBlock = big.NewInt(7622879)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xa1afd40dfecdde96b59d2a85eac5fd0c618bc336cdfad6ee9c093cf30a02335b")
		merkleTreeCID = "bafybeighiqr4fyqrsnkxbz5vhndmt4cuokg3vbc542jtxh2er4z35hynva"
		intervalStartTime = time.Unix(1695879339, 0)
		intervalEndTime = time.Unix(1698298539, 0)
		submissionTime = time.Unix(1698299831, 0)
		userETH.SetString("161261121697283843558", 10)
	case 16:
		treasuryRPL.SetString("17929264745987131427422", 10)
		trustedNodeRPL.SetString("4389322074743737569096", 10)
		nodeRPL.SetString("52076702581705360988290", 10)
		nodeETH.SetString("145511116579242423330", 10)
		executionBlock = big.NewInt(18632383)
		consensusBlock = big.NewInt(7824479)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x72515a30759d598e36aea4b2729ee38136da3da8c3d40c17e8ddaf6381746b61")
		merkleTreeCID = "bafybeiez5n7vn4jrzsk3dy7npdorqe2letdvyojqbo4ltxubzujaj63ndu"
		intervalStartTime = time.Unix(1698298539, 0)
		intervalEndTime = time.Unix(1700717739, 0)
		submissionTime = time.Unix(1700719223, 0)
		userETH.SetString("208347503555437494510", 10)
	case 17:
		treasuryRPL.SetString("17996496264451663466746", 10)
		trustedNodeRPL.SetString("4405781243164515122322", 10)
		nodeRPL.SetString("52271980851104416704641", 10)
		nodeETH.SetString("197988839753658327068", 10)
		executionBlock = big.NewInt(18832201)
		consensusBlock = big.NewInt(8026079)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x8466396383460910d80e9a5af0b962213b3a0e33c4a41f3b817a84487364c123")
		merkleTreeCID = "bafybeifldymulw6qvlfjgntj6mrlbwcl46xn6njickcydlquxc33nseoxi"
		intervalStartTime = time.Unix(1700717739, 0)
		intervalEndTime = time.Unix(1703136939, 0)
		submissionTime = time.Unix(1703138135, 0)
		userETH.SetString("285455180775757420844", 10)
	case 18:
		treasuryRPL.SetString("18063979889019768905884", 10)
		trustedNodeRPL.SetString("4422302130506914378890", 10)
		nodeRPL.SetString("52467991378895594323702", 10)
		nodeETH.SetString("119998368939736398237", 10)
		executionBlock = big.NewInt(19031703)
		consensusBlock = big.NewInt(8227679)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xd37ce5368a3790cfd9f3fff6a56079b0026a9a06f2fee02a5055f6838cfd5a00")
		merkleTreeCID = "bafybeigfgadbxblmn45ab7c2m7zt3bvimt66wwp543fl6a33hxfuiopvha"
		intervalStartTime = time.Unix(1703136939, 0)
		intervalEndTime = time.Unix(1705556139, 0)
		submissionTime = time.Unix(1705557263, 0)
		userETH.SetString("175552968220321206323", 10)
	case 19:
		treasuryRPL.SetString("18131716565043998308966", 10)
		trustedNodeRPL.SetString("4438884968205792116706", 10)
		nodeRPL.SetString("52664736910916177654259", 10)
		nodeETH.SetString("148353447414176591282", 10)
		executionBlock = big.NewInt(19231284)
		consensusBlock = big.NewInt(8429279)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x35d1be64d49aa71dc5b5ea13dd6f91d8613c81aef2593796d6dee599cd228aea")
		merkleTreeCID = "bafybeiazkzsqe7molppbhbxg2khdgocrip36eoezroa7anbe53za7mxjpq"
		intervalStartTime = time.Unix(1705556139, 0)
		intervalEndTime = time.Unix(1707975339, 0)
		submissionTime = time.Unix(1707976475, 0)
		userETH.SetString("219460060796213936870", 10)
	case 20:
		treasuryRPL.SetString("20691783336720225637003", 10)
		trustedNodeRPL.SetString("1963453893265422870516", 10)
		nodeRPL.SetString("52862220203299846512817", 10)
		nodeETH.SetString("129970171333179451391", 10)
		executionBlock = big.NewInt(19431215)
		consensusBlock = big.NewInt(8630879)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x55afb0387fb4c1c8e479f49433b996bce1d1c4658cf8e58f7a9ea3d3c5099eb2")
		merkleTreeCID = "bafybeiazuw7gubmv3csk54hqy6hhtr2uyvcsvhykw36ofbrrp436g23yme"
		intervalStartTime = time.Unix(1707975339, 0)
		intervalEndTime = time.Unix(1710394539, 0)
		submissionTime = time.Unix(1710395711, 0)
		userETH.SetString("194196448160947806147", 10)
	case 21:
		treasuryRPL.SetString("20769373803098840664266", 10)
		trustedNodeRPL.SetString("1970816492264853493584", 10)
		nodeRPL.SetString("53060444022515286364660", 10)
		nodeETH.SetString("86124043557534943014", 10)
		executionBlock = big.NewInt(19630338)
		consensusBlock = big.NewInt(8832479)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x7273138f110f464d45c3ec8b42aec0280b0e5a434a93907a5c9db047cb180866")
		merkleTreeCID = "bafybeiagdvc6chbfeyyttovpkhx4ltfyd3itah42lzmr445nxzjxgygmgu"
		intervalStartTime = time.Unix(1710394539, 0)
		intervalEndTime = time.Unix(1712813739, 0)
		submissionTime = time.Unix(1712824835, 0)
		userETH.SetString("129212548027034380183", 10)
	case 22:
		treasuryRPL.SetString("20847255219772791871405", 10)
		trustedNodeRPL.SetString("1978206699686469301548", 10)
		nodeRPL.SetString("53259411145404942733410", 10)
		nodeETH.SetString("87118787916582229383", 10)
		executionBlock = big.NewInt(19830438)
		consensusBlock = big.NewInt(9034079)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xc2724d07659bb0cb7bbc926b873a56226ebf368bea729c007e214f464e8aa69d")
		merkleTreeCID = "bafybeig73barjh4kyj2inqgq255zeefnl2pozz7dyi5wqfddqmmgulxopq"
		intervalStartTime = time.Unix(1712813739, 0)
		intervalEndTime = time.Unix(1715232939, 0)
		submissionTime = time.Unix(1715234147, 0)
		userETH.SetString("131641498401095297397", 10)
	case 23:
		treasuryRPL.SetString("20925428677753363338954", 10)
		trustedNodeRPL.SetString("1985624619056888491962", 10)
		nodeRPL.SetString("53459124359223920937447", 10)
		nodeETH.SetString("82988716547997166273", 10)
		executionBlock = big.NewInt(20030738)
		consensusBlock = big.NewInt(9235679)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xde037dafa7973dbab1cbe8cf46d85e2e125fd5bf193fc69b9f7811a5b6b7dc9d")
		merkleTreeCID = "bafybeifw6dledle4a4tnqti4b6dzvus3iezkkm2e32zaqliun4gnbgkwhm"
		intervalStartTime = time.Unix(1715232939, 0)
		intervalEndTime = time.Unix(1717652139, 0)
		submissionTime = time.Unix(1717653455, 0)
		userETH.SetString("125184956455787164441", 10)
	case 24:
		treasuryRPL.SetString("21847117345112177188009", 10)
		trustedNodeRPL.SetString("1149848281321693536186", 10)
		nodeRPL.SetString("53659586461679031688296", 10)
		nodeETH.SetString("81335123762071854501", 10)
		executionBlock = big.NewInt(20231099)
		consensusBlock = big.NewInt(9437279)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x2dccc9f3a86e7c857177c0f84abd314eeae162c398adcfbf5c94bd62812f1f1d")
		merkleTreeCID = "bafybeiebk3s32zmilztud4gxpwuagaddig4ildgqhomdx7eii6h5n7gzie"
		intervalStartTime = time.Unix(1717652139, 0)
		intervalEndTime = time.Unix(1720071339, 0)
		submissionTime = time.Unix(1720072511, 0)
		userETH.SetString("121962895985768410912", 10)
	case 25:
		treasuryRPL.SetString("21929040106251250266871", 10)
		trustedNodeRPL.SetString("1154160005592171066642", 10)
		nodeRPL.SetString("53860800260967983110475", 10)
		nodeETH.SetString("59836872361825739644", 10)
		executionBlock = big.NewInt(20431646)
		consensusBlock = big.NewInt(9638879)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xd8f3480bb44f993586001225adcc8f87f22d12184cd804d8f717dc9d26cacaa2")
		merkleTreeCID = "bafybeigb5v2oj7vcc7aarfqypiudksuyn3t2pkxegtvxjoynysttcxf5vy"
		intervalStartTime = time.Unix(1720071339, 0)
		intervalEndTime = time.Unix(1722490539, 0)
		submissionTime = time.Unix(1722491723, 0)
		userETH.SetString("89677960178182813533", 10)
	case 26:
		treasuryRPL.SetString("22011270063011907313857", 10)
		trustedNodeRPL.SetString("1158487898053258279638", 10)
		nodeRPL.SetString("54062768575818719716793", 10)
		nodeETH.SetString("73118591648836191604", 10)
		executionBlock = big.NewInt(20632194)
		consensusBlock = big.NewInt(9840479)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0xbffdc74404912050acccc79ec6556bd159f7a6f9db3be0c045519461c96047ef")
		merkleTreeCID = "bafybeiar5swdyqenqmdi25h5ihnyrcaf3vof5325mbsdb2w62fcvrvlpey"
		intervalStartTime = time.Unix(1722490539, 0)
		intervalEndTime = time.Unix(1724909739, 0)
		submissionTime = time.Unix(1724911139, 0)
		userETH.SetString("109309541129377923619", 10)
	case 27:
		treasuryRPL.SetString("22093808367322484339645", 10)
		trustedNodeRPL.SetString("1162832019332762333628", 10)
		nodeRPL.SetString("54265494235528908903092", 10)
		nodeETH.SetString("73502004443651237958", 10)
		executionBlock = big.NewInt(20832734)
		consensusBlock = big.NewInt(10042079)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x19f080c89283f80aa29cef2155457400d3e1a3989a16108f214fbf0e1a683e13")
		merkleTreeCID = "bafybeiecbehhk4xsp5eonauycaflwwpzbiq7dwtbvzochwzo6uotmebehu"
		intervalStartTime = time.Unix(1724909739, 0)
		intervalEndTime = time.Unix(1727328939, 0)
		submissionTime = time.Unix(1727330339, 0)
		userETH.SetString("109414335754265646582", 10)
	case 28:
		treasuryRPL.SetString("22176656175430841456904", 10)
		trustedNodeRPL.SetString("1167192430285833760860", 10)
		nodeRPL.SetString("54468980080005575506939", 10)
		nodeETH.SetString("65852251763664532329", 10)
		executionBlock = big.NewInt(21033406)
		consensusBlock = big.NewInt(10243679)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x5fd98bc90296e3ebcdf7098a539f3c04388aa639d885bacfea2b088796375d8c")
		merkleTreeCID = "bafybeihoaivrxlnqhstydoditx4pyr4waihduawgodieen6axv2km24imi"
		intervalStartTime = time.Unix(1727328939, 0)
		intervalEndTime = time.Unix(1729748139, 0)
		submissionTime = time.Unix(1729749551, 0)
		userETH.SetString("97856965684776269348", 10)
	case 29:
		treasuryRPL.SetString("22259814647920560318926", 10)
		trustedNodeRPL.SetString("1171569191995818964092", 10)
		nodeRPL.SetString("54673228959804884990101", 10)
		nodeETH.SetString("71988611161220176226", 10)
		executionBlock = big.NewInt(21234057)
		consensusBlock = big.NewInt(10445279)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x82e89c1b2cfa0248ee5d2ff9bd4b0013388bc0da2005ee03e666c83c7bb51a92")
		merkleTreeCID = "bafybeigp23ay6zk3rsrqtzukbsire5fyak5q3tke2nyjip2irkk2imy24m"
		intervalStartTime = time.Unix(1729748139, 0)
		intervalEndTime = time.Unix(1732167339, 0)
		submissionTime = time.Unix(1732168487, 0)
		userETH.SetString("114014988744595356456", 10)
	case 30:
		treasuryRPL.SetString("22343284949727202292854", 10)
		trustedNodeRPL.SetString("1175962365775115910096", 10)
		nodeRPL.SetString("54878243736172075803559", 10)
		nodeETH.SetString("69245099392448778466", 10)
		executionBlock = big.NewInt(21434520)
		consensusBlock = big.NewInt(10646879)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x021b434c4f6bc565225d03d77ba722f41668f5805574f8825c49d4f68b4e89b9")
		merkleTreeCID = "bafybeiemh5r7e7imqdxvmt4oc6pyktv6kqiwwyyqtps3wozf374ztvzpqi"
		intervalStartTime = time.Unix(1732167339, 0)
		intervalEndTime = time.Unix(1734586539, 0)
		submissionTime = time.Unix(1734594971, 0)
		userETH.SetString("93870760781122417496", 10)
	case 31:
		treasuryRPL.SetString("22427068250154627605694", 10)
		trustedNodeRPL.SetString("1180372013166033031824", 10)
		nodeRPL.SetString("55084027281081541484240", 10)
		nodeETH.SetString("60850105918353222556", 10)
		executionBlock = big.NewInt(21634996)
		consensusBlock = big.NewInt(10848479)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x6502a89b162c41b1cbc019aaa82c6c1e76f872bfca0cd7dfd688bf217b9a820d")
		merkleTreeCID = "bafybeigbvc6vtvsu3azzzjn3tzb4lhnt63xfkpzpxjgzt6y2fgsjwptesu"
		intervalStartTime = time.Unix(1734586539, 0)
		intervalEndTime = time.Unix(1737005739, 0)
		submissionTime = time.Unix(1737017999, 0)
		userETH.SetString("78963743041980491460", 10)
	case 32:
		treasuryRPL.SetString("22511165722891375675992", 10)
		trustedNodeRPL.SetString("1184798195941651351312", 10)
		nodeRPL.SetString("55290582477277063060540", 10)
		nodeETH.SetString("61524968493175262980", 10)
		executionBlock = big.NewInt(21835515)
		consensusBlock = big.NewInt(11050079)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x44941bda178c03ca6d8a5e72fd9433fd566ac801d57cb5ccb8d833436bc07f95")
		merkleTreeCID = "bafybeicimfw4h73vpgwnukkkro5iecgrxomqvvdr66ecy54ulihi3bnixy"
		intervalStartTime = time.Unix(1737005739, 0)
		intervalEndTime = time.Unix(1739424939, 0)
		submissionTime = time.Unix(1739436551, 0)
		userETH.SetString("78337779811500941545", 10)
	case 33:
		treasuryRPL.SetString("22595578546027106873451", 10)
		trustedNodeRPL.SetString("1189240976106689835392", 10)
		nodeRPL.SetString("55497912218312192317529", 10)
		nodeETH.SetString("133016473781938819503", 10)
		executionBlock = big.NewInt(22035900)
		consensusBlock = big.NewInt(11251679)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x6c71cf2beb6bf80c90a0bed50b2eff376919bf3016ac5102ce403a4bd74c35b3")
		merkleTreeCID = "bafybeiavd6deyed4b2t73g7wthhr3ozsw7s5x2ektglhwepbd44mi2hm2i"
		intervalStartTime = time.Unix(1739424939, 0)
		intervalEndTime = time.Unix(1741844139, 0)
		submissionTime = time.Unix(1741856879, 0)
		userETH.SetString("192141107613389949606", 10)
	case 34:
		treasuryRPL.SetString("21884507624803523266610", 10)
		trustedNodeRPL.SetString("1989500693163956660512", 10)
		nodeRPL.SetString("55706019408590786493514", 10)
		nodeETH.SetString("51299429061085367685", 10)
		executionBlock = big.NewInt(22236473)
		consensusBlock = big.NewInt(11453279)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x0311e81c20f7161dd37db58cec5108bc687ba7b4352ccbc797a04028b08fd8a1")
		merkleTreeCID = "bafybeidmkskyk2po4dzyak3t5ve6stujyd3diyolim4hwsnejrv5cv2b3a"
		intervalStartTime = time.Unix(1741844139, 0)
		intervalEndTime = time.Unix(1744263339, 0)
		submissionTime = time.Unix(1744283327, 0)
		userETH.SetString("60649284015156972974", 10)
	case 35:
		treasuryRPL.SetString("21966570592767308741140", 10)
		trustedNodeRPL.SetString("1996960962978846249104", 10)
		nodeRPL.SetString("55914906963407694974189", 10)
		nodeETH.SetString("38282673601408514244", 10)
		executionBlock = big.NewInt(22436785)
		consensusBlock = big.NewInt(11654879)
		intervalsPassed = big.NewInt(1)
		merkleRoot = common.HexToHash("0x057eb4b982774871cd1aefd2962589620fee1e9cfcbf36d0f08e292da837dc49")
		merkleTreeCID = "bafybeid4ttvmr547lz57opya7sjhlorpi3egdo7qccvokfx5w72ziul6uu"
		intervalStartTime = time.Unix(1744263339, 0)
		intervalEndTime = time.Unix(1746682539, 0)
		submissionTime = time.Unix(1746691523, 0)
		userETH.SetString("38934754761348367759", 10)
	default:
		return RewardsEvent{}, false, nil
	}

	eventDataInterval := RewardsEvent{
		Index:             big.NewInt(int64(index)),
		ExecutionBlock:    executionBlock,
		ConsensusBlock:    consensusBlock,
		IntervalsPassed:   intervalsPassed,
		TreasuryRPL:       treasuryRPL,
		TrustedNodeRPL:    []*big.Int{trustedNodeRPL},
		NodeRPL:           []*big.Int{nodeRPL},
		NodeETH:           []*big.Int{nodeETH},
		UserETH:           userETH,
		MerkleRoot:        merkleRoot,
		MerkleTreeCID:     merkleTreeCID,
		IntervalStartTime: intervalStartTime,
		IntervalEndTime:   intervalEndTime,
		SubmissionTime:    submissionTime,
	}

	return eventDataInterval, true, nil
}

// Get contracts
var rocketRewardsPoolLock sync.Mutex

func getRocketRewardsPool(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketRewardsPoolLock.Lock()
	defer rocketRewardsPoolLock.Unlock()
	return rp.GetContract("rocketRewardsPool", opts)
}
