package node_state

import (
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/rocket-pool/smartnode/shared/services/config"
	"gopkg.in/yaml.v2"
)

const (
	version uint64 = 1
)

type Reward struct {
	Date   time.Time `yaml:"date"`
	Amount *big.Int  `yaml:"amount"`
}

type RewardsRecord struct {
	Total   *big.Int `yaml:"total,omitempty"`
	Rewards []Reward `yaml:"rewards,omitempty"`
}

type NodeState struct {
	Version              uint64         `yaml:"version,omitempty"`
	NextBlockToCheck     uint64         `yaml:"lastBlockChecked,omitempty"`
	RplRewards           *RewardsRecord `yaml:"rplRewards,omitempty"`
	SmoothingPoolRewards *RewardsRecord `yaml:"smoothingPoolRewards,omitempty"`
	BeaconRewards        *RewardsRecord `yaml:"beaconRewards,omitempty"`
}

// Creates a new NodeState record
func NewNodeState() *NodeState {
	state := &NodeState{
		Version:          version,
		NextBlockToCheck: 0,
		RplRewards: &RewardsRecord{
			Total:   big.NewInt(0),
			Rewards: []Reward{},
		},
		SmoothingPoolRewards: &RewardsRecord{
			Total:   big.NewInt(0),
			Rewards: []Reward{},
		},
		BeaconRewards: &RewardsRecord{
			Total:   big.NewInt(0),
			Rewards: []Reward{},
		},
	}

	return state
}

// Load the latest saved state from disk
func LoadState(cfg *config.RocketPoolConfig) (*NodeState, bool, error) {
	// Get the config path
	path := cfg.Smartnode.GetNodeStatePath(false)

	// Check if the file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("error checking node state file: %w", err)
	}

	// Load it
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, false, fmt.Errorf("error loading node state file: %w", err)
	}

	// Deserialize it
	var state NodeState
	err = yaml.Unmarshal(bytes, &state)
	if err != nil {
		return nil, false, fmt.Errorf("error deserializing node state file: %w", err)
	}

	return &state, true, nil
}

// Save the node state to disk
func (s *NodeState) Save(cfg *config.RocketPoolConfig) error {
	// Serialize it
	bytes, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("error serializing node state file: %w", err)
	}

	// Save it
	path := cfg.Smartnode.GetNodeStatePath(false)
	err = os.WriteFile(path, bytes, 0664)
	if err != nil {
		return fmt.Errorf("error saving node state file: %w", err)
	}

	return nil
}

// Adds a rewards record to the reward container
func (r *RewardsRecord) AddReward(amount *big.Int, date time.Time) {
	r.Total.Add(r.Total, amount)
	r.Rewards = append(r.Rewards, Reward{
		Date:   date,
		Amount: amount,
	})
}
