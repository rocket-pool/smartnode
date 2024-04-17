package config

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/config"
)

const randomBnPrysmID string = "step-random-bn-prysm"
const randomBnID string = "step-random-bn"

func createRandomBnStep(wiz *wizard, currentStep int, totalSteps int, goodOptions []*config.ParameterOption[config.BeaconNode]) *choiceWizardStep {
	var selectedClientName string
	selectedClient := wiz.md.Config.LocalBeaconClient.BeaconNode.Value
	for _, clientOption := range goodOptions {
		if clientOption.Value == selectedClient {
			selectedClientName = clientOption.Name
			break
		}
	}

	helperText := fmt.Sprintf("You have been randomly assigned to %s for your Beacon Node.", selectedClientName)

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		wiz.checkpointSyncProviderModal.show()
	}

	back := func() {
		wiz.localBnModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Ok"},
		[]string{},
		76,
		"Beacon Node > Selection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		randomBnID,
	)
}
