package config

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/smartnode/shared/services/config"
)

const localCcStepID string = "step-local-cc"

func createLocalCcStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	// Get the list of clients
	goodClients, badClients := wiz.md.Config.GetCompatibleConsensusClients()

	// Create the button names and descriptions from the config
	clients := wiz.md.Config.ConsensusClient.Options
	clientNames := []string{}
	clientDescriptions := []string{}
	for _, client := range goodClients {
		clientNames = append(clientNames, client.Name)
		clientDescriptions = append(clientDescriptions, client.Description)
	}

	incompatibleClientWarning := ""
	if len(badClients) > 0 {
		incompatibleClientWarning = fmt.Sprintf("\n\n[orange]NOTE: The following clients are incompatible with your choice of Execution and/or fallback Execution clients: %s", strings.Join(badClients, ", "))
	}

	helperText := fmt.Sprintf("Please select the Consensus client you would like to use.\n\n"+
		"Highlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth2-clients to learn more about them.%s", incompatibleClientWarning)

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		for i, option := range wiz.md.Config.ConsensusClient.Options {
			if option.Value == wiz.md.Config.ConsensusClient.Value {
				modal.focus(i)
				break
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		selectedClient := clients[buttonIndex].Value.(config.ConsensusClient)
		wiz.md.Config.ConsensusClient.Value = selectedClient
		wiz.graffitiModal.show()
	}

	back := func() {
		wiz.consensusModeModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		clientNames,
		clientDescriptions,
		76,
		"Consensus Client > Selection",
		DirectionalModalVertical,
		show,
		done,
		back,
		localCcStepID,
	)

}
