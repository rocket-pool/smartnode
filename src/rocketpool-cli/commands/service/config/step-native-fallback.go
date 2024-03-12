package config

import (
	"fmt"
)

func createNativeFallbackStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	ecHttpLabel := wiz.md.Config.FallbackNormal.EcHttpUrl.Name
	ccHttpLabel := wiz.md.Config.FallbackNormal.CcHttpUrl.Name

	helperText := "Please enter the URLs of the HTTP APIs for your fallback clients.\n\nFor example: `http://192.168.1.45:8545` for your Execution client and `http://192.168.1.45:5052` for your Consensus client."

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.FallbackNormal.GetParameters() {
				if param.Name == label {
					box.SetText(fmt.Sprint(param.Value))
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.FallbackNormal.EcHttpUrl.Value = text[ecHttpLabel]
		wiz.md.Config.FallbackNormal.CcHttpUrl.Value = text[ccHttpLabel]
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
		"Fallback Client URLs",
		[]string{ecHttpLabel, ccHttpLabel},
		[]int{wiz.md.Config.FallbackNormal.EcHttpUrl.MaxLength, wiz.md.Config.FallbackNormal.CcHttpUrl.MaxLength},
		[]string{wiz.md.Config.FallbackNormal.EcHttpUrl.Regex, wiz.md.Config.FallbackNormal.CcHttpUrl.Regex},
		show,
		done,
		back,
		"step-native-fallback",
	)

}
