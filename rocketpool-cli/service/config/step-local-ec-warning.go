package config

func createInfuraWarningStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "[orange]WARNING: Infura is deprecated as light clients are NOT SUPPORTED by the upcoming Ethereum Merge. It will be removed from the Smartnode in a future release. If you use Infura as a primary Execution client, your validator will NO LONGER WORK after the Merge. We strongly encourage you to pick a Full Execution client instead."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.executionLocalModal.show()
		} else {
			wiz.infuraModal.show()
		}
	}

	back := func() {
		wiz.executionLocalModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Choose Again", "Keep Infura"},
		[]string{},
		76,
		"Execution Client > Selection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-infura-warning",
	)

}

func createPocketWarningStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "[orange]WARNING: Pocket is deprecated as light clients are NOT SUPPORTED by the upcoming Ethereum Merge. It will be removed from the Smartnode in a future release. If you use Infura as a primary Execution client, your validator will NO LONGER WORK after the Merge. We strongly encourage you to pick a Full Execution client instead."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.executionLocalModal.show()
		} else {
			wiz.fallbackExecutionModal.show()
		}
	}

	back := func() {
		wiz.executionLocalModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Choose Again", "Keep Pocket"},
		[]string{},
		76,
		"Execution Client > Selection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-pocket-warning",
	)

}
