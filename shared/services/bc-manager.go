package services

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/beacon/client"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

const bnContainerName string = "eth2"

// This is a proxy for multiple Beacon clients, providing natural fallback support if one of them fails.
type BeaconClientManager struct {
	primaryBc       beacon.Client
	fallbackBc      beacon.Client
	logger          log.ColorLogger
	primaryReady    bool
	fallbackReady   bool
	ignoreSyncCheck bool
}

// This is a signature for a wrapped Beacon client function that only returns an error
type bcFunction0 func(beacon.Client) error

// This is a signature for a wrapped Beacon client function that returns 1 var and an error
type bcFunction1 func(beacon.Client) (interface{}, error)

// This is a signature for a wrapped Beacon client function that returns 2 vars and an error
type bcFunction2 func(beacon.Client) (interface{}, interface{}, error)

// Creates a new BeaconClientManager instance based on the Rocket Pool config
func NewBeaconClientManager(cfg *config.RocketPoolConfig) (*BeaconClientManager, error) {

	// Primary CC
	var primaryProvider string
	var selectedCC cfgtypes.ConsensusClient
	if cfg.IsNativeMode {
		primaryProvider = cfg.Native.CcHttpUrl.Value.(string)
		selectedCC = cfg.Native.ConsensusClient.Value.(cfgtypes.ConsensusClient)
	} else if cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		primaryProvider = fmt.Sprintf("http://%s:%d", bnContainerName, cfg.ConsensusCommon.ApiPort.Value.(uint16))
		selectedCC = cfg.ConsensusClient.Value.(cfgtypes.ConsensusClient)
	} else if cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_External {
		selectedConsensusConfig, err := cfg.GetSelectedConsensusClientConfig()
		if err != nil {
			return nil, err
		}
		primaryProvider = selectedConsensusConfig.(cfgtypes.ExternalConsensusConfig).GetApiUrl()
		selectedCC = cfg.ExternalConsensusClient.Value.(cfgtypes.ConsensusClient)
	} else {
		return nil, fmt.Errorf("Unknown Consensus client mode '%v'", cfg.ConsensusClientMode.Value)
	}

	// Fallback CC
	var fallbackProvider string
	if cfg.UseFallbackClients.Value == true {
		if cfg.IsNativeMode {
			fallbackProvider = cfg.FallbackNormal.CcHttpUrl.Value.(string)
		} else {
			switch selectedCC {
			case cfgtypes.ConsensusClient_Prysm:
				fallbackProvider = cfg.FallbackPrysm.CcHttpUrl.Value.(string)
			default:
				fallbackProvider = cfg.FallbackNormal.CcHttpUrl.Value.(string)
			}
		}
	}

	var primaryBc beacon.Client
	var fallbackBc beacon.Client
	primaryBc = client.NewStandardHttpClient(primaryProvider)
	if fallbackProvider != "" {
		fallbackBc = client.NewStandardHttpClient(fallbackProvider)
	}

	return &BeaconClientManager{
		primaryBc:     primaryBc,
		fallbackBc:    fallbackBc,
		logger:        log.NewColorLogger(color.FgHiBlue),
		primaryReady:  true,
		fallbackReady: fallbackBc != nil,
	}, nil

}

/// ======================
/// BeaconClient Functions
/// ======================

// Get the client's process mode
func (m *BeaconClientManager) GetClientType() (beacon.BeaconClientType, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetClientType()
	})
	if err != nil {
		return beacon.Unknown, err
	}
	return result.(beacon.BeaconClientType), nil
}

// Get the client's sync status
func (m *BeaconClientManager) GetSyncStatus() (beacon.SyncStatus, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetSyncStatus()
	})
	if err != nil {
		return beacon.SyncStatus{}, err
	}
	return result.(beacon.SyncStatus), nil
}

