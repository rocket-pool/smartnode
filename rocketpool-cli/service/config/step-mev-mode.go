package config

import (
	"fmt"

	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func createMevModeStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	// Create the button names and descriptions from the config
	modes := wiz.md.Config.MevBoost.Mode.Options
	modeNames := []string{}
	modeDescriptions := []string{}
	for _, mode := range modes {
		modeNames = append(modeNames, mode.Name)
		modeDescriptions = append(modeDescriptions, mode.Description)
	}

	helperText := "By default, your Smartnode has MEV-Boost enabled. This allows you to capture extra profits from block proposals. Would you like Rocket Pool to manage MEV-Boost for you, or would you like to manage it yourself?\n\n[lime]Please read our guide to learn more about MEV:\nhttps://docs.rocketpool.net/guides/node/mev.html\n"

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		for i, option := range wiz.md.Config.MevBoost.Mode.Options {
			if option.Value == wiz.md.Config.MevBoost.Mode.Value {
				modal.focus(i)
				break
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		wiz.md.Config.MevBoost.Mode.Value = modes[buttonIndex].Value
		switch modes[buttonIndex].Value {
		case cfgtypes.Mode_Local:
			wiz.localMevModal.show()
		case cfgtypes.Mode_External:
			switch wiz.md.Config.ExecutionClientMode.Value {
			case cfgtypes.Mode_Local:
				wiz.externalMevModal.show()
			case cfgtypes.Mode_External:
				wiz.md.Config.EnableMevBoost.Value = true
				wiz.md.Config.MevBoost.Mode.Value = cfgtypes.Mode_External
				wiz.finishedModal.show()
			default:
				panic(fmt.Sprintf("Unknown EC mode %s during MEV mode selection", wiz.md.Config.ExecutionClientMode.Value))
			}
		default:
			panic(fmt.Sprintf("Unknown MEV mode %s", modes[buttonIndex].Value))
		}
	}

	back := func() {
		wiz.metricsModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		modeNames,
		modeDescriptions,
		76,
		"MEV-Boost Mode",
		DirectionalModalVertical,
		show,
		done,
		back,
		"step-mev-mode",
	)
}
