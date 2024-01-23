package service

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
		router, "terminate-data-folder", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type serviceTerminateDataFolderContext struct {
	handler *ServiceHandler
}

// Deletes the contents of the data folder including the wallet file, password file, and all validator keys.
// Don't use this unless you have a very good reason to do it (such as switching from Prater to Mainnet).
func (c *serviceTerminateDataFolderContext) PrepareData(data *api.ServiceTerminateDataFolderData, opts *bind.TransactOpts) error {
	// Check if it exists
	_, err := os.Stat(dataFolder)
	if os.IsNotExist(err) {
		data.FolderExisted = false
		return nil
	}
	data.FolderExisted = true

	// Traverse it
	files, err := os.ReadDir(dataFolder)
	if err != nil {
		return fmt.Errorf("error enumerating files: %w", err)
	}

	// Delete the children
	for _, file := range files {
		// Skip the validators directory - that get special treatment
		if file.Name() != "validators" && !file.IsDir() {
			fullPath := filepath.Join(dataFolder, file.Name())
			if file.IsDir() {
				err = os.RemoveAll(fullPath)
			} else {
				err = os.Remove(fullPath)
			}
			if err != nil {
				return fmt.Errorf("error removing [%s]: %w", file.Name(), err)
			}
		}
	}

	// Traverse the validators dir
	validatorsDir := filepath.Join(dataFolder, "validators")
	files, err = os.ReadDir(validatorsDir)
	if err != nil {
		return fmt.Errorf("error enumerating validator files: %w", err)
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
			return fmt.Errorf("error removing [%s]: %w", file.Name(), err)
		}
	}

	return nil
}
