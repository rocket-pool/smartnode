package rewards

import (
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/yaml.v2"
)

const (
	networkMainnet = "mainnet"
	networkTestnet = "testnet"

	rewardsEventLocalFilenameFormat  = "rp-rewards-event-%s-%d.yaml"
	rewardsEventRemoteFilenameFormat = "rp-rewards-event-%d.yaml"
)

var (
	rewardsEventFileURLFormat    = "https://github.com/rocket-pool/rewards-trees/raw/main/%s/%s"
	rewardsEventDownloadTimeouts = []time.Duration{200 * time.Millisecond, 2 * time.Second, 60 * time.Second}
)

var rewardsEventFileLock sync.Mutex

// On-disk / GitHub encoding of a RewardSnapshot event.
type rewardsEventFile struct {
	Index             string   `yaml:"index"`
	ExecutionBlock    string   `yaml:"executionBlock"`
	ConsensusBlock    string   `yaml:"consensusBlock"`
	MerkleRoot        string   `yaml:"merkleRoot"`
	IntervalsPassed   string   `yaml:"intervalsPassed"`
	TreasuryRPL       string   `yaml:"treasuryRPL"`
	TrustedNodeRPL    []string `yaml:"trustedNodeRPL"`
	NodeRPL           []string `yaml:"nodeRPL"`
	NodeETH           []string `yaml:"nodeETH"`
	UserETH           string   `yaml:"userETH"`
	IntervalStartTime int64    `yaml:"intervalStartTime"`
	IntervalEndTime   int64    `yaml:"intervalEndTime"`
	SubmissionTime    int64    `yaml:"submissionTime"`
}

func supportsRewardsEventFile(network string) bool {
	return network == networkMainnet || network == networkTestnet
}

func rewardsEventLocalPath(cacheDir, network string, index uint64) string {
	return filepath.Join(cacheDir, fmt.Sprintf(rewardsEventLocalFilenameFormat, network, index))
}

func rewardsEventRemoteURL(network string, index uint64) string {
	return fmt.Sprintf(rewardsEventFileURLFormat, network, fmt.Sprintf(rewardsEventRemoteFilenameFormat, index))
}

// Load a rewards event from the local cache or GitHub. A miss (ok=false, err=nil)
// means the caller should fall through to the execution-client log query.
func getIntervalRewardsEvent(index uint64, network, cacheDir string) (RewardsEvent, bool, error) {
	if !supportsRewardsEventFile(network) {
		return RewardsEvent{}, false, nil
	}

	rewardsEventFileLock.Lock()
	defer rewardsEventFileLock.Unlock()

	var localPath string
	if cacheDir != "" {
		localPath = rewardsEventLocalPath(cacheDir, network, index)
		if event, ok, err := readRewardsEventFile(localPath, index); err == nil && ok {
			return event, true, nil
		}
	}

	event, raw, ok, err := downloadRewardsEventFile(network, index)
	if err != nil {
		return RewardsEvent{}, false, err
	}
	if !ok {
		return RewardsEvent{}, false, nil
	}

	if localPath != "" {
		if writeErr := writeRewardsEventFile(localPath, raw); writeErr != nil {
			return RewardsEvent{}, false, fmt.Errorf("error caching rewards event file %s: %w", localPath, writeErr)
		}
	}
	return event, true, nil
}

func readRewardsEventFile(path string, expectedIndex uint64) (RewardsEvent, bool, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return RewardsEvent{}, false, nil
		}
		return RewardsEvent{}, false, err
	}
	event, err := parseRewardsEventFile(bytes, expectedIndex)
	if err != nil {
		return RewardsEvent{}, false, err
	}
	return event, true, nil
}

func writeRewardsEventFile(path string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

func downloadRewardsEventFile(network string, index uint64) (RewardsEvent, []byte, bool, error) {
	url := rewardsEventRemoteURL(network, index)
	for _, timeout := range rewardsEventDownloadTimeouts {
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			continue
		}
		if resp.StatusCode == http.StatusNotFound {
			return RewardsEvent{}, nil, false, nil
		}
		if resp.StatusCode != http.StatusOK {
			continue
		}
		event, err := parseRewardsEventFile(body, index)
		if err != nil {
			return RewardsEvent{}, nil, false, fmt.Errorf("error parsing rewards event file from %s: %w", url, err)
		}
		return event, body, true, nil
	}
	// Treat persistent download failures as a miss so the execution-client path can run.
	return RewardsEvent{}, nil, false, nil
}

func parseRewardsEventFile(raw []byte, expectedIndex uint64) (RewardsEvent, error) {
	var file rewardsEventFile
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return RewardsEvent{}, fmt.Errorf("error unmarshaling rewards event yaml: %w", err)
	}

	index, err := parseBigInt(file.Index)
	if err != nil {
		return RewardsEvent{}, fmt.Errorf("invalid index: %w", err)
	}
	if index.Uint64() != expectedIndex {
		return RewardsEvent{}, fmt.Errorf("rewards event file index %s does not match requested interval %d", file.Index, expectedIndex)
	}

	executionBlock, err := parseBigInt(file.ExecutionBlock)
	if err != nil {
		return RewardsEvent{}, fmt.Errorf("invalid executionBlock: %w", err)
	}
	consensusBlock, err := parseBigInt(file.ConsensusBlock)
	if err != nil {
		return RewardsEvent{}, fmt.Errorf("invalid consensusBlock: %w", err)
	}
	intervalsPassed, err := parseBigInt(file.IntervalsPassed)
	if err != nil {
		return RewardsEvent{}, fmt.Errorf("invalid intervalsPassed: %w", err)
	}
	treasuryRPL, err := parseBigInt(file.TreasuryRPL)
	if err != nil {
		return RewardsEvent{}, fmt.Errorf("invalid treasuryRPL: %w", err)
	}
	userETH, err := parseBigInt(file.UserETH)
	if err != nil {
		return RewardsEvent{}, fmt.Errorf("invalid userETH: %w", err)
	}
	trustedNodeRPL, err := parseBigIntSlice(file.TrustedNodeRPL)
	if err != nil {
		return RewardsEvent{}, fmt.Errorf("invalid trustedNodeRPL: %w", err)
	}
	nodeRPL, err := parseBigIntSlice(file.NodeRPL)
	if err != nil {
		return RewardsEvent{}, fmt.Errorf("invalid nodeRPL: %w", err)
	}
	nodeETH, err := parseBigIntSlice(file.NodeETH)
	if err != nil {
		return RewardsEvent{}, fmt.Errorf("invalid nodeETH: %w", err)
	}

	return RewardsEvent{
		Index:             index,
		ExecutionBlock:    executionBlock,
		ConsensusBlock:    consensusBlock,
		IntervalsPassed:   intervalsPassed,
		TreasuryRPL:       treasuryRPL,
		TrustedNodeRPL:    trustedNodeRPL,
		NodeRPL:           nodeRPL,
		NodeETH:           nodeETH,
		UserETH:           userETH,
		MerkleRoot:        common.HexToHash(file.MerkleRoot),
		IntervalStartTime: time.Unix(file.IntervalStartTime, 0),
		IntervalEndTime:   time.Unix(file.IntervalEndTime, 0),
		SubmissionTime:    time.Unix(file.SubmissionTime, 0),
	}, nil
}

func parseBigInt(s string) (*big.Int, error) {
	if s == "" {
		return big.NewInt(0), nil
	}
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("%q is not a valid integer", s)
	}
	return v, nil
}

func parseBigIntSlice(values []string) ([]*big.Int, error) {
	out := make([]*big.Int, len(values))
	for i, s := range values {
		v, err := parseBigInt(s)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}
