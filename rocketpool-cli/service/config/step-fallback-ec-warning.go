package config

func createFallbackInfuraWarningStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "[orange]WARNING: Infura is deprecated as light clients are NOT SUPPORTED by the upcoming Ethereum Merge. It will be removed from the Smartnode in a future release, and you will have to use a separate Full Execution client on a separate machine as your fallback by using the \"Externally Managed\" mode."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.fallbackExecutionModal.show()
		} else {
			wiz.fallbackInfuraModal.show()
		}
	}

	back := func() {
		wiz.fallbackExecutionModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Choose Again", "Keep Infura"},
		[]string{},
		76,
		"Fallback Execution Client > Selection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-fallback-infura-warning",
	)

}

func createFallbackPocketWarningStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "[orange]WARNING: Pocket is deprecated as light clients are NOT SUPPORTED by the upcoming Ethereum Merge. It will be removed from the Smartnode in a future release, and you will have to use a separate Full Execution client on a separate machine as your fallback by using the \"Externally Managed\" mode."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.fallbackExecutionModal.show()
		} else {
			wiz.consensusModeModal.show()
		}
	}

	back := func() {
		wiz.fallbackExecutionModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Choose Again", "Keep Pocket"},
		[]string{},
		76,
		"Fallback Execution Client > Selection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-fallback-pocket-warning",
	)

}
