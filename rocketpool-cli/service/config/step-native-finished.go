package config

func createNativeFinishedStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "All done! You're ready to run.\n\nIf you'd like, you can review and change all of the Smartnode and Native settings next or just save and exit."

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

			wiz.md.pages.RemovePage(settingsNativeHomeID)
			wiz.md.settingsNativeHome = newSettingsNativeHome(wiz.md)
			wiz.md.setPage(wiz.md.settingsNativeHome.homePage)
		} else {
			wiz.md.ShouldSave = true
			wiz.md.app.Stop()
		}
	}

	back := func() {
		wiz.nativeMetricsModal.show()
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
		"step-native-finished",
	)

}
