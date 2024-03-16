package config

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"

	"github.com/pbnjay/memory"
	"github.com/rocket-pool/node-manager-core/config"
)

const localBnStepID string = "step-local-bn"

func createLocalCcStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {
	// Create the button names and descriptions from the config
	clientNames := []string{"Random (Recommended)"}
	clientDescriptions := []string{"Select a client randomly to help promote the diversity of the Beacon Chain. We recommend you do this unless you have a strong reason to pick a specific client. To learn more about why client diversity is important, please visit https://clientdiversity.org for an explanation."}
	clients := []*config.ParameterOption[config.BeaconNode]{}
	for _, client := range wiz.md.Config.LocalBeaconClient.BeaconNode.Options {
		clientNames = append(clientNames, client.Name)
		clientDescriptions = append(clientDescriptions, getAugmentedBnDescription(client.Value, client.Description))
		clients = append(clients, client)
	}

	helperText := "Please select the Consensus client you would like to use.\n\nHighlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth2-clients to learn more about them."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		if !wiz.md.isNew {
			var bnName string
			for _, option := range wiz.md.Config.LocalBeaconClient.BeaconNode.Options {
				if option.Value == wiz.md.Config.LocalBeaconClient.BeaconNode.Value {
					bnName = option.Name
					break
				}
			}
			for i, clientName := range clientNames {
				if bnName == clientName {
					modal.focus(i)
					break
				}
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.pages.RemovePage(randomBnPrysmID)
			wiz.md.pages.RemovePage(randomBnID)
			selectRandomBn(clients, true, wiz, currentStep, totalSteps)
		} else {
			buttonLabel = strings.TrimSpace(buttonLabel)
			selectedClient := config.BeaconNode_Unknown
			for _, client := range wiz.md.Config.LocalBeaconClient.BeaconNode.Options {
				if client.Name == buttonLabel {
					selectedClient = client.Value
					break
				}
			}
			if selectedClient == config.BeaconNode_Unknown {
				panic(fmt.Sprintf("Local BN selection buttons didn't match any known clients, buttonLabel = %s\n", buttonLabel))
			}
			wiz.md.Config.LocalBeaconClient.BeaconNode.Value = selectedClient
			switch selectedClient {
			//case config.ConsensusClient_Prysm:
			//	wiz.consensusLocalPrysmWarning.show()
			case config.BeaconNode_Teku:
				totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
				if runtime.GOARCH == "arm64" || totalMemoryGB < 15 {
					wiz.localBnTekuWarning.show()
				} else {
					wiz.checkpointSyncProviderModal.show()
				}
			default:
				wiz.checkpointSyncProviderModal.show()
			}
		}
	}

	back := func() {
		wiz.localEcModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		clientNames,
		clientDescriptions,
		100,
		"Beacon Node > Selection",
		DirectionalModalVertical,
		show,
		done,
		back,
		localBnStepID,
	)
}

// Get a random client compatible with the user's hardware and EC choices.
func selectRandomBn(goodOptions []*config.ParameterOption[config.BeaconNode], includeSupermajority bool, wiz *wizard, currentStep int, totalSteps int) {
	// Get system specs
	totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
	isLowPower := (totalMemoryGB < 15 || runtime.GOARCH == "arm64")

	// Filter out the clients based on system specs
	filteredClients := []config.BeaconNode{}
	for _, clientOption := range goodOptions {
		client := clientOption.Value
		switch client {
		case config.BeaconNode_Teku:
			if !isLowPower {
				filteredClients = append(filteredClients, client)
			}
		/*
			case config.BeaconNode_Prysm:
				if includeSupermajority {
					filteredClients = append(filteredClients, client)
				}
		*/
		default:
			filteredClients = append(filteredClients, client)
		}
	}

	// Select a random client
	selectedClient := filteredClients[rand.Intn(len(filteredClients))]
	wiz.md.Config.LocalBeaconClient.BeaconNode.Value = selectedClient

	// Show the selection page
	/*
		if selectedClient == config.BeaconNode_Prysm {
			wiz.consensusLocalRandomPrysmModal = createRandomPrysmStep(wiz, currentStep, totalSteps, goodOptions)
			wiz.consensusLocalRandomPrysmModal.show()
		} else {
			wiz.consensusLocalRandomModal = createRandomStep(wiz, currentStep, totalSteps, goodOptions)
			wiz.consensusLocalRandomModal.show()
		}
	*/
	wiz.localBnRandomModal = createRandomBnStep(wiz, currentStep, totalSteps, goodOptions)
	wiz.localBnRandomModal.show()
}

// Get a more verbose client description, including warnings
func getAugmentedBnDescription(client config.BeaconNode, originalDescription string) string {
	switch client {
	/*
		case config.BeaconNode_Prysm:
			return fmt.Sprintf("%s\n\n[orange]NOTE: Prysm currently has a very high representation of the Beacon Chain. For the health of the network and the overall safety of your funds, please consider choosing a client with a lower representation. Please visit https://clientdiversity.org to learn more.", originalDescription)
	*/
	case config.BeaconNode_Teku:
		totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
		if runtime.GOARCH == "arm64" || totalMemoryGB < 15 {
			return fmt.Sprintf("%s\n\n[orange]WARNING: Teku is a resource-heavy client and will likely not perform well on your system given your CPU power or amount of available RAM. We recommend you pick a lighter client instead.", originalDescription)
		}
	}

	return originalDescription
}
