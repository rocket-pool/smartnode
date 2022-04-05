package config

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared"
)

func createNativeWelcomeStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := fmt.Sprintf("%s\n\nWelcome to the Smartnode configuration wizard!\n\nWe've detected that you're running in Native mode. We'll keep it simple and only show you the settings that are relevant to you.\n\nIf you're upgrading from a previous version, your settings have been migrated.", shared.Logo)

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(1)
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 1 {
			wiz.nativeNetworkModal.show()
		} else {
			wiz.md.app.Stop()
		}
	}

	back := func() {
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Quit", "Next"},
		nil,
		60,
		"Welcome",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-native-welcome",
	)

}
