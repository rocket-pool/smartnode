package config

func createNativeEcUrlStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {
	// Create the labels
	httpLabel := wiz.md.Config.ExternalExecutionClient.HttpUrl.Name

	helperText := "Please enter the URL of the HTTP-based RPC API for your Execution Client.\n\nFor example: `http://127.0.0.1:8545`"

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.ExternalExecutionClient.GetParameters() {
				if param.GetCommon().Name == label {
					box.SetText(param.String())
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.ExternalExecutionClient.HttpUrl.Value = text[httpLabel]
		wiz.nativeBnModal.show()
	}

	back := func() {
		wiz.nativeEcModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		84,
		"Execution Client > URL",
		[]string{httpLabel},
		[]int{wiz.md.Config.ExternalExecutionClient.HttpUrl.MaxLength},
		[]string{wiz.md.Config.ExternalExecutionClient.HttpUrl.Regex},
		show,
		done,
		back,
		"step-native-ec",
	)
}
