package rocketpool

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// This manages a primary and optional fallback Execution client
type EcManager struct {
	PrimaryEc     *ethclient.Client
	PrimaryEcUrl  string
	FallbackEc    *ethclient.Client
	FallbackEcUrl string
}

// Creates a new execution client manager. If not using a fallback client, use "" for the fallbackEcUrl.
func NewEcManager(cfg *config.RocketPoolConfig) (*EcManager, error) {

	var primaryEcUrl string
	var fallbackEcUrl string

	// Get the primary EC url
	if cfg.ExecutionClientMode.Value.(config.Mode) == config.Mode_Local {
		primaryEcUrl = fmt.Sprintf("http://%s:%d", config.Eth1ContainerName, cfg.ExecutionCommon.HttpPort.Value)
	} else {
		primaryEcUrl = cfg.ExternalExecution.HttpUrl.Value.(string)
	}

	// Get the fallback EC url, if applicable
	if cfg.UseFallbackExecutionClient.Value == true {
		if cfg.FallbackExecutionClientMode.Value.(config.Mode) == config.Mode_Local {
			fallbackEcUrl = fmt.Sprintf("http://%s:%d", config.Eth1FallbackContainerName, cfg.FallbackExecutionCommon.HttpPort.Value)
		} else {
			fallbackEcUrl = cfg.FallbackExternalExecution.HttpUrl.Value.(string)
		}
	}

	primaryEc, err := ethclient.Dial(primaryEcUrl)
	if err != nil {
		return nil, fmt.Errorf("error connecting to primary EC at [%s]: %w", primaryEcUrl, err)
	}

	var fallbackEc *ethclient.Client
	if fallbackEcUrl != "" {
		fallbackEc, err = ethclient.Dial(fallbackEcUrl)
		if err != nil {
			return nil, fmt.Errorf("error connecting to fallback EC at [%s]: %w", fallbackEcUrl, err)
		}
	}

	manager := &EcManager{
		PrimaryEc:     primaryEc,
		PrimaryEcUrl:  primaryEcUrl,
		FallbackEc:    fallbackEc,
		FallbackEcUrl: fallbackEcUrl,
	}

	return manager, nil
}

// Gets the EC URL to use for a CLI function's sequence of calls, printing warnings where appropriate
func (m *EcManager) GetWorkingEcUrl() (string, error) {

	// Check the primary's sync progress
	primaryProgress, err := m.PrimaryEc.SyncProgress(context.Background())
	if err != nil {
		fmt.Printf("%sWARNING: Primary EC's sync progress check failed with [%s], using fallback EC...%s\n", colorYellow, err.Error(), colorReset)

		err = m.testFallbackEc()
		if err != nil {
			return "", err
		}
		return m.FallbackEcUrl, nil
	}

	if primaryProgress == nil {
		// Make sure it's up to date
		isUpToDate, blockTime, err := services.IsSyncWithinThreshold(m.PrimaryEc)
		if err != nil {
			fmt.Printf("%sWARNING: Error checking if primary EC's sync progress is up to date: [%s], using fallback EC...%s\n", colorYellow, err.Error(), colorReset)

			err = m.testFallbackEc()
			if err != nil {
				return "", err
			}
			return m.FallbackEcUrl, nil
		}
		if !isUpToDate {
			fmt.Printf("%sWARNING: Primary EC claims to have finished syncing, but its last block was from %s ago. It likely doesn't have enough peers. Using fallback EC...%s\n", colorYellow, time.Since(blockTime), err)

			err = m.testFallbackEc()
			if err != nil {
				return "", err
			}
			return m.FallbackEcUrl, nil
		}

		// Primary is synced and up to date!
		return m.PrimaryEcUrl, nil

	} else {
		// It's not synced yet, print the progress
		p := float64(primaryProgress.CurrentBlock-primaryProgress.StartingBlock) / float64(primaryProgress.HighestBlock-primaryProgress.StartingBlock)
		if p > 1 {
			fmt.Printf("%sNOTE: Primary EC is still syncing, using fallback EC...%s\n", colorYellow, colorReset)
			err = m.testFallbackEc()
			if err != nil {
				return "", err
			}
			return m.FallbackEcUrl, nil
		} else {
			fmt.Printf("%sNOTE: Primary EC is still syncing (%.2f%%), using fallback EC...%s\n", colorYellow, p*100, colorReset)
			err = m.testFallbackEc()
			if err != nil {
				return "", err
			}
			return m.FallbackEcUrl, nil
		}
	}

}

// Test the Fallback EC
func (m *EcManager) testFallbackEc() error {

	// Make sure there's a fallback configured
	if m.FallbackEc == nil {
		fmt.Printf("%sNo fallback EC configured.\n%s", colorYellow, colorReset)
		return fmt.Errorf("all execution clients failed")
	}

	// Get the fallback's sync progress
	fallbackProgress, err := m.FallbackEc.SyncProgress(context.Background())
	if err != nil {
		fmt.Printf("%sWARNING: Fallback EC's sync progress check failed with [%s].%s\n", colorRed, err.Error(), colorReset)
		return fmt.Errorf("all execution clients failed")
	}

	// Make sure it's up to date
	if fallbackProgress == nil {

		isUpToDate, blockTime, err := services.IsSyncWithinThreshold(m.FallbackEc)
		if err != nil {
			fmt.Printf("%sWARNING: Error checking if fallback EC's sync progress is up to date: [%s].%s\n", colorRed, err.Error(), colorReset)
			return fmt.Errorf("all execution clients failed")
		}
		if !isUpToDate {
			fmt.Printf("%sWARNING: Fallback EC claims to have finished syncing, but its last block was from %s ago. It likely doesn't have enough peers.%s\n", colorYellow, time.Since(blockTime), err)
			return fmt.Errorf("all execution clients failed")
		}
		// It's synced and it works!
		return nil

	} else {
		// It's not synced yet, print the progress
		p := float64(fallbackProgress.CurrentBlock-fallbackProgress.StartingBlock) / float64(fallbackProgress.HighestBlock-fallbackProgress.StartingBlock)
		if p > 1 {
			fmt.Printf("%sFallback EC is still syncing.%s\n", colorYellow, colorReset)
			return fmt.Errorf("all execution clients failed")
		} else {
			fmt.Printf("%sFallback EC is still syncing: %.2f%%%s\n", colorYellow, p*100, colorReset)
			return fmt.Errorf("all execution clients failed")
		}
	}

}
