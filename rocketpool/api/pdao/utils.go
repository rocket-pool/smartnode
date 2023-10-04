package pdao

import (
	"encoding/base64"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/klauspost/compress/zstd"
	"github.com/rocket-pool/rocketpool-go/dao/protocol/voting"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
)

// Constructs a pollard for the latest finalized block
func createPollard(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (uint32, []types.VotingTreeNode, string, error) {
	// Get the latest finalized block
	mgr, err := state.NewNetworkStateManager(rp, cfg, rp.Client, bc, nil)
	if err != nil {
		return 0, nil, "", fmt.Errorf("error creating state manager: %w", err)
	}
	block, err := mgr.GetLatestFinalizedBeaconBlock()
	if err != nil {
		return 0, nil, "", fmt.Errorf("error determining latest finalized block: %w", err)
	}
	blockNumber := uint32(block.ExecutionBlockNumber)

	// Create the proposal tree
	gen, err := voting.NewVotingTreeGenerator(rp)
	if err != nil {
		return 0, nil, "", fmt.Errorf("error creating voting power tree generator: %w", err)
	}

	// Create the voting power pollard for the proposal
	pollard, err := gen.CreatePollardRowForProposal(blockNumber, nil)
	if err != nil {
		return 0, nil, "", fmt.Errorf("error creating voting power pollard: %w", err)
	}

	// Serialize it to JSON
	pollardBytes, err := json.Marshal(pollard)
	if err != nil {
		return 0, nil, "", fmt.Errorf("error serializing pollard: %w", err)
	}

	// Compress it
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return 0, nil, "", fmt.Errorf("error creating compressor: %w", err)
	}
	compressedBytes := encoder.EncodeAll(pollardBytes, make([]byte, 0, len(pollardBytes)))
	encodedPollard := base64.StdEncoding.EncodeToString(compressedBytes)

	// Return it all
	return blockNumber, pollard, encodedPollard, nil
}

// Decodes a serialized pollard
func decodePollard(pollard string) ([]types.VotingTreeNode, error) {
	// Decompress the pollard
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating decompressor: %w", err)
	}
	compressedBytes, err := base64.StdEncoding.DecodeString(pollard)
	if err != nil {
		return nil, fmt.Errorf("error decoding pollard: %w", err)
	}
	decompressedBytes, err := decoder.DecodeAll(compressedBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("error decompressing pollard: %w", err)
	}

	// Decompress it from JSON
	var truePollard []types.VotingTreeNode
	err = json.Unmarshal(decompressedBytes, &truePollard)
	if err != nil {
		return nil, fmt.Errorf("error deserializing pollard: %w", err)
	}

	return truePollard, nil
}