// Get the Beacon configuration
func (m *BeaconClientManager) GetEth2Config() (beacon.Eth2Config, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetEth2Config()
	})
	if err != nil {
		return beacon.Eth2Config{}, err
	}
	return result.(beacon.Eth2Config), nil
}

// Get the Beacon configuration
func (m *BeaconClientManager) GetEth2DepositContract() (beacon.Eth2DepositContract, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetEth2DepositContract()
	})
	if err != nil {
		return beacon.Eth2DepositContract{}, err
	}
	return result.(beacon.Eth2DepositContract), nil
}

// Get the attestations in a Beacon chain block
func (m *BeaconClientManager) GetAttestations(blockId string) ([]beacon.AttestationInfo, bool, error) {
	result1, result2, err := m.runFunction2(func(client beacon.Client) (interface{}, interface{}, error) {
		return client.GetAttestations(blockId)
	})
	if err != nil {
		return nil, false, err
	}
	return result1.([]beacon.AttestationInfo), result2.(bool), nil
}

// Get a Beacon chain block
func (m *BeaconClientManager) GetBeaconBlock(blockId string) (beacon.BeaconBlock, bool, error) {
	result1, result2, err := m.runFunction2(func(client beacon.Client) (interface{}, interface{}, error) {
		return client.GetBeaconBlock(blockId)
	})
	if err != nil {
		return beacon.BeaconBlock{}, false, err
	}
	return result1.(beacon.BeaconBlock), result2.(bool), nil
}

// Get the Beacon chain's head information
func (m *BeaconClientManager) GetBeaconHead() (beacon.BeaconHead, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetBeaconHead()
	})
	if err != nil {
		return beacon.BeaconHead{}, err
	}
	return result.(beacon.BeaconHead), nil
}

// Get a validator's status by its index
func (m *BeaconClientManager) GetValidatorStatusByIndex(index string, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetValidatorStatusByIndex(index, opts)
	})
	if err != nil {
		return beacon.ValidatorStatus{}, err
	}
	return result.(beacon.ValidatorStatus), nil
}

// Get a validator's status by its pubkey
func (m *BeaconClientManager) GetValidatorStatus(pubkey types.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetValidatorStatus(pubkey, opts)
	})
	if err != nil {
		return beacon.ValidatorStatus{}, err
	}
	return result.(beacon.ValidatorStatus), nil
}

// Get the statuses of multiple validators by their pubkeys
func (m *BeaconClientManager) GetValidatorStatuses(pubkeys []types.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (map[types.ValidatorPubkey]beacon.ValidatorStatus, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetValidatorStatuses(pubkeys, opts)
	})
	if err != nil {
		return nil, err
	}
	return result.(map[types.ValidatorPubkey]beacon.ValidatorStatus), nil
}

// Get a validator's index
func (m *BeaconClientManager) GetValidatorIndex(pubkey types.ValidatorPubkey) (string, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetValidatorIndex(pubkey)
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// Get a validator's sync duties
func (m *BeaconClientManager) GetValidatorSyncDuties(indices []string, epoch uint64) (map[string]bool, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetValidatorSyncDuties(indices, epoch)
	})
	if err != nil {
		return nil, err
	}
	return result.(map[string]bool), nil
}

// Get a validator's proposer duties
func (m *BeaconClientManager) GetValidatorProposerDuties(indices []string, epoch uint64) (map[string]uint64, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetValidatorProposerDuties(indices, epoch)
	})
	if err != nil {
		return nil, err
	}
	return result.(map[string]uint64), nil
}

// Get the Beacon chain's domain data
func (m *BeaconClientManager) GetDomainData(domainType []byte, epoch uint64, useGenesisFork bool) ([]byte, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetDomainData(domainType, epoch, useGenesisFork)
	})
	if err != nil {
		return nil, err
	}
	return result.([]byte), nil
}

