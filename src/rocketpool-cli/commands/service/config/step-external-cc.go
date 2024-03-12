package config

import (
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func createExternalCcStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	// Create the button names and descriptions from the config
	clients := wiz.md.Config.ExternalConsensusClient.Options
	clientNames := []string{}
	for _, client := range clients {
		clientNames = append(clientNames, client.Name)
	}

	helperText := "Which Consensus client are you externally managing? Each of them has small behavioral differences, so we'll need to know which one you're using in order to connect to it properly."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		for i, option := range wiz.md.Config.ExternalConsensusClient.Options {
			if option.Value == wiz.md.Config.ExternalConsensusClient.Value {
				modal.focus(i)
				break
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		selectedClient := clients[buttonIndex].Value.(cfgtypes.ConsensusClient)
		wiz.md.Config.ExternalConsensusClient.Value = selectedClient
		switch selectedClient {
		case cfgtypes.ConsensusClient_Lighthouse:
			wiz.lighthouseExternalSettingsModal.show()
		case cfgtypes.ConsensusClient_Nimbus:
			wiz.nimbusExternalSettingsModal.show()
		case cfgtypes.ConsensusClient_Lodestar:
			wiz.lodestarExternalSettingsModal.show()
		case cfgtypes.ConsensusClient_Prysm:
			wiz.prysmExternalSettingsModal.show()
		case cfgtypes.ConsensusClient_Teku:
			wiz.tekuExternalSettingsModal.show()
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
		[]string{},
		70,
		"Consensus Client (External) > Selection",
		DirectionalModalVertical,
		show,
		done,
		back,
		"step-external-cc",
	)

}
