package config

import (
	"github.com/rocket-pool/smartnode/shared/services/config"
)

func createLocalEcStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	// Create the button names and descriptions from the config
	clients := wiz.md.Config.ExecutionClient.Options
	clientNames := []string{}
	clientDescriptions := []string{}
	for _, client := range clients {
		clientNames = append(clientNames, client.Name)
		clientDescriptions = append(clientDescriptions, client.Description)
	}

	helperText := "Please select the Execution client you would like to use.\n\nHighlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth1-clients to learn more about them."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		for i, option := range wiz.md.Config.ExecutionClient.Options {
			if option.Value == wiz.md.Config.ExecutionClient.Value {
				modal.focus(i)
				break
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		selectedClient := clients[buttonIndex].Value.(config.ExecutionClient)
		wiz.md.Config.ExecutionClient.Value = selectedClient
		switch selectedClient {
		case config.ExecutionClient_Infura:
			// Switch to the Infura dialog
			wiz.infuraModal.show()
		default:
			wiz.fallbackExecutionModal.show()
		}
	}

	back := func() {
		wiz.executionModeModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		clientNames,
		clientDescriptions,
		76,
		"Execution Client > Selection",
		DirectionalModalVertical,
		show,
		done,
		back,
		"step-ec-local",
	)
}