// Voluntarily exit a validator
func (m *BeaconClientManager) ExitValidator(validatorIndex string, epoch uint64, signature types.ValidatorSignature) error {
	err := m.runFunction0(func(client beacon.Client) error {
		return client.ExitValidator(validatorIndex, epoch, signature)
	})
	return err
}

// Close the connection to the Beacon client
func (m *BeaconClientManager) Close() error {
	err := m.runFunction0(func(client beacon.Client) error {
		return client.Close()
	})
	return err
}

// Get the EL data for a CL block
func (m *BeaconClientManager) GetEth1DataForEth2Block(blockId string) (beacon.Eth1Data, bool, error) {
	result1, result2, err := m.runFunction2(func(client beacon.Client) (interface{}, interface{}, error) {
		return client.GetEth1DataForEth2Block(blockId)
	})
	if err != nil {
		return beacon.Eth1Data{}, false, err
	}
	return result1.(beacon.Eth1Data), result2.(bool), nil
}

// Get the attestation committees for an epoch
func (m *BeaconClientManager) GetCommitteesForEpoch(epoch *uint64) (beacon.Committees, error) {
	result, err := m.runFunction1(func(client beacon.Client) (interface{}, error) {
		return client.GetCommitteesForEpoch(epoch)
	})
	if err != nil {
		return nil, err
	}
	return result.(beacon.Committees), nil
}

// Change the withdrawal credentials for a validator
func (m *BeaconClientManager) ChangeWithdrawalCredentials(validatorIndex string, fromBlsPubkey types.ValidatorPubkey, toExecutionAddress common.Address, signature types.ValidatorSignature) error {
	err := m.runFunction0(func(client beacon.Client) error {
		return client.ChangeWithdrawalCredentials(validatorIndex, fromBlsPubkey, toExecutionAddress, signature)
	})
	if err != nil {
		return err
	}
	return nil
}

/// ==================
/// Internal Functions
/// ==================

func (m *BeaconClientManager) CheckStatus() *api.ClientManagerStatus {

	status := &api.ClientManagerStatus{
		FallbackEnabled: m.fallbackBc != nil,
	}

	// Ignore the sync check and just use the predefined settings if requested
	if m.ignoreSyncCheck {
		status.PrimaryClientStatus.IsWorking = m.primaryReady
		status.PrimaryClientStatus.IsSynced = m.primaryReady
		if status.FallbackEnabled {
			status.FallbackClientStatus.IsWorking = m.fallbackReady
			status.FallbackClientStatus.IsSynced = m.fallbackReady
		}
		return status
	}

	// Get the primary BC status
	status.PrimaryClientStatus = checkBcStatus(m.primaryBc)

	// Get the fallback BC status if applicable
	if status.FallbackEnabled {
		status.FallbackClientStatus = checkBcStatus(m.fallbackBc)
	}

	// Flag the ready clients
	m.primaryReady = (status.PrimaryClientStatus.IsWorking && status.PrimaryClientStatus.IsSynced)
	m.fallbackReady = (status.FallbackEnabled && status.FallbackClientStatus.IsWorking && status.FallbackClientStatus.IsSynced)

	return status

}

// Check the client status
func checkBcStatus(client beacon.Client) api.ClientStatus {

	status := api.ClientStatus{}

	// Get the fallback's sync progress
	syncStatus, err := client.GetSyncStatus()
	if err != nil {
		status.Error = fmt.Sprintf("Sync progress check failed with [%s]", err.Error())
		status.IsSynced = false
		status.IsWorking = false
		return status
	}

	// Return the sync status
	if !syncStatus.Syncing {
		status.IsWorking = true
		status.IsSynced = true
		status.SyncProgress = 1
	} else {
		status.IsWorking = true
		status.IsSynced = false
		status.SyncProgress = syncStatus.Progress
	}
	return status

}

