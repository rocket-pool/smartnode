package config

import (
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func createUseFallbackStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {
	helperText := "If you have an extra externally-managed Execution Client and Beacon Node pair that you trust, you can use them as \"fallback\" clients.\nThe Smart Node and your Validator Client will connect to these if your primary clients go offline for any reason, so your node will continue functioning properly until your primary clients are back online.\n\nWould you like to use a fallback client pair?"

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		if wiz.md.Config.Fallback.UseFallbackClients.Value == false {
			modal.focus(0)
		} else {
			modal.focus(1)
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 1 {
			wiz.md.Config.Fallback.UseFallbackClients.Value = true
			bn := wiz.md.Config.GetSelectedBeaconNode()
			switch bn {
			case cfgtypes.ConsensusClient_Prysm:
				wiz.fallbackPrysmModal.show()
			default:
				wiz.fallbackNormalModal.show()
			}
		} else {
			wiz.md.Config.Fallback.UseFallbackClients.Value = false
			wiz.metricsModal.show()
		}
	}

	back := func() {
		wiz.doppelgangerDetectionModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"No", "Yes"},
		[]string{},
		76,
		"Use Fallback Clients",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-use-fallback",
	)
}
