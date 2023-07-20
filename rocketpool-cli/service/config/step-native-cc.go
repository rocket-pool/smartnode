package config

import (
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func createNativeCcStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	// Create the button names and descriptions from the config
	clients := wiz.md.Config.Native.ConsensusClient.Options
	clientNames := []string{}
	clientDescriptions := []string{}
	for _, client := range clients {
		clientNames = append(clientNames, client.Name)
		clientDescriptions = append(clientDescriptions, getAugmentedCcDescription(client.Value.(cfgtypes.ConsensusClient), client.Description))
	}

	helperText := "Please select the Consensus client you are / will be using.\n\nIf you're still deciding on one, highlight each one below to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth2-clients to learn more about them."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		if !wiz.md.isNew {
			for i, option := range wiz.md.Config.Native.ConsensusClient.Options {
				if option.Value == wiz.md.Config.Native.ConsensusClient.Value {
					modal.focus(i)
					break
				}
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		selectedClient := clients[buttonIndex].Value.(cfgtypes.ConsensusClient)
		wiz.md.Config.Native.ConsensusClient.Value = selectedClient
		wiz.nativeCcUrlModal.show()
	}

	back := func() {
		wiz.nativeEcModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		clientNames,
		clientDescriptions,
		100,
		"Consensus Client > Selection",
		DirectionalModalVertical,
		show,
		done,
		back,
		"step-native-cc",
	)

}
