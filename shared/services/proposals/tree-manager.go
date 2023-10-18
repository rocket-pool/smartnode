package proposals

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/klauspost/compress/zstd"
	"github.com/rocket-pool/rocketpool-go/dao/protocol/voting"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

const (
	snapshotFilenameFormat        string = "vi-%s-%d.json.zst"
	snapshotFilenamePattern       string = ".*\\-(?P<block>\\d+)\\.json\\.zst"
	latestCompatibleVersionString string = "1.12.0-dev"
)

// Snapshot of the entire network's voting power and delegation status at a specific block
type NetworkVotingInfoSnapshot struct {
	SmartnodeVersion string                 `json:"smartnodeVersion"`
	Network          string                 `json:"network"`
	BlockNumber      uint32                 `json:"blockNumber"`
	VotingInfo       []types.NodeVotingInfo `json:"votingInfo"`
}

// Struct for managing proposal trees
type ProposalTreeManager struct {
	log       *log.ColorLogger
	logPrefix string
	cfg       *config.RocketPoolConfig
	rp        *rocketpool.RocketPool
	bc        beacon.Client

	gen                   *voting.VotingTreeGenerator
	compressor            *zstd.Encoder
	decompressor          *zstd.Decoder
	snapshotFilenameRegex *regexp.Regexp
}

// Create a new ProposalTreeManager instance
func NewProposalTreeManager(log *log.ColorLogger, cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, bc beacon.Client) (*ProposalTreeManager, error) {
	// Create the proposal tree generator
	gen, err := voting.NewVotingTreeGenerator(rp, common.HexToAddress(cfg.Smartnode.GetMulticallAddress()))
	if err != nil {
		return nil, fmt.Errorf("error creating voting power tree generator: %w", err)
	}

	// Create the zstd compressor and decompressor
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return nil, fmt.Errorf("error creating zstd compressor for proposal tree manager: %w", err)
	}
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating zstd decompressor for proposal tree manager: %w", err)
	}

	// Create the snapshot filename regex
	logPrefix := "[Proposal Tree]"
	snapshotFilenameRegex := regexp.MustCompile(snapshotFilenamePattern)

	manager := &ProposalTreeManager{
		log:                   log,
		logPrefix:             logPrefix,
		cfg:                   cfg,
		rp:                    rp,
		bc:                    bc,
		gen:                   gen,
		compressor:            encoder,
		decompressor:          decoder,
		snapshotFilenameRegex: snapshotFilenameRegex,
	}

	return manager, nil
}

// Construct a snapshot of network voting power for the given slot, and save the snapshot for it to disk
func (m *ProposalTreeManager) CreateSnapshotForBlock(blockNumber uint32) (*NetworkVotingInfoSnapshot, error) {
	nodeVotingInfo, err := m.gen.GetNodeVotingInfo(blockNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting node voting info for block %d: %w", blockNumber, err)
	}
	snapshot := &NetworkVotingInfoSnapshot{
		SmartnodeVersion: shared.RocketPoolVersion,
		Network:          string(m.cfg.Smartnode.Network.Value.(cfgtypes.Network)),
		BlockNumber:      blockNumber,
		VotingInfo:       nodeVotingInfo,
	}
	return snapshot, nil
}

// Construct the artifacts necessary for submitting a Protocol DAO proposal
func (m *ProposalTreeManager) CreateArtifactsForProposal(snapshot *NetworkVotingInfoSnapshot) (uint32, []types.VotingTreeNode, string, error) {
	// Create the voting power pollard for the proposal
	pollard := m.gen.CreatePollardRowForProposal(snapshot.VotingInfo)

	// Serialize it to JSON
	pollardBytes, err := json.Marshal(pollard)
	if err != nil {
		return 0, nil, "", fmt.Errorf("error serializing pollard: %w", err)
	}

	// Compress it
	compressedBytes := m.compressor.EncodeAll(pollardBytes, make([]byte, 0, len(pollardBytes)))
	encodedPollard := base64.StdEncoding.EncodeToString(compressedBytes)

	// Return it all
	return snapshot.BlockNumber, pollard, encodedPollard, nil
}

