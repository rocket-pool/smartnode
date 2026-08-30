package config

// ClientTagSet selects production vs test Docker tags for execution/consensus clients.
// Image tag *values* stay in Go (they change with Smart Node releases); this flag is per-network.
type ClientTagSet string

const (
	ClientTagSetProduction ClientTagSet = "production"
	ClientTagSetTest       ClientTagSet = "test"
)

// NetworkInfo is a selectable Ethereum network loaded from YAML.
type NetworkInfo struct {
	Name        string `yaml:"name"`
	Label       string `yaml:"label"`
	Description string `yaml:"description"`
	Default     bool   `yaml:"default"`
	ChainID     uint   `yaml:"chainID"`

	TxWatchUrl          string `yaml:"txWatchUrl"`
	NodeManagerUrl      string `yaml:"nodeManagerUrl"`
	BeaconExplorerUrl   string `yaml:"beaconExplorerUrl"`
	SnapshotApiDomain   string `yaml:"snapshotApiDomain"`
	FlashbotsProtectUrl string `yaml:"flashbotsProtectUrl"`
	FlashbotsRelayUrl   string `yaml:"flashbotsRelayUrl"`

	BeaconNetwork        string       `yaml:"beaconNetwork"`
	CommitBoostChainName string       `yaml:"commitBoostChainName"`
	ClientTagSet         ClientTagSet `yaml:"clientTagSet"`
	CustomChainConfigDir string       `yaml:"customChainConfigDir"`

	SupportsMevBoost                 bool `yaml:"supportsMevBoost"`
	AllowNonBlsWithdrawalCredentials bool `yaml:"allowNonBlsWithdrawalCredentials"`
	IsProduction                     bool `yaml:"isProduction"`

	NethermindPruneThresholdMb uint64 `yaml:"nethermindPruneThresholdMb"`

	Rewards   NetworkRewards    `yaml:"rewards"`
	MevRelays map[string]string `yaml:"mevRelays"`
	Addresses NetworkAddresses  `yaml:"addresses"`
}

func (n *NetworkInfo) ID() Network {
	if n == nil {
		return Network_Unknown
	}
	return Network(n.Name)
}

func (n *NetworkInfo) RelayURL(id MevRelayID) string {
	if n == nil || n.MevRelays == nil {
		return ""
	}
	return n.MevRelays[string(id)]
}

// NetworkRewards holds per-network rewards-tree ruleset cutovers.
type NetworkRewards struct {
	RulesetStartIntervals        []RulesetStartInterval `yaml:"rulesetStartIntervals"`
	ConsensusBonusCutoffInterval uint64                 `yaml:"consensusBonusCutoffInterval"`
}

type RulesetStartInterval struct {
	Version       uint64 `yaml:"version"`
	StartInterval uint64 `yaml:"startInterval"`
}

func (r NetworkRewards) StartInterval(version uint64) uint64 {
	for _, s := range r.RulesetStartIntervals {
		if s.Version == version {
			return s.StartInterval
		}
	}
	return 0
}

// NetworkAddresses is the Rocket Pool / helper contract set for a network.
type NetworkAddresses struct {
	Storage                  string   `yaml:"storage"`
	RocketSignerRegistry     string   `yaml:"rocketSignerRegistry"`
	RplToken                 string   `yaml:"rplToken"`
	Reth                     string   `yaml:"reth"`
	V1_0_0_RewardsPool       string   `yaml:"v1_0_0_RewardsPool"`
	V1_0_0_ClaimNode         string   `yaml:"v1_0_0_ClaimNode"`
	V1_0_0_ClaimTrustedNode  string   `yaml:"v1_0_0_ClaimTrustedNode"`
	V1_0_0_MinipoolManager   string   `yaml:"v1_0_0_MinipoolManager"`
	V1_1_0_NetworkPrices     string   `yaml:"v1_1_0_NetworkPrices"`
	V1_1_0_NodeStaking       string   `yaml:"v1_1_0_NodeStaking"`
	V1_1_0_NodeDeposit       string   `yaml:"v1_1_0_NodeDeposit"`
	V1_1_0_MinipoolQueue     string   `yaml:"v1_1_0_MinipoolQueue"`
	V1_1_0_MinipoolFactory   string   `yaml:"v1_1_0_MinipoolFactory"`
	V1_2_0_NetworkPrices     string   `yaml:"v1_2_0_NetworkPrices"`
	V1_2_0_NetworkBalances   string   `yaml:"v1_2_0_NetworkBalances"`
	PreviousRewardsPools     []string `yaml:"previousRewardsPools"`
	PreviousDAOVerifiers     []string `yaml:"previousRocketDAOProtocolVerifier"`
	OptimismPriceMessenger   string   `yaml:"optimismPriceMessenger"`
	PolygonPriceMessenger    string   `yaml:"polygonPriceMessenger"`
	ArbitrumPriceMessenger   string   `yaml:"arbitrumPriceMessenger"`
	ArbitrumPriceMessengerV2 string   `yaml:"arbitrumPriceMessengerV2"`
	ZkSyncEraPriceMessenger  string   `yaml:"zkSyncEraPriceMessenger"`
	BasePriceMessenger       string   `yaml:"basePriceMessenger"`
	ScrollPriceMessenger     string   `yaml:"scrollPriceMessenger"`
	ScrollFeeEstimator       string   `yaml:"scrollFeeEstimator"`
	RplTwapPool              string   `yaml:"rplTwapPool"`
	Multicall                string   `yaml:"multicall"`
	BalanceBatcher           string   `yaml:"balancebatcher"`
}
