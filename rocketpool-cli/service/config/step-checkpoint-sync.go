package config

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func createCheckpointSyncStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels and args
	checkpointSyncLabel := wiz.md.Config.ConsensusCommon.CheckpointSyncProvider.Name

	helperText := "Your client supports Checkpoint Sync. This powerful feature allows it to copy the most recent state from a separate Consensus client that you trust, so you don't have to wait for it to sync from scratch - you can start using it instantly!\n\nTake a look at our documentation for an example of how to use it:\nhttps://docs.rocketpool.net/guides/node/config-docker.html#beacon-chain-checkpoint-syncing\n\nIf you would like to use Checkpoint Sync, please provide the provider URL here. If you don't want to use it, leave it blank."

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
		wiz.md.Config.ConsensusCommon.CheckpointSyncProvider.Value = text[checkpointSyncLabel]
		// Get the selected client
		client, err := wiz.md.Config.GetSelectedConsensusClientConfig()
		if err != nil {
			wiz.md.app.Stop()
			fmt.Printf("Error setting the consensus client checkpoint sync provider: %s", err.Error())
		}

		// Check to see if it supports doppelganger detection
		unsupportedParams := client.(cfgtypes.LocalConsensusConfig).GetUnsupportedCommonParams()
		supportsDoppelganger := true
		for _, param := range unsupportedParams {
			if param == config.DoppelgangerDetectionID {
				supportsDoppelganger = false
			}
		}

		// Move to the next appropriate dialog
		if supportsDoppelganger {
			wiz.doppelgangerDetectionModal.show()
		} else {
			wiz.useFallbackModal.show()
		}
	}

	back := func() {
		wiz.graffitiModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		76,
		"Consensus Client > Checkpoint Sync",
		[]string{checkpointSyncLabel},
		[]int{wiz.md.Config.ConsensusCommon.CheckpointSyncProvider.MaxLength},
		[]string{wiz.md.Config.ConsensusCommon.CheckpointSyncProvider.Regex},
		show,
		done,
		back,
		"step-checkpoint-sync",
	)

}
