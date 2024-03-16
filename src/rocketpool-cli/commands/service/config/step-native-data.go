package config

func createNativeDataStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {
	// Create the labels
	dataPathLabel := wiz.md.Config.UserDataPath.Name
	vrcLabel := wiz.md.Config.ValidatorClient.NativeValidatorRestartCommand.Name
	vscLabel := wiz.md.Config.ValidatorClient.NativeValidatorStopCommand.Name

	helperText := "Please enter the path of your 'data' directory.\nThis folder holds your wallet and password files, and your validator keys.\n\nAlso enter the path of the restart and stop scripts which will restart or stop your Validator Client if the Smart Node detects a configuration change or issue."

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.ValidatorClient.GetParameters() {
				if param.GetCommon().Name == label {
					box.SetText(param.String())
				}
			}
			for _, param := range wiz.md.Config.GetParameters() {
				if param.GetCommon().Name == label {
					box.SetText(param.String())
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.UserDataPath.Value = text[dataPathLabel]
		wiz.md.Config.ValidatorClient.NativeValidatorRestartCommand.Value = text[vrcLabel]
		wiz.md.Config.ValidatorClient.NativeValidatorStopCommand.Value = text[vscLabel]
		wiz.nativeMetricsModal.show()
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
		"Other Settings",
		[]string{dataPathLabel, vrcLabel, vscLabel},
		[]int{wiz.md.Config.UserDataPath.MaxLength, wiz.md.Config.ValidatorClient.NativeValidatorRestartCommand.MaxLength, wiz.md.Config.ValidatorClient.NativeValidatorStopCommand.MaxLength},
		[]string{wiz.md.Config.UserDataPath.Regex, wiz.md.Config.ValidatorClient.NativeValidatorRestartCommand.Regex, wiz.md.Config.ValidatorClient.NativeValidatorStopCommand.Regex},
		show,
		done,
		back,
		"step-native-data",
	)
}
