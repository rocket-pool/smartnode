package config

func createNativeFallbackStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {
	// Create the labels
	ecHttpLabel := wiz.md.Config.Fallback.EcHttpUrl.Name
	bnHttpLabel := wiz.md.Config.Fallback.BnHttpUrl.Name

	helperText := "Please enter the URLs of the HTTP APIs for your fallback clients.\n\nFor example: `http://192.168.1.45:8545` for your Execution Client and `http://192.168.1.45:5052` for your Beacon Node."

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.Fallback.GetParameters() {
				if param.GetCommon().Name == label {
					box.SetText(param.String())
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.Fallback.EcHttpUrl.Value = text[ecHttpLabel]
		wiz.md.Config.Fallback.BnHttpUrl.Value = text[bnHttpLabel]
		wiz.nativeDataModal.show()
	}

	back := func() {
		wiz.nativeUseFallbackModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		96,
		"Fallback Client URLs",
		[]string{ecHttpLabel, bnHttpLabel},
		[]int{wiz.md.Config.Fallback.EcHttpUrl.MaxLength, wiz.md.Config.Fallback.BnHttpUrl.MaxLength},
		[]string{wiz.md.Config.Fallback.EcHttpUrl.Regex, wiz.md.Config.Fallback.BnHttpUrl.Regex},
		show,
		done,
		back,
		"step-native-fallback",
	)
}
