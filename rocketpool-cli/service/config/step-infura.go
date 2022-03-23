package config

import (
	"fmt"
)

func createInfuraStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	projectIdLabel := wiz.md.Config.Infura.ProjectID.Name

	helperText := "Please enter the Project ID for your Infura Ethereum project. You can find this on the Infura website, in your Ethereum project settings."

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.Infura.GetParameters() {
				if param.Name == label {
					box.SetText(fmt.Sprint(param.Value))
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.Infura.ProjectID.Value = text[projectIdLabel]
		wiz.fallbackExecutionModal.show()
	}

	back := func() {
		wiz.executionLocalModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		70,
		"Execution Client > Infura",
		[]string{projectIdLabel},
		[]int{wiz.md.Config.Infura.ProjectID.MaxLength},
		[]string{wiz.md.Config.Infura.ProjectID.Regex},
		show,
		done,
		back,
		"step-ec-infura",
	)

}
