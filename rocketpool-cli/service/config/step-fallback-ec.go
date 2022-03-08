package config

import (
	"github.com/rocket-pool/smartnode/shared/services/config"
)

func createFallbackEcStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	// Create the button names and descriptions from the config
	clients := wiz.md.Config.FallbackExecutionClient.Options
	clientNames := []string{"None"}
	clientDescriptions := []string{"Do not use a fallback client."}
	for _, client := range clients {
		clientNames = append(clientNames, client.Name)
		clientDescriptions = append(clientDescriptions, client.Description)
	}

	helperText := "If you would like to add a fallback Execution client, please choose it below.\n\nThe Smartnode will temporarily use this instead of your main Execution client if the main client ever fails.\nIt will switch back to the main client when it starts working again."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		if wiz.md.Config.UseFallbackExecutionClient.Value == true {
			// Focus the selected option
			for i, option := range wiz.md.Config.FallbackExecutionClient.Options {
				if option.Value == wiz.md.Config.FallbackExecutionClient.Value {
					modal.focus(i)
					break
				}
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.Config.UseFallbackExecutionClient.Value = false
			wiz.consensusModeModal.show()
		} else {
			wiz.md.Config.UseFallbackExecutionClient.Value = true
			selectedClient := clients[buttonIndex-1].Value.(config.ExecutionClient)
			wiz.md.Config.FallbackExecutionClient.Value = selectedClient
			switch selectedClient {
			case config.ExecutionClient_Infura:
				wiz.fallbackInfuraModal.show()
			default:
				wiz.consensusModeModal.show()
			}
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
		70,
		"Fallback Execution Client",
		DirectionalModalVertical,
		show,
		done,
		back,
		"step-fallback-ec",
	)

}
