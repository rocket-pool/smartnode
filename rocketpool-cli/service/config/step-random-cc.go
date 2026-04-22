package config

import (
	"fmt"

	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

const randomCcPrysmID string = "step-random-cc-prysm"
const randomCcID string = "step-random-cc"

func createRandomCCStep(wiz *wizard, currentStep int, totalSteps int, goodOptions []cfgtypes.ParameterOption) *choiceWizardStep {

	var selectedClientName string
	selectedClient := wiz.md.Config.ConsensusClient.Value
	for _, clientOption := range goodOptions {
		if clientOption.Value == selectedClient {
			selectedClientName = clientOption.Name
			break
		}
	}

	helperText := fmt.Sprintf("You have been randomly assigned to %s for your Consensus client.", selectedClientName)

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		wiz.graffitiModal.show()
	}

	back := func() {
		wiz.consensusLocalModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Ok"},
		[]string{},
		76,
		"Consensus Client > Selection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		randomCcID,
	)

}
