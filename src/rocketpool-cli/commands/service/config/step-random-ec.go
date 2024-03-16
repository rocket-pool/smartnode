package config

import (
	"fmt"

	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

const randomEcID string = "step-random-ec"

func createRandomECStep(wiz *wizard, currentStep int, totalSteps int, goodOptions []cfgtypes.ParameterOption) *choiceWizardStep {
	var selectedClientName string
	selectedClient := wiz.md.Config.LocalExecutionClient.ExecutionClient.Value
	for _, clientOption := range goodOptions {
		if clientOption.Value == selectedClient {
			selectedClientName = clientOption.Name
			break
		}
	}

	helperText := fmt.Sprintf("You have been randomly assigned to %s for your Execution Client.", selectedClientName)

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		wiz.localBnModal.show()
	}

	back := func() {
		wiz.localEcModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Ok"},
		[]string{},
		76,
		"Execution Client > Selection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		randomEcID,
	)
}
