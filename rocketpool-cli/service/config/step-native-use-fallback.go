package config

func createNativeUseFallbackStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "If you have an extra externally-managed Execution and Consensus client pair that you trust, you can use them as \"fallback\" clients.\nThe Smartnode will connect to these if your primary Execution and Consensus clients go offline for any reason.\n\nWould you like to use a fallback client pair?"

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		if wiz.md.Config.UseFallbackClients.Value == false {
			modal.focus(0)
		} else {
			modal.focus(1)
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 1 {
			wiz.md.Config.UseFallbackClients.Value = true
			wiz.nativeFallbackModal.show()
		} else {
			wiz.md.Config.UseFallbackClients.Value = false
			wiz.nativeMetricsModal.show()
		}
	}

	back := func() {
		wiz.nativeDataModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"No", "Yes"},
		[]string{},
		76,
		"Use Fallback Clients",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-native-use-fallback",
	)

}