// Save a network voting snapshot to a file and update the snapshot info catalog
func (m *ProposalTreeManager) SaveSnapshotToFile(snapshot *NetworkVotingInfoSnapshot) error {
	// Serialize the snapshot
	bytes, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("error saving voting info snapshot: %w", err)
	}

	// Compress the snapshot
	compressedBytes := m.compressor.EncodeAll(bytes, make([]byte, 0, len(bytes)))

	// Get the snapshot filename
	votingPath := m.cfg.Smartnode.GetVotingPath()
	filename := filepath.Join(votingPath, fmt.Sprintf(snapshotFilenameFormat, snapshot.Network, snapshot.BlockNumber))

	// Write it to a file
	err = os.WriteFile(filename, compressedBytes, 0664)
	if err != nil {
		return fmt.Errorf("error writing file [%s]: %w", filename, err)
	}

	// Compute the SHA384 hash to act as a checksum
	checksum := sha512.Sum384(compressedBytes)

	// Load the existing checksum table
	_, lines, err := m.parseChecksumFile()
	if err != nil {
		return fmt.Errorf("error parsing checkpoint file: %w", err)
	}
	if lines == nil {
		lines = []string{}
	}

	// Add the new snapshot checksum
	baseFilename := filepath.Base(filename)
	checksumLine := fmt.Sprintf("%s  %s", hex.EncodeToString(checksum[:]), baseFilename)

	// Sort the lines by their slot
	err = m.sortChecksumEntries(lines)
	if err != nil {
		return fmt.Errorf("error sorting checkpoint file entries: %w", err)
	}

	overwritten := false
	for i, line := range lines {
		if strings.HasSuffix(line, baseFilename) {
			// If there is already a line with the filename, overwrite it
			lines[i] = checksumLine
			overwritten = true
			break
		}
	}
	if !overwritten {
		// If there's no existing lines, add this to the end
		lines = append(lines, checksumLine)
	}

	fileContents := strings.Join(lines, "\n")
	checksumBytes := []byte(fileContents)

	// Save the new file
	checksumFilename := filepath.Join(votingPath, config.ChecksumTableFilename)
	err = os.WriteFile(checksumFilename, checksumBytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing checksum file: %w", err)
	}

	return nil
}

// Load the snapshot for the provided block from disk if it exists, using the checksum table as an index
func (m *ProposalTreeManager) LoadSnapshotFromDisk(blockNumber uint32) (*NetworkVotingInfoSnapshot, error) {
	latestCompatibleVersion, err := semver.New(latestCompatibleVersionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing latest compatible version string [%s]: %w", latestCompatibleVersionString, err)
	}

	// Parse the checksum file
	exists, lines, err := m.parseChecksumFile()
	if err != nil {
		return nil, fmt.Errorf("error parsing checkpoint file: %w", err)
	}
	if !exists {
		// There isn't a checksum file so start over
		m.logMessage("%s Checksum file not found, cannot load any previously saved snapshots.", m.logPrefix)
		return nil, nil
	}

	// Iterate over each file, counting backwards from the bottom
	snapshotPath := m.cfg.Smartnode.GetVotingPath()
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]

		// Extract the checksum, filename, and block number
		checksumString, filename, candidateBlock, err := m.parseChecksumEntry(line)
		if err != nil {
			return nil, err
		}

		// Check if the slot was too far into the future
		if candidateBlock != uint64(blockNumber) {
			continue
		}

		// Make sure the checksum parses properly
		checksum, err := hex.DecodeString(checksumString)
		if err != nil {
			return nil, fmt.Errorf("error scanning checkpoint line (%s): checksum (%s) could not be parsed", line, checksumString)
		}

		// Try to load it
		fullFilename := filepath.Join(snapshotPath, filename)
		snapshot, err := m.loadSnapshotFromFile(fullFilename, checksum)
		if err != nil {
			m.logMessage("%s WARNING: error loading snapshot from file [%s]: %s... cannot use the saved snapshot.", m.logPrefix, fullFilename, err.Error())
			return nil, nil
		}

		// Check if it has the proper network
		if snapshot.Network != string(m.cfg.Smartnode.Network.Value.(cfgtypes.Network)) {
			m.logMessage("%s File [%s] is for network %s instead of %s so it cannot be used.", m.logPrefix, filename, snapshot.Network, string(m.cfg.Smartnode.Network.Value.(cfgtypes.Network)))
			return nil, nil
		}

		// Check if it's using a compatible version
		snapshotVersion, err := semver.New(snapshot.SmartnodeVersion)
		if err != nil {
			m.logMessage("%s Failed to parse the version info for file [%s] so it cannot be used.", m.logPrefix, filename)
			return nil, nil
		}
		if snapshotVersion.LT(*latestCompatibleVersion) {
			m.logMessage("%s File [%s] was made with Smartnode v%s which is not compatible (lowest compatible = v%s) so it cannot be used.", m.logPrefix, filename, snapshot.SmartnodeVersion, latestCompatibleVersionString)
			return nil, nil
		}

		m.logMessage("%s Loaded file [%s].", m.logPrefix, filename)
		return snapshot, nil
	}

	// If we got here then none of the saved files worked so we have to make a new snapshot
	m.logMessage("%s A snapshot for block %d could not be found.", m.logPrefix, blockNumber)
	return nil, nil
}

