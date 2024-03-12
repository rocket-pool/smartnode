package config

import (
	"fmt"
)

func createExternalTekuStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	httpUrlLabel := wiz.md.Config.ExternalTeku.HttpUrl.Name

	helperText := "Please provide the URL of your Teku client's HTTP API (for example: `http://192.168.1.40:5052`).\n\nNote that if you're running it on the same machine as the Smartnode, you cannot use `localhost` or `127.0.0.1`; you must use your machine's LAN IP address."

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.ExternalTeku.GetParameters() {
				if param.Name == label {
					box.SetText(fmt.Sprint(param.Value))
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.ExternalTeku.HttpUrl.Value = text[httpUrlLabel]
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
		[]string{httpUrlLabel},
		[]int{wiz.md.Config.ExternalTeku.HttpUrl.MaxLength},
		[]string{wiz.md.Config.ExternalTeku.HttpUrl.Regex},
		show,
		done,
		back,
		"step-external-teku",
	)

}
