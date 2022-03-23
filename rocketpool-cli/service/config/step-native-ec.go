package config

import "fmt"

func createNativeEcStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	httpLabel := wiz.md.Config.Native.EcHttpUrl.Name

	helperText := "Please enter the URL of the HTTP-based RPC API for your Execution client (e.g. Geth).\n\nFor example: `http://127.0.0.1:8545`"

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
		wiz.md.Config.Native.EcHttpUrl.Value = text[httpLabel]
		wiz.nativeCcModal.show()
	}

	back := func() {
		wiz.nativeNetworkModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		84,
		"Execution Client > URL",
		[]string{httpLabel},
		[]int{wiz.md.Config.Native.EcHttpUrl.MaxLength},
		[]string{wiz.md.Config.Native.EcHttpUrl.Regex},
		show,
		done,
		back,
		"step-native-ec",
	)

}
