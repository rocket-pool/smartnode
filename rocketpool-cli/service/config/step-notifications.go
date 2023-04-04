package config

func createNotificationsStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "Would you like to enable the Smartnode's notification system?"

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		if wiz.md.Config.EnableNotifications.Value == false {
			modal.focus(0)
		} else {
			modal.focus(1)
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 1 {
			wiz.md.Config.EnableNotifications.Value = true
		} else {
			wiz.md.Config.EnableNotifications.Value = false
		}
		wiz.mevModeModal.show()
	}

	back := func() {
		wiz.useFallbackModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"No", "Yes"},
		[]string{},
		76,
		"Notifications",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-notifications",
	)

}
