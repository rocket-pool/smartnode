package config

import "time"

const (
	// Watchtower
	WatchtowerMaxFeeDefault      uint64 = 200
	WatchtowerPriorityFeeDefault uint64 = 3

	// Daemon
	EventLogInterval        int    = 1000
	SmartNodeDaemonRoute    string = "smartnode"
	SmartNodeSocketFilename string = SmartNodeDaemonRoute + ".sock"
	ConfigFilename          string = "user-settings.yml"

	// Wallet
	UserAddressFilename       string = "address"
	UserWalletDataFilename    string = "wallet"
	UserNextAccountFilename   string = "next_account"
	UserPasswordFilename      string = "password"
	ValidatorsFolderName      string = "validators"
	CustomKeysFolderName      string = "custom-keys"
	CustomKeyPasswordFilename string = "custom-key-passwords"

	// Scripts
	EcStartScript       string = "start-ec.sh"
	BnStartScript       string = "start-bn.sh"
	VcStartScript       string = "start-vc.sh"
	MevBoostStartScript string = "start-mev-boost.sh"

	// HTTP
	ClientTimeout time.Duration = 8 * time.Second

	// Volumes
	ExecutionClientDataVolume string = "eth1clientdata"
	BeaconNodeDataVolume      string = "eth2clientdata"

	// Smart Node
	AddonsFolderName                   string = "addons"
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
	FeeRecipientFilename               string = "rp-fee-recipient.txt"
	MinipoolPerformanceFilenameFormat  string = "rp-minipool-performance-%s-%d.json"

	// Snapshot
	SnapshotID string = "rocketpool-dao.eth"

	// Utility Containers
	PruneProvisionerTag string = "rocketpool/eth1-prune-provision:v0.0.1"
	EcMigratorTag       string = "rocketpool/ec-migrator:v1.0.0"
)
