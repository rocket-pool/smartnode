package service

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	dataFolder string = "/.rocketpool/data"
)

// ===============
// === Factory ===
// ===============

type serviceTerminateDataFolderContextFactory struct {
	handler *ServiceHandler
}

func (f *serviceTerminateDataFolderContextFactory) Create(args url.Values) (*serviceTerminateDataFolderContext, error) {
	c := &serviceTerminateDataFolderContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *serviceTerminateDataFolderContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*serviceTerminateDataFolderContext, api.ServiceTerminateDataFolderData](
		router, "terminate-data-folder", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type serviceTerminateDataFolderContext struct {
	handler *ServiceHandler
}

// Deletes the contents of the data folder including the wallet file, password file, and all validator keys.
// Don't use this unless you have a very good reason to do it (such as switching from a Testnet to Mainnet).
func (c *serviceTerminateDataFolderContext) PrepareData(data *api.ServiceTerminateDataFolderData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Check if it exists
	_, err := os.Stat(dataFolder)
	if os.IsNotExist(err) {
		data.FolderExisted = false
		return types.ResponseStatus_Success, nil
	}
	data.FolderExisted = true

	// Traverse it
	files, err := os.ReadDir(dataFolder)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error enumerating files: %w", err)
	}

	// Delete the children
	for _, file := range files {
		// Skip the validators directory - that get special treatment
		if file.Name() != config.ValidatorsFolderName && !file.IsDir() {
			fullPath := filepath.Join(dataFolder, file.Name())
			if file.IsDir() {
				err = os.RemoveAll(fullPath)
			} else {
				err = os.Remove(fullPath)
			}
			if err != nil {
				return types.ResponseStatus_Error, fmt.Errorf("error removing [%s]: %w", file.Name(), err)
			}
		}
	}

	// Traverse the validators dir
	validatorsDir := filepath.Join(dataFolder, config.ValidatorsFolderName)
	files, err = os.ReadDir(validatorsDir)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error enumerating validator files: %w", err)
	}

	// Delete the children
	for _, file := range files {
		fullPath := filepath.Join(validatorsDir, file.Name())
		if file.IsDir() {
			err = os.RemoveAll(fullPath)
		} else {
			err = os.Remove(fullPath)
		}
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error removing [%s]: %w", file.Name(), err)
		}
	}

	return types.ResponseStatus_Success, nil
}
