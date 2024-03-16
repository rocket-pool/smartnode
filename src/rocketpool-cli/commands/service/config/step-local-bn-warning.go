package config

func createPrysmWarningStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {
	helperText := "[orange]NOTE: Prysm currently has a very high representation of the Beacon Chain. For the health of the network and the overall safety of your funds, please consider choosing a client with a lower representation. Please visit https://clientdiversity.org to learn more."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.localBnModal.show()
		} else {
			wiz.checkpointSyncProviderModal.show()
		}
	}

	back := func() {
		wiz.localBnModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Choose Again", "Keep Prysm"},
		[]string{},
		76,
		"Beacon Node > Selection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-prysm-warning",
	)
}

func createTekuWarningStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {
	helperText := "[orange]WARNING: Teku is a resource-heavy client and will likely not perform well on your system given your CPU power or amount of available RAM. We recommend you pick a lighter client instead."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.localBnModal.show()
		} else {
			wiz.checkpointSyncProviderModal.show()
		}
	}

	back := func() {
		wiz.localBnModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Choose Again", "Keep Teku"},
		[]string{},
		76,
		"Beacon Node > Selection",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-teku-warning",
	)
}
