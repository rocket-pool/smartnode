package config

import "github.com/rocket-pool/node-manager-core/config"

func createExternalBnStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {
	// Create the button names and descriptions from the config
	clients := wiz.md.Config.ExternalBeaconClient.BeaconNode.Options
	clientNames := []string{}
	for _, client := range clients {
		clientNames = append(clientNames, client.Name)
	}

	helperText := "Which Beacon Node are you externally managing? Each of them has small behavioral differences, so we'll need to know which one you're using in order to connect to it properly."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		for i, option := range wiz.md.Config.ExternalBeaconClient.BeaconNode.Options {
			if option.Value == wiz.md.Config.ExternalBeaconClient.BeaconNode.Value {
				modal.focus(i)
				break
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		selectedClient := clients[buttonIndex].Value
		wiz.md.Config.ExternalBeaconClient.BeaconNode.Value = selectedClient
		switch selectedClient {
		case config.BeaconNode_Prysm:
			wiz.externalPrysmSettingsModal.show()
		default:
			wiz.externalBnSettingsModal.show()
		}
	}

	back := func() {
		wiz.externalEcSettingsModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		clientNames,
		[]string{},
		70,
		"Beacon Node (External) > Selection",
		DirectionalModalVertical,
		show,
		done,
		back,
		"step-external-bn-select",
	)
}
