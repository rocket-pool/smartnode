package config

func createFinishedStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "All done! You're ready to run.\n\nIf you'd like, you can review and change all of the Smartnode and client settings next or just save and exit."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.pages.RemovePage(settingsHomeID)
			wiz.md.settingsHome = newSettingsHome(wiz.md)
			wiz.md.setPage(wiz.md.settingsHome.homePage)
		} else {
			wiz.md.ShouldSave = true
			wiz.md.app.Stop()
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
