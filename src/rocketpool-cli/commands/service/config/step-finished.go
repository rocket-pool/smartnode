package config

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
	"github.com/rocket-pool/node-manager-core/config"
	snCfg "github.com/rocket-pool/smartnode/v2/shared/config"
)

func createFinishedStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {
	helperText := "All done! You're ready to run.\n\nIf you'd like, you can review and change all of the Smart Node and client settings next or just save and exit."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			// If this is a new installation, reset it with the current settings as the new ones
			if wiz.md.isNew {
				wiz.md.PreviousConfig = wiz.md.Config.CreateCopy()
			}

			wiz.md.pages.RemovePage(settingsHomeID)
			wiz.md.settingsHome = newSettingsHome(wiz.md)
			wiz.md.setPage(wiz.md.settingsHome.homePage)
		} else {
			processConfigAfterQuit(wiz.md)
		}
	}

	back := func() {
		if wiz.md.Config.Network.Value == config.Network_Holesky || wiz.md.Config.Network.Value == snCfg.Network_Devnet {
			// Skip MEV for Holesky
			wiz.metricsModal.show()
		} else {
			wiz.mevModeModal.show()
		}
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{
			"Review All Settings",
			"Save and Exit",
		},
		nil,
		40,
		"Finished",
		DirectionalModalVertical,
		show,
		done,
		back,
		"step-finished",
	)
}

// Processes a configuration after saving and exiting without looking at the review screen
func processConfigAfterQuit(md *mainDisplay) {
	errors := md.Config.Validate()
	if len(errors) > 0 {
		builder := strings.Builder{}
		builder.WriteString("[orange]WARNING: Your configuration encountered errors. You must correct the following in order to save it:\n\n")
		for _, err := range errors {
			builder.WriteString(fmt.Sprintf("%s\n\n", err))
		}

		modal := tview.NewModal().
			SetText(builder.String()).
			AddButtons([]string{"Go to Settings Manager"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				// If this is a new installation, reset it with the current settings as the new ones
				if md.isNew {
					md.PreviousConfig = md.Config.CreateCopy()
				}

				md.app.SetRoot(md.mainGrid, true)
				md.pages.RemovePage(settingsHomeID)
				md.settingsHome = newSettingsHome(md)
				md.setPage(md.settingsHome.homePage)
			})

		md.app.SetRoot(modal, false).SetFocus(modal)
	} else {
		// Get the map of changed settings by category
		_, totalAffectedContainers, changeNetworks := md.Config.GetChanges(md.PreviousConfig)

		if md.isUpdate {
			totalAffectedContainers[config.ContainerID_Daemon] = true
		}

		var containersToRestart []config.ContainerID
		for container := range totalAffectedContainers {
			containersToRestart = append(containersToRestart, container)
		}

		md.ShouldSave = true
		md.ContainersToRestart = containersToRestart
		if changeNetworks && !md.isNew {
			md.ChangeNetworks = true
		}
		md.app.Stop()
	}
}