// Attempts to run a function progressively through each client until one succeeds or they all fail.
func (m *BeaconClientManager) runFunction0(function bcFunction0) error {

	// Check if we can use the primary
	if m.primaryReady {
		// Try to run the function on the primary
		err := function(m.primaryBc)
		if err != nil {
			if m.isDisconnected(err) {
				// If it's disconnected, log it and try the fallback
				m.logger.Printlnf("WARNING: Primary Beacon client disconnected (%s), using fallback...", err.Error())
				m.primaryReady = false
				return m.runFunction0(function)
			}
			// If it's a different error, just return it
			return err
		}
		// If there's no error, return the result
		return nil
	}

	if m.fallbackReady {
		// Try to run the function on the fallback
		err := function(m.fallbackBc)
		if err != nil {
			if m.isDisconnected(err) {
				// If it's disconnected, log it and try the fallback
				m.logger.Printlnf("WARNING: Fallback Beacon client disconnected (%s)", err.Error())
				m.fallbackReady = false
				return fmt.Errorf("all Beacon clients failed")
			}

			// If it's a different error, just return it
			return err
		}
		// If there's no error, return the result
		return nil
	}

	return fmt.Errorf("no Beacon clients were ready")
}

// Attempts to run a function progressively through each client until one succeeds or they all fail.
func (m *BeaconClientManager) runFunction1(function bcFunction1) (interface{}, error) {

	// Check if we can use the primary
	if m.primaryReady {
		// Try to run the function on the primary
		result, err := function(m.primaryBc)
		if err != nil {
			if m.isDisconnected(err) {
				// If it's disconnected, log it and try the fallback
				m.logger.Printlnf("WARNING: Primary Beacon client disconnected (%s), using fallback...", err.Error())
				m.primaryReady = false
				return m.runFunction1(function)
			}
			// If it's a different error, just return it
			return nil, err
		}
		// If there's no error, return the result
		return result, nil
	}

	if m.fallbackReady {
		// Try to run the function on the fallback
		result, err := function(m.fallbackBc)
		if err != nil {
			if m.isDisconnected(err) {
				// If it's disconnected, log it and try the fallback
				m.logger.Printlnf("WARNING: Fallback Beacon client disconnected (%s)", err.Error())
				m.fallbackReady = false
				return nil, fmt.Errorf("all Beacon clients failed")
			}
			// If it's a different error, just return it
			return nil, err
		}
		// If there's no error, return the result
		return result, nil
	}

	return nil, fmt.Errorf("no Beacon clients were ready")

}

// Attempts to run a function progressively through each client until one succeeds or they all fail.
func (m *BeaconClientManager) runFunction2(function bcFunction2) (interface{}, interface{}, error) {

	// Check if we can use the primary
	if m.primaryReady {
		// Try to run the function on the primary
		result1, result2, err := function(m.primaryBc)
		if err != nil {
			if m.isDisconnected(err) {
				// If it's disconnected, log it and try the fallback
				m.logger.Printlnf("WARNING: Primary Beacon client disconnected (%s), using fallback...", err.Error())
				m.primaryReady = false
				return m.runFunction2(function)
			}
			// If it's a different error, just return it
			return nil, nil, err
		}
		// If there's no error, return the result
		return result1, result2, nil
	}

	if m.fallbackReady {
		// Try to run the function on the fallback
		result1, result2, err := function(m.fallbackBc)
		if err != nil {
			if m.isDisconnected(err) {
				// If it's disconnected, log it and try the fallback
				m.logger.Printlnf("WARNING: Fallback Beacon client disconnected (%s)", err.Error())
				m.fallbackReady = false
				return nil, nil, fmt.Errorf("all Beacon clients failed")
			}
			// If it's a different error, just return it
			return nil, nil, err
		}
		// If there's no error, return the result
		return result1, result2, nil
	}

	return nil, nil, fmt.Errorf("no Beacon clients were ready")

}

// Returns true if the error was a connection failure and a backup client is available
func (m *BeaconClientManager) isDisconnected(err error) bool {
	return strings.Contains(err.Error(), "dial tcp")
}
