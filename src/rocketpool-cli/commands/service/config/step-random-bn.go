package config

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/config"
)

const randomBnPrysmID string = "step-random-bn-prysm"
const randomBnID string = "step-random-bn"

func createRandomPrysmStep(wiz *wizard, currentStep int, totalSteps int, goodOptions []*config.ParameterOption[config.BeaconNode]) *choiceWizardStep {
	helperText := "You have been randomly assigned to Prysm for your Beacon Node.\n\n[orange]NOTE: Prysm currently has a very high representation of the Beacon Chain. For the health of the network and the overall safety of your funds, please consider choosing a client with a lower representation. Please visit https://clientdiversity.org to learn more."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			selectRandomBn(goodOptions, false, wiz, currentStep, totalSteps)
		} else {
			wiz.checkpointSyncProviderModal.show()
		}
	}

	back := func() {
		wiz.localBnModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Choose Another Random Client", "Keep Prysm"},
		[]string{},
		76,
		"Beacon Node > Selection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		randomBnPrysmID,
	)
}

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
