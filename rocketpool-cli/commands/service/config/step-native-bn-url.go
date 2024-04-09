package config

func createNativeBnUrlStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {
	// Create the labels
	httpLabel := wiz.md.Config.ExternalBeaconClient.HttpUrl.Name

	helperText := "Please enter the URL of the HTTP-based API for your Beacon Node.\n\nFor example: `http://127.0.0.1:5052`"

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.ExternalBeaconClient.GetParameters() {
				if param.GetCommon().Name == label {
					box.SetText(param.String())
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.ExternalBeaconClient.HttpUrl.Value = text[httpLabel]
		wiz.nativeUseFallbackModal.show()
	}

	back := func() {
		wiz.nativeBnModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		84,
		"Beacon Node > URL",
		[]string{httpLabel},
		[]int{wiz.md.Config.ExternalBeaconClient.HttpUrl.MaxLength},
		[]string{wiz.md.Config.ExternalBeaconClient.HttpUrl.Regex},
		show,
		done,
		back,
		"step-native-bn-url",
	)
}
