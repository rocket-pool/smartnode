package config

func createFallbackPrysmStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {
	// Create the labels
	ecHttpLabel := wiz.md.Config.Fallback.EcHttpUrl.Name
	bnHttpLabel := wiz.md.Config.Fallback.BnHttpUrl.Name
	jsonRpcLabel := wiz.md.Config.Fallback.PrysmRpcUrl.Name

	helperText := "[orange]NOTE: you have selected Prysm as your primary Beacon Node.\n**Make sure your fallback is also running Prysm, or it will not be able to connect.**\n\n[white]Please enter the URLs of the HTTP APIs for your fallback clients. For example: `http://192.168.1.45:8545` for your Execution client and `http://192.168.1.45:5052` for your fallback Prysm node. You will also need to provide the JSON-RPC URL for your fallback Prysm node."

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
		wiz.md.Config.Fallback.PrysmRpcUrl.Value = text[jsonRpcLabel]
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
		[]string{ecHttpLabel, bnHttpLabel, jsonRpcLabel},
		[]int{wiz.md.Config.Fallback.EcHttpUrl.MaxLength, wiz.md.Config.Fallback.BnHttpUrl.MaxLength, wiz.md.Config.Fallback.PrysmRpcUrl.MaxLength},
		[]string{wiz.md.Config.Fallback.EcHttpUrl.Regex, wiz.md.Config.Fallback.BnHttpUrl.Regex, wiz.md.Config.Fallback.PrysmRpcUrl.Regex},
		show,
		done,
		back,
		"step-fallback-prysm",
	)
}
