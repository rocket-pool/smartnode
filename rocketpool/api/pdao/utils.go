package pdao

import (
	"encoding/base64"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/klauspost/compress/zstd"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/proposals"
)

// Constructs a pollard for the latest finalized block
func createPollard(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, bc beacon.Client) (uint32, []types.VotingTreeNode, string, error) {
	mgr, err := proposals.NewProposalTreeManager(nil, cfg, rp, bc)
	if err != nil {
		return 0, nil, "", fmt.Errorf("error creating proposal tree manager: %w", err)
	}

	return mgr.CreateArtifactsForProposal()
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
