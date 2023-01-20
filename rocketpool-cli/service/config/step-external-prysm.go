package config

import (
	"fmt"
)

func createExternalPrysmStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	httpUrlLabel := wiz.md.Config.ExternalPrysm.HttpUrl.Name
	jsonRpcUrlLabel := wiz.md.Config.ExternalPrysm.JsonRpcUrl.Name

	helperText := "Please provide the URL of your Prysm client's HTTP API (for example: `http://192.168.1.40:5052`) and the URL of its JSON RPC API (e.g., `192.168.1.40:5053`).\n\nNote that if you're running it on the same machine as the Smartnode, you cannot use `localhost` or `127.0.0.1`; you must use your machine's LAN IP address."

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.ExternalPrysm.GetParameters() {
				if param.Name == label {
					box.SetText(fmt.Sprint(param.Value))
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.ExternalPrysm.HttpUrl.Value = text[httpUrlLabel]
		wiz.md.Config.ExternalPrysm.JsonRpcUrl.Value = text[jsonRpcUrlLabel]
		wiz.externalGraffitiModal.show()
	}

	back := func() {
		wiz.consensusExternalSelectModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		70,
		"Consensus Client (External) > Settings",
		[]string{httpUrlLabel, jsonRpcUrlLabel},
		[]int{wiz.md.Config.ExternalPrysm.HttpUrl.MaxLength, wiz.md.Config.ExternalPrysm.JsonRpcUrl.MaxLength},
		[]string{wiz.md.Config.ExternalPrysm.HttpUrl.Regex, wiz.md.Config.ExternalPrysm.JsonRpcUrl.Regex},
		show,
		done,
		back,
		"step-external-prysm",
	)

}
