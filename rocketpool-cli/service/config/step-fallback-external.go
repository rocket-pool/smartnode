package config

import (
	"fmt"
)

func createFallbackExternalEcStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	httpLabel := wiz.md.Config.FallbackExecution.HttpUrl.Name
	wsLabel := wiz.md.Config.FallbackExecution.WsUrl.Name

	helperText := "Please enter the URL of the HTTP-based RPC API and the URL of the Websocket-based RPC API for your existing client.\n\nFor example: `http://192.168.1.45:8545` and `ws://192.168.1.45:8546`"

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.FallbackExecution.GetParameters() {
				if param.Name == label {
					box.SetText(fmt.Sprint(param.Value))
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.FallbackExecution.HttpUrl.Value = text[httpLabel]
		wiz.md.Config.FallbackExecution.WsUrl.Value = text[wsLabel]
		wiz.consensusModeModal.show()
	}

	back := func() {
		wiz.fallbackExecutionModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		70,
		"Fallback Execution Client (External)",
		[]string{httpLabel, wsLabel},
		[]int{wiz.md.Config.FallbackExecution.HttpUrl.MaxLength, wiz.md.Config.FallbackExecution.WsUrl.MaxLength},
		[]string{wiz.md.Config.FallbackExecution.HttpUrl.Regex, wiz.md.Config.FallbackExecution.WsUrl.Regex},
		show,
		done,
		back,
		"step-fallback-ec-external",
	)

}
