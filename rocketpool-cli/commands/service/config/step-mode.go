package config

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/config"
)

func createModeStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {
	// Create the button names and descriptions from the config
	modes := wiz.md.Config.ClientMode.Options
	modeNames := []string{}
	modeDescriptions := []string{}
	for _, mode := range modes {
		modeNames = append(modeNames, mode.Name)
		modeDescriptions = append(modeDescriptions, mode.Description)
	}

	helperText := "Now let's decide which mode you'd like to use for the Execution Client and Beacon Node.\n\nWould you like the Smart Node to run and manage its own clients, or would you like it to use an existing clients you run and manage outside of the Smart Node (also known as \"Hybrid Mode\")?"

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		for i, option := range wiz.md.Config.ClientMode.Options {
			if option.Value == wiz.md.Config.ClientMode.Value {
				modal.focus(i)
				break
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		wiz.md.Config.ClientMode.Value = modes[buttonIndex].Value
		switch modes[buttonIndex].Value {
		case config.ClientMode_Local:
			wiz.localEcModal.show()
		case config.ClientMode_External:
			wiz.externalEcSelectModal.show()
		default:
			panic(fmt.Sprintf("Unknown client mode %s", modes[buttonIndex].Value))
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
		"Client Mode",
		DirectionalModalVertical,
		show,
		done,
		back,
		"step-mode",
	)
}
