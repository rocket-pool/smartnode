package config

import "time"

const (
	// Watchtower
	WatchtowerMaxFeeDefault      uint64 = 200
	WatchtowerPriorityFeeDefault uint64 = 3

	// Daemon
	EventLogInterval         int    = 1000
	SmartNodeDaemonRoute     string = "smartnode"
	HyperdriveSocketFilename string = SmartNodeDaemonRoute + ".sock"
	ConfigFilename           string = "user-settings.yml"

	// Wallet
	UserAddressFilename    string = "address"
	UserWalletDataFilename string = "wallet"
	UserPasswordFilename   string = "password"

	// Scripts
	EcStartScript string = "start-ec.sh"
	BnStartScript string = "start-bn.sh"
	VcStartScript string = "start-vc.sh"

	// HTTP
	ClientTimeout time.Duration = 8 * time.Second

	// Volumes
	ExecutionClientDataVolume string = "eth1clientdata"
	BeaconNodeDataVolume      string = "eth2clientdata"

	// Smart Node
	ChecksumTableFilename              string = "checksums.sha384"
	RewardsTreeIpfsExtension           string = ".zst"
	RewardsTreeFilenameFormat          string = "rp-rewards-%s-%d.json"
	VotingFolder                       string = "voting"
	RecordsFolder                      string = "records"
	RewardsTreesFolder                 string = "rewards-trees"
	PrimaryRewardsFileUrl              string = "https://%s.ipfs.dweb.link/%s"
	SecondaryRewardsFileUrl            string = "https://ipfs.io/ipfs/%s/%s"
	GithubRewardsFileUrl               string = "https://github.com/rocket-pool/rewards-trees/raw/main/%s/%s"
	WatchtowerFolder                   string = "watchtower"
	RegenerateRewardsTreeRequestSuffix string = ".request"
	RegenerateRewardsTreeRequestFormat string = "%d" + RegenerateRewardsTreeRequestSuffix

	// Snapshot
	SnapshotID string = "rocketpool-dao.eth"
)
