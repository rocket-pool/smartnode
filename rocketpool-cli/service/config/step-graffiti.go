package config

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func createGraffitiStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	graffitiLabel := wiz.md.Config.ConsensusCommon.Graffiti.Name

	helperText := "If you would like to add a short custom message to each block that your minipools propose (called the block's \"graffiti\"), please enter it here.\n\nThis is completely optional and just for fun. Leave it blank if you don't want to add any graffiti.\n\nThe graffiti is limited to 16 characters max."

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.ConsensusCommon.GetParameters() {
				if param.Name == label {
					box.SetText(fmt.Sprint(param.Value))
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.ConsensusCommon.Graffiti.Value = text[graffitiLabel]
		// Get the selected client
		client, err := wiz.md.Config.GetSelectedConsensusClientConfig()
		if err != nil {
			wiz.md.app.Stop()
			fmt.Printf("Error setting the consensus client graffiti: %s", err.Error())
		}

		// Check to see if it supports checkpoint sync or doppelganger detection
		unsupportedParams := client.(cfgtypes.LocalConsensusConfig).GetUnsupportedCommonParams()
		supportsCheckpointSync := true
		supportsDoppelganger := true
		for _, param := range unsupportedParams {
			if param == config.CheckpointSyncUrlID {
				supportsCheckpointSync = false
			} else if param == config.DoppelgangerDetectionID {
				supportsDoppelganger = false
			}
		}

		// Move to the next appropriate dialog
		if supportsCheckpointSync {
			wiz.checkpointSyncProviderModal.show()
		} else if supportsDoppelganger {
			wiz.doppelgangerDetectionModal.show()
		} else {
			wiz.metricsModal.show()
		}
	}

	back := func() {
		wiz.consensusLocalModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		70,
		"Consensus Client > Graffiti",
		[]string{graffitiLabel},
		[]int{wiz.md.Config.ConsensusCommon.Graffiti.MaxLength},
		[]string{wiz.md.Config.ConsensusCommon.Graffiti.Regex},
		show,
		done,
		back,
		"step-cc-graffiti",
	)

}
