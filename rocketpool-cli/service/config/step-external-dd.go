package config

import (
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func createExternalDoppelgangerStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "Your client supports Doppelganger Protection. This feature can prevent your minipools from being slashed (penalized for a lot of ETH and removed from the Beacon Chain) if you accidentally run your validator keys on multiple machines at the same time.\n\nIf enabled, whenever your validator client restarts, it will intentionally miss 2-3 attestations (for each minipool). If all of them are missed successfully, you can be confident that you are safe to start attesting.\n\nWould you like to enable Doppelganger Protection?"

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		ddEnabled := true
		switch wiz.md.Config.ExternalConsensusClient.Value.(cfgtypes.ConsensusClient) {
		case cfgtypes.ConsensusClient_Lighthouse:
			ddEnabled = (wiz.md.Config.ExternalLighthouse.DoppelgangerDetection.Value == true)
		case cfgtypes.ConsensusClient_Lodestar:
			ddEnabled = (wiz.md.Config.ExternalLodestar.DoppelgangerDetection.Value == true)
		case cfgtypes.ConsensusClient_Prysm:
			ddEnabled = (wiz.md.Config.ExternalPrysm.DoppelgangerDetection.Value == true)
		case cfgtypes.ConsensusClient_Teku:
			ddEnabled = (wiz.md.Config.ExternalTeku.DoppelgangerDetection.Value == true)
		}

		if ddEnabled {
			modal.focus(1)
		} else {
			modal.focus(0)
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		ddEnabled := false
		if buttonIndex == 1 {
			ddEnabled = true
		}
		switch wiz.md.Config.ExternalConsensusClient.Value.(cfgtypes.ConsensusClient) {
		case cfgtypes.ConsensusClient_Lighthouse:
			wiz.md.Config.ExternalLighthouse.DoppelgangerDetection.Value = ddEnabled
		case cfgtypes.ConsensusClient_Nimbus:
			wiz.md.Config.ExternalNimbus.DoppelgangerDetection.Value = ddEnabled
		case cfgtypes.ConsensusClient_Lodestar:
			wiz.md.Config.ExternalLodestar.DoppelgangerDetection.Value = ddEnabled
		case cfgtypes.ConsensusClient_Prysm:
			wiz.md.Config.ExternalPrysm.DoppelgangerDetection.Value = ddEnabled
		case cfgtypes.ConsensusClient_Teku:
			wiz.md.Config.ExternalTeku.DoppelgangerDetection.Value = ddEnabled
		}
		wiz.useFallbackModal.show()
	}

	back := func() {
		wiz.externalGraffitiModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"No", "Yes"},
		[]string{},
		76,
		"Consensus Client (External) > Doppelganger Protection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-external-doppelganger",
	)

}
