package config

import (
	"fmt"
)

func createFallbackPrysmStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	ecHttpLabel := wiz.md.Config.FallbackPrysm.EcHttpUrl.Name
	ccHttpLabel := wiz.md.Config.FallbackPrysm.CcHttpUrl.Name
	jsonRpcLabel := wiz.md.Config.FallbackPrysm.JsonRpcUrl.Name

	helperText := "[orange]NOTE: you have selected Prysm as your primary Consensus client.\n**Make sure your fallback Consensus client is also running Prysm, or it will not be able to connect.**\n\n[white]Please enter the URLs of the HTTP APIs for your fallback clients. For example: `http://192.168.1.45:8545` for your Execution client and `http://192.168.1.45:5052` for your fallback Prysm node. You will also need to provide the JSON-RPC URL for your fallback Prysm node."

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
		wiz.md.Config.FallbackPrysm.EcHttpUrl.Value = text[ecHttpLabel]
		wiz.md.Config.FallbackPrysm.CcHttpUrl.Value = text[ccHttpLabel]
		wiz.md.Config.FallbackPrysm.JsonRpcUrl.Value = text[jsonRpcLabel]
		wiz.metricsModal.show()
	}

	back := func() {
		wiz.useFallbackModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		96,
		"Fallback Client URLs",
		[]string{ecHttpLabel, ccHttpLabel, jsonRpcLabel},
		[]int{wiz.md.Config.FallbackPrysm.EcHttpUrl.MaxLength, wiz.md.Config.FallbackPrysm.CcHttpUrl.MaxLength, wiz.md.Config.FallbackPrysm.JsonRpcUrl.MaxLength},
		[]string{wiz.md.Config.FallbackPrysm.EcHttpUrl.Regex, wiz.md.Config.FallbackPrysm.CcHttpUrl.Regex, wiz.md.Config.FallbackPrysm.JsonRpcUrl.Regex},
		show,
		done,
		back,
		"step-fallback-prysm",
	)

}
