package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/alerting"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/utils"
)

// Settings
const (
	EthClientSyncTimeout    int64 = 8 // 8 seconds
	BeaconClientSyncTimeout int64 = 8 // 8 seconds

	checkRocketStorageInterval  time.Duration = time.Second * 15
	checkNodeRegisteredInterval time.Duration = time.Second * 15
	checkNodeWalletInterval     time.Duration = time.Second * 15
)

// ====================
// === Requirements ===
// ====================

func (sp *ServiceProvider) RequireEthClientSynced2() error {
	// TODO: wrap the NMC one for alerting
	cfg := sp.GetConfig()
	alerting.AlertExecutionClientSyncComplete(cfg)
	// Intentially broken for now, won't compile until this is fixed
}

func (sp *ServiceProvider) RequireNodeAddress() error {
	status, err := sp.GetWallet().GetStatus()
	if err != nil {
		return err
	}
	if !status.Address.HasAddress {
		return errors.New("The node currently does not have an address set. Please run 'rocketpool wallet init' and try again.")
	}
	return nil
}

func (sp *ServiceProvider) RequireWalletReady() error {
	status, err := sp.GetWallet().GetStatus()
	if err != nil {
		return err
	}
	return utils.CheckIfWalletReady(status)
}

func (sp *ServiceProvider) RequireNodeRegistered(context context.Context) error {
	if err := sp.RequireNodeAddress(); err != nil {
		return err
	}
	if err := sp.ServiceProvider.RequireEthClientSynced(context); err != nil {
		return err
	}
	nodeRegistered, err := sp.getNodeRegistered()
	if err != nil {
		return err
	}
	if !nodeRegistered {
		return errors.New("The node is not registered with Rocket Pool. Please run 'rocketpool node register' and try again.")
	}
	return nil
}

func (sp *ServiceProvider) RequireRplFaucet() error {
	if sp.rplFaucet == nil {
		network := string(sp.cfg.Network.Value)
		return fmt.Errorf("The RPL faucet is not available on the %s network.", network)
	}
	return nil
}

func (sp *ServiceProvider) RequireSnapshot() error {
	if sp.snapshotDelegation == nil {
		network := string(sp.cfg.Network.Value)
		return fmt.Errorf("Snapshot voting is not available on the %s network.", network)
	}
	return nil
}

func (sp *ServiceProvider) RequireOnOracleDao(context context.Context) error {
	if err := sp.RequireNodeAddress(); err != nil {
		return err
	}
	if err := sp.ServiceProvider.RequireEthClientSynced(context); err != nil {
		return err
	}
	nodeTrusted, err := sp.isMemberOfOracleDao()
	if err != nil {
		return err
	}
	if !nodeTrusted {
		return errors.New("The node is not a member of the oracle DAO. Nodes can only join the oracle DAO by invite.")
	}
	return nil
}

func (sp *ServiceProvider) RequireOnSecurityCouncil(context context.Context) error {
	if err := sp.RequireNodeAddress(); err != nil {
		return err
	}
	if err := sp.ServiceProvider.RequireEthClientSynced(context); err != nil {
		return err
	}
	nodeTrusted, err := sp.isMemberOfSecurityCouncil()
	if err != nil {
		return err
	}
	if !nodeTrusted {
		return errors.New("The node is not a member of the security council. Nodes can only join the security council by invite.")
	}
	return nil
}

// ===============================
// === Service Synchronization ===
// ===============================

func (sp *ServiceProvider) WaitWalletReady(verbose bool) error {
	for {
		status, err := sp.GetWallet().GetStatus()
		if err != nil {
			return err
		}
		var message string

		if !status.Address.HasAddress {
			message = "The node currently does not have an address set"
		} else if !status.Wallet.IsLoaded {
			if status.Wallet.IsOnDisk {
				if !status.Password.IsPasswordSaved {
					message = "The node has a node wallet on disk but does not have the password for it loaded"
				} else {
					message = "The node has a node wallet and a password on disk but there was an error loading it - perhaps the password is incorrect? Please check the daemon logs for more information"
				}
			} else {
				message = "The node currently does not have a node wallet keystore"
			}
		} else if status.Wallet.WalletAddress != status.Address.NodeAddress {
			message = "The node's wallet keystore does not match the node address. This node is currently in read-only mode."
		} else {
			return nil
		}

		if verbose {
			log.Printf("%s, retrying in %s...\n", message, checkNodeWalletInterval.String())
		}
		time.Sleep(checkNodeWalletInterval)
	}
}

// Wait until the node has been registered with the Rocket Pool network
func (sp *ServiceProvider) WaitNodeRegistered(context context.Context, verbose bool) error {
	if err := sp.WaitWalletReady(verbose); err != nil {
		return err
	}
	if err := sp.WaitEthClientSynced(context, verbose); err != nil {
		return err
	}
	if err := sp.LoadContractsIfStale(); err != nil {
		return fmt.Errorf("error loading contract bindings: %w", err)
	}
	for {
		nodeRegistered, err := sp.getNodeRegistered()
		if err != nil {
			return err
		}
		if nodeRegistered {
			return nil
		}
		if verbose {
			log.Printf("The node is not registered with Rocket Pool, retrying in %s...\n", checkNodeRegisteredInterval.String())
		}
		time.Sleep(checkNodeRegisteredInterval)
	}
}

// ===============
// === Helpers ===
// ===============

// Check if the node is registered
func (sp *ServiceProvider) getNodeRegistered() (bool, error) {
	rp := sp.rocketPool
	address, _ := sp.GetWallet().GetAddress()

	// Create a node binding
	node, err := node.NewNode(rp, address)
	if err != nil {
		return false, fmt.Errorf("error creating node binding: %w", err)
	}

	// Get contract state
	err = rp.Query(nil, nil, node.Exists)
	if err != nil {
		return false, fmt.Errorf("error getting node registration status: %w", err)
	}
	return node.Exists.Get(), nil
}

// Check if the node is a member of the oracle DAO
func (sp *ServiceProvider) isMemberOfOracleDao() (bool, error) {
	rp := sp.rocketPool
	address, _ := sp.GetWallet().GetAddress()

	// Create the bindings
	odaoMember, err := oracle.NewOracleDaoMember(rp, address)
	if err != nil {
		return false, fmt.Errorf("error creating oDAO member binding: %w", err)
	}

	// Get contract state
	err = rp.Query(nil, nil, odaoMember.Exists)
	if err != nil {
		return false, fmt.Errorf("error getting oDAO member status: %w", err)
	}
	return odaoMember.Exists.Get(), nil
}

// Check if the node is a member of the security council
func (sp *ServiceProvider) isMemberOfSecurityCouncil() (bool, error) {
	rp := sp.rocketPool
	address, _ := sp.GetWallet().GetAddress()

	// Create the bindings
	scMember, err := security.NewSecurityCouncilMember(rp, address)
	if err != nil {
		return false, fmt.Errorf("error creating security council member binding: %w", err)
	}

	// Get contract state
	err = rp.Query(nil, nil, scMember.Exists)
	if err != nil {
		return false, fmt.Errorf("error getting security council member status: %w", err)
	}
	return scMember.Exists.Get(), nil
}
