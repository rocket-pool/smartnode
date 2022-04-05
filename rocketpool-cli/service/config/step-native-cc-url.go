package config

import "fmt"

func createNativeCcUrlStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	httpLabel := wiz.md.Config.Native.CcHttpUrl.Name

	helperText := "Please enter the URL of the HTTP-based API for your Consensus client.\n\nFor example: `http://127.0.0.1:5052`"

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.Native.GetParameters() {
				if param.Name == label {
					box.SetText(fmt.Sprint(param.Value))
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.Native.CcHttpUrl.Value = text[httpLabel]
		wiz.nativeDataModal.show()
	}

	back := func() {
		wiz.nativeCcModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		84,
		"Consensus Client > URL",
		[]string{httpLabel},
		[]int{wiz.md.Config.Native.CcHttpUrl.MaxLength},
		[]string{wiz.md.Config.Native.CcHttpUrl.Regex},
		show,
		done,
		back,
		"step-native-cc-url",
	)

}
