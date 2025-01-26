package debug

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)

func getBeaconStateForSlot(c *cli.Context, slot uint64, validatorIndex uint64) error {
	// Create a new response
	response := api.BeaconStateResponse{}

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return err
	}

	// Get beacon state
	beaconState, err := bc.GetBeaconState(slot)
	if err != nil {
		return err
	}

	proof, err := beaconState.ValidatorProof(validatorIndex)
	if err != nil {
		return err
	}

	// Convert the proof to a list of 0x-prefixed hex strings
	response.Proof = make([]string, 0, len(proof))
	for _, hash := range proof {
		response.Proof = append(response.Proof, hexutil.EncodeToString(hash))
	}

	// Render response json
	json, err := json.Marshal(response)
	if err != nil {
		return err
	}
	fmt.Println(string(json))

	return nil
}
