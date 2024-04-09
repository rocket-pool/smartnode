package config

import "github.com/rocket-pool/node-manager-core/config"

func createGraffitiStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {
	// Create the labels
	graffitiLabel := wiz.md.Config.ValidatorClient.VcCommon.Graffiti.Name

	helperText := "If you would like to add a short custom message to each block that your minipools propose (called the block's \"graffiti\"), please enter it here.\n\nThis is completely optional and just for fun. Leave it blank if you don't want to add any graffiti.\n\nThe graffiti is limited to 16 characters max."

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.ValidatorClient.VcCommon.GetParameters() {
				if param.GetCommon().Name == label {
					box.SetText(param.String())
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.ValidatorClient.VcCommon.Graffiti.Value = text[graffitiLabel]
		wiz.doppelgangerDetectionModal.show()
	}

	back := func() {
		if wiz.md.Config.IsLocalMode() {
			wiz.checkpointSyncProviderModal.show()
		} else if wiz.md.Config.ExternalBeaconClient.BeaconNode.Value == config.BeaconNode_Prysm {
			wiz.externalPrysmSettingsModal.show()
		} else {
			wiz.externalBnSettingsModal.show()
		}
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		70,
		"Validator Client > Graffiti",
		[]string{graffitiLabel},
		[]int{wiz.md.Config.ValidatorClient.VcCommon.Graffiti.MaxLength},
		[]string{wiz.md.Config.ValidatorClient.VcCommon.Graffiti.Regex},
		show,
		done,
		back,
		"step-vc-graffiti",
	)
}
