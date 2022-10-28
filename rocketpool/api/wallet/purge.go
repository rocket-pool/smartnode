package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	"github.com/urfave/cli"
)

func purge(c *cli.Context) (*api.PurgeResponse, error) {

	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	pm, err := services.GetPasswordManager(c)
	if err != nil {
		return nil, err
	}

	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	d, err := services.GetDocker(c)
	if err != nil {
		return nil, err
	}

	response := api.PurgeResponse{}

	// Stop the VC to unlock keystores and slashing DBs
	err = validator.StopValidator(cfg, bc, nil, d)
	if err != nil {
		return nil, fmt.Errorf("error stopping validator client: %w", err)
	}

	// Delete the VC directories
	err = w.DeleteValidatorStores()
	if err != nil {
		return nil, fmt.Errorf("error deleting validator storage: %w", err)
	}

	// Delete the wallet and password
	err = w.Delete()
	if err != nil {
		return nil, fmt.Errorf("error deleting wallet: %w", err)
	}
	err = pm.DeletePassword()
	if err != nil {
		return nil, fmt.Errorf("error deleting password: %w", err)
	}

	// Restart the VC once cleanup is done
	err = validator.RestartValidator(cfg, bc, nil, d)
	if err != nil {
		return nil, fmt.Errorf("error restarting validator client: %w", err)
	}

	return &response, nil
}
