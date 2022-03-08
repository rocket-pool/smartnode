package config

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/config"
)

func createEcModeStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	// Create the button names and descriptions from the config
	modes := wiz.md.Config.ExecutionClientMode.Options
	modeNames := []string{}
	modeDescriptions := []string{}
	for _, mode := range modes {
		modeNames = append(modeNames, mode.Name)
		modeDescriptions = append(modeDescriptions, mode.Description)
	}

	helperText := "Now let's decide which mode you'd like to use for the Execution client (formerly eth1 client).\n\nWould you like Rocket Pool to run and manage its own client, or would you like it to use an existing client you run and manage outside of Rocket Pool (also known as \"Hybrid Mode\")?"

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		for i, option := range wiz.md.Config.ExecutionClientMode.Options {
			if option.Value == wiz.md.Config.ExecutionClientMode.Value {
				modal.focus(i)
				break
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		wiz.md.Config.ExecutionClientMode.Value = modes[buttonIndex].Value
		switch modes[buttonIndex].Value {
		case config.Mode_Local:
			wiz.executionLocalModal.show()
		case config.Mode_External:
			wiz.executionExternalModal.show()
		default:
			panic(fmt.Sprintf("Unknown execution client mode %s", modes[buttonIndex].Value))
		}
	}

	back := func() {
		wiz.networkModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		modeNames,
		modeDescriptions,
		76,
		"Execution Client Mode",
		DirectionalModalVertical,
		show,
		done,
		back,
		"step-ec-mode",
	)
}