// Load a snapshot from a file, making sure its contents match the provided checksum
func (m *ProposalTreeManager) loadSnapshotFromFile(filename string, expectedChecksum []byte) (*NetworkVotingInfoSnapshot, error) {
	// Read the file
	compressedBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Calculate the hash and validate it
	checksum := sha512.Sum384(compressedBytes)
	if !bytes.Equal(expectedChecksum, checksum[:]) {
		expectedString := hex.EncodeToString(expectedChecksum)
		actualString := hex.EncodeToString(checksum[:])
		return nil, fmt.Errorf("checksum mismatch (expected %s, but it was %s)", expectedString, actualString)
	}

	// Decompress it
	bytes, err := m.decompressor.DecodeAll(compressedBytes, []byte{})
	if err != nil {
		return nil, fmt.Errorf("error decompressing data: %w", err)
	}

	// Deserialize it
	var snapshot NetworkVotingInfoSnapshot
	err = json.Unmarshal(bytes, &snapshot)
	if err != nil {
		return nil, fmt.Errorf("error deserializing network voting info snapshot: %w", err)
	}

	return &snapshot, nil
}

// Get the lines from the checksum file
func (m *ProposalTreeManager) parseChecksumFile() (bool, []string, error) {
	// Get the checksum filename
	votingPath := m.cfg.Smartnode.GetVotingPath()
	checksumFilename := filepath.Join(votingPath, config.ChecksumTableFilename)

	// Check if the file exists
	_, err := os.Stat(checksumFilename)
	if os.IsNotExist(err) {
		return false, nil, nil
	}

	// Open the checksum file
	checksumTable, err := os.ReadFile(checksumFilename)
	if err != nil {
		return false, nil, fmt.Errorf("error loading checksum table (%s): %w", checksumFilename, err)
	}

	// Parse out each line
	originalLines := strings.Split(string(checksumTable), "\n")

	// Remove empty lines
	lines := make([]string, 0, len(originalLines))
	for _, line := range originalLines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			lines = append(lines, line)
		}
	}

	return true, lines, nil
}

// Sort the checksum file entries by their block number
func (m *ProposalTreeManager) sortChecksumEntries(lines []string) error {
	var sortErr error
	sort.Slice(lines, func(i int, j int) bool {
		_, _, firstBlock, err := m.parseChecksumEntry(lines[i])
		if err != nil && sortErr == nil {
			sortErr = err
			return false
		}

		_, _, secondBlock, err := m.parseChecksumEntry(lines[j])
		if err != nil && sortErr == nil {
			sortErr = err
			return false
		}

		return firstBlock < secondBlock
	})
	return sortErr
}

// Get the checksum, the filename, and the block number from a checksum entry.
func (m *ProposalTreeManager) parseChecksumEntry(line string) (string, string, uint64, error) {
	// Extract the checksum and filename
	elems := strings.Split(line, "  ")
	if len(elems) != 2 {
		return "", "", 0, fmt.Errorf("error parsing checkpoint line (%s): expected 2 elements, but got %d", line, len(elems))
	}
	checksumString := elems[0]
	filename := elems[1]

	// Extract the block number for this file
	blockNumber, err := m.getBlockNumberFromFilename(filename)
	if err != nil {
		return "", "", 0, fmt.Errorf("error scanning checkpoint line (%s): %w", line, err)
	}

	return checksumString, filename, blockNumber, nil
}

// Get the block number from a snapshot filename
func (m *ProposalTreeManager) getBlockNumberFromFilename(filename string) (uint64, error) {
	matches := m.snapshotFilenameRegex.FindStringSubmatch(filename)
	if matches == nil {
		return 0, fmt.Errorf("filename (%s) did not match the expected format", filename)
	}
	blockIndex := m.snapshotFilenameRegex.SubexpIndex("block")
	if blockIndex == -1 {
		return 0, fmt.Errorf("block number not found in filename (%s)", filename)
	}
	blockString := matches[blockIndex]
	blockNumber, err := strconv.ParseUint(blockString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("block number (%s) could not be parsed to a number", blockString)
	}

	return blockNumber, nil
}

// Log a message to the logger
func (m *ProposalTreeManager) logMessage(message string, args ...any) {
	if m.log != nil {
		m.log.Printlnf(message, args)
	}
}
