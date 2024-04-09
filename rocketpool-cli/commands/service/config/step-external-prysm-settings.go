package config

func createExternalPrysmSettingsStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {
	// Create the labels
	httpUrlLabel := wiz.md.Config.ExternalBeaconClient.HttpUrl.Name
	jsonRpcUrlLabel := wiz.md.Config.ExternalBeaconClient.PrysmRpcUrl.Name

	helperText := "Please provide the URL of your Prysm client's HTTP API (for example: `http://192.168.1.40:5052`) and the URL of its JSON RPC API (e.g., `192.168.1.40:5053`) too.\n\nNote that if you're running it on the same machine as the Smart Node, you cannot use `localhost` or `127.0.0.1`; you must use your machine's LAN IP address."

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
		wiz.md.Config.ExternalBeaconClient.HttpUrl.Value = text[httpUrlLabel]
		wiz.md.Config.ExternalBeaconClient.PrysmRpcUrl.Value = text[jsonRpcUrlLabel]
		wiz.graffitiModal.show()
	}

	back := func() {
		wiz.externalBnSelectModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		70,
		"Beacon Node (External) > Settings",
		[]string{httpUrlLabel, jsonRpcUrlLabel},
		[]int{wiz.md.Config.ExternalBeaconClient.HttpUrl.MaxLength, wiz.md.Config.ExternalBeaconClient.PrysmRpcUrl.MaxLength},
		[]string{wiz.md.Config.ExternalBeaconClient.HttpUrl.Regex, wiz.md.Config.ExternalBeaconClient.PrysmRpcUrl.Regex},
		show,
		done,
		back,
		"step-external-prysm-settings",
	)
}
