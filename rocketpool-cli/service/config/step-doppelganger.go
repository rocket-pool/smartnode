package config

import (
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func createDoppelgangerStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "Your client supports Doppelganger Protection. This feature can prevent your minipools from being slashed (penalized for a lot of ETH and removed from the Beacon Chain) if you accidentally run your validator keys on multiple machines at the same time.\n\nIf enabled, whenever your validator client restarts, it will intentionally miss 2-3 attestations (for each minipool). If all of them are missed successfully, you can be confident that you are safe to start attesting.\n\nWould you like to enable Doppelganger Protection?"

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		if wiz.md.Config.ConsensusCommon.DoppelgangerDetection.Value == false {
			modal.focus(0)
		} else {
			modal.focus(1)
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 1 {
			wiz.md.Config.ConsensusCommon.DoppelgangerDetection.Value = true
		} else {
			wiz.md.Config.ConsensusCommon.DoppelgangerDetection.Value = false
		}
		cc, _ := wiz.md.Config.GetSelectedConsensusClient()
		switch cc {
		case cfgtypes.ConsensusClient_Nimbus, cfgtypes.ConsensusClient_Teku:
			wiz.md.Config.UseFallbackClients.Value = false
			wiz.metricsModal.show()
		default:
			wiz.useFallbackModal.show()
		}
	}

	back := func() {
		wiz.graffitiModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"No", "Yes"},
		[]string{},
		76,
		"Consensus Client > Doppelganger Protection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-doppelganger",
	)

}
