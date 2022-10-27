package config

import "fmt"

func createNativeDataStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	dataPathLabel := wiz.md.Config.Smartnode.DataPath.Name
	vrcLabel := wiz.md.Config.Native.ValidatorRestartCommand.Name
	vscLabel := wiz.md.Config.Native.ValidatorStopCommand.Name

	helperText := "Please enter the path of your `data` directory.\nThis folder holds your wallet and password files, and your validator key folder.\n\nAlso enter the path of the restart and stop scripts which will restart or stop your validator container if the Smartnode detects a configuration change or issue."

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.Native.GetParameters() {
				if param.Name == label {
					box.SetText(fmt.Sprint(param.Value))
				}
			}
			for _, param := range wiz.md.Config.Smartnode.GetParameters() {
				if param.Name == label {
					box.SetText(fmt.Sprint(param.Value))
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.Smartnode.DataPath.Value = text[dataPathLabel]
		wiz.md.Config.Native.ValidatorRestartCommand.Value = text[vrcLabel]
		wiz.md.Config.Native.ValidatorStopCommand.Value = text[vscLabel]
		wiz.nativeUseFallbackModal.show()
	}

	back := func() {
		wiz.nativeCcUrlModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		96,
		"Other Settings",
		[]string{dataPathLabel, vrcLabel, vscLabel},
		[]int{wiz.md.Config.Smartnode.DataPath.MaxLength, wiz.md.Config.Native.ValidatorRestartCommand.MaxLength, wiz.md.Config.Native.ValidatorStopCommand.MaxLength},
		[]string{wiz.md.Config.Smartnode.DataPath.Regex, wiz.md.Config.Native.ValidatorRestartCommand.Regex, wiz.md.Config.Native.ValidatorStopCommand.Regex},
		show,
		done,
		back,
		"step-native-data",
	)

}
