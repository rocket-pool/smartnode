package config

import (
	"fmt"
)

func createFallbackInfuraStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	projectIdLabel := wiz.md.Config.FallbackInfura.ProjectID.Name

	helperText := "Please enter the Project ID for your Infura Ethereum project. You can find this on the Infura website, in your Ethereum project settings."

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.FallbackInfura.GetParameters() {
				if param.Name == label {
					box.SetText(fmt.Sprint(param.Value))
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.FallbackInfura.ProjectID.Value = text[projectIdLabel]
		wiz.consensusModeModal.show()
	}

	back := func() {
		wiz.fallbackExecutionModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		70,
		"Fallback Execution Client > Infura",
		[]string{projectIdLabel},
		show,
		done,
		back,
		"step-fallback-ec-infura",
	)

}
