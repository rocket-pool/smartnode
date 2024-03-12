package config

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func createLocalEcStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	// Create the button names and descriptions from the config
	clients := wiz.md.Config.ExecutionClient.Options
	clientNames := []string{"Random (Recommended)"}
	clientDescriptions := []string{"Select a client randomly to help promote the diversity of the Ethereum Chain. We recommend you do this unless you have a strong reason to pick a specific client."}
	for _, client := range clients {
		clientNames = append(clientNames, client.Name)
		clientDescriptions = append(clientDescriptions, client.Description)
	}

	goodClients := []cfgtypes.ParameterOption{}
	for _, client := range wiz.md.Config.ExecutionClient.Options {
		if !strings.HasPrefix(client.Name, "*") {
			goodClients = append(goodClients, client)
		}
	}

	helperText := "Please select the Execution client you would like to use.\n\nHighlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth1-clients to learn more about them."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		if !wiz.md.isNew {
			var ecName string
			for _, option := range wiz.md.Config.ExecutionClient.Options {
				if option.Value == wiz.md.Config.ExecutionClient.Value {
					ecName = option.Name
					break
				}
			}
			for i, clientName := range clientNames {
				if ecName == clientName {
					modal.focus(i)
					break
				}
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.pages.RemovePage(randomCcPrysmID)
			wiz.md.pages.RemovePage(randomCcID)
			selectRandomEC(goodClients, wiz, currentStep, totalSteps)
		} else {
			buttonLabel = strings.TrimSpace(buttonLabel)
			selectedClient := cfgtypes.ExecutionClient_Unknown
			for _, client := range wiz.md.Config.ExecutionClient.Options {
				if client.Name == buttonLabel {
					selectedClient = client.Value.(cfgtypes.ExecutionClient)
					break
				}
			}
			if selectedClient == cfgtypes.ExecutionClient_Unknown {
				panic(fmt.Sprintf("Local EC selection buttons didn't match any known clients, buttonLabel = %s\n", buttonLabel))
			}
			wiz.md.Config.ExecutionClient.Value = selectedClient
			if wiz.md.Config.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
				wiz.consensusLocalModal.show()
			} else {
				wiz.consensusExternalSelectModal.show()
			}
		}
	}

	back := func() {
		wiz.modeModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		clientNames,
		clientDescriptions,
		100,
		"Execution Client > Selection",
		DirectionalModalVertical,
		show,
		done,
		back,
		"step-ec-local",
	)
}

// Get a random execution client
func selectRandomEC(goodOptions []cfgtypes.ParameterOption, wiz *wizard, currentStep int, totalSteps int) {

	// Get system specs
	//totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
	//isLowPower := (totalMemoryGB < 15 || runtime.GOARCH == "arm64")

	// Filter out the clients based on system specs
	filteredClients := []cfgtypes.ExecutionClient{}
	for _, clientOption := range goodOptions {
		client := clientOption.Value.(cfgtypes.ExecutionClient)
		switch client {
		default:
			filteredClients = append(filteredClients, client)
		}
	}

	// Select a random client
	rand.Seed(time.Now().UnixNano())
	selectedClient := filteredClients[rand.Intn(len(filteredClients))]
	wiz.md.Config.ExecutionClient.Value = selectedClient

	// Show the selection page
	wiz.executionLocalRandomModal = createRandomECStep(wiz, currentStep, totalSteps, goodOptions)
	wiz.executionLocalRandomModal.show()

}
