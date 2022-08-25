package config

import (
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func createExternalGraffitiStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels - use the vanilla graffiti name
	graffitiLabel := wiz.md.Config.ConsensusCommon.Graffiti.Name

	helperText := "If you would like to add a short custom message to each block that your minipools propose (called the block's \"graffiti\"), please enter it here.\n\nThis is completely optional and just for fun. Leave it blank if you don't want to add any graffiti.\n\nThe graffiti is limited to 16 characters max."

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		switch wiz.md.Config.ExternalConsensusClient.Value.(cfgtypes.ConsensusClient) {
		case cfgtypes.ConsensusClient_Lighthouse:
			modal.textboxes[graffitiLabel].SetText(wiz.md.Config.ExternalLighthouse.Graffiti.Value.(string))
		case cfgtypes.ConsensusClient_Lodestar:
			modal.textboxes[graffitiLabel].SetText(wiz.md.Config.ExternalLodestar.Graffiti.Value.(string))
		case cfgtypes.ConsensusClient_Prysm:
			modal.textboxes[graffitiLabel].SetText(wiz.md.Config.ExternalPrysm.Graffiti.Value.(string))
		case cfgtypes.ConsensusClient_Teku:
			modal.textboxes[graffitiLabel].SetText(wiz.md.Config.ExternalTeku.Graffiti.Value.(string))
		}
	}

	done := func(text map[string]string) {
		// Get the selected client
		switch wiz.md.Config.ExternalConsensusClient.Value.(cfgtypes.ConsensusClient) {
		case cfgtypes.ConsensusClient_Lighthouse:
			wiz.md.Config.ExternalLighthouse.Graffiti.Value = text[graffitiLabel]
			wiz.externalDoppelgangerModal.show()
		case cfgtypes.ConsensusClient_Nimbus:
			wiz.md.Config.ExternalNimbus.Graffiti.Value = text[graffitiLabel]
			wiz.externalDoppelgangerModal.show()
		case cfgtypes.ConsensusClient_Lodestar:
			wiz.md.Config.ExternalLodestar.Graffiti.Value = text[graffitiLabel]
			wiz.externalDoppelgangerModal.show()
		case cfgtypes.ConsensusClient_Prysm:
			wiz.md.Config.ExternalPrysm.Graffiti.Value = text[graffitiLabel]
			wiz.externalDoppelgangerModal.show()
		case cfgtypes.ConsensusClient_Teku:
			wiz.md.Config.ExternalTeku.Graffiti.Value = text[graffitiLabel]
			wiz.externalDoppelgangerModal.show()
		}
	}

	back := func() {
		if wiz.md.Config.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
			wiz.consensusLocalModal.show()
		} else {
			wiz.consensusExternalSelectModal.show()
		}
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		70,
		"Consensus Client (External) > Graffiti",
		[]string{graffitiLabel},
		[]int{wiz.md.Config.ConsensusCommon.Graffiti.MaxLength},
		[]string{wiz.md.Config.ConsensusCommon.Graffiti.Regex},
		show,
		done,
		back,
		"step-external-graffiti",
	)

}
