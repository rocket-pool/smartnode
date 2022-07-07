package config

import (
	"fmt"

	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func createExternalEcStep(wiz *wizard, currentStep int, totalSteps int) *textBoxWizardStep {

	// Create the labels
	httpLabel := wiz.md.Config.ExternalExecution.HttpUrl.Name
	wsLabel := wiz.md.Config.ExternalExecution.WsUrl.Name

	helperText := "Please enter the URL of the HTTP-based RPC API and the URL of the Websocket-based RPC API for your existing client.\n\nFor example: `http://192.168.1.45:8545` and `ws://192.168.1.45:8546`"

	show := func(modal *textBoxModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus()
		for label, box := range modal.textboxes {
			for _, param := range wiz.md.Config.ExternalExecution.GetParameters() {
				if param.Name == label {
					box.SetText(fmt.Sprint(param.Value))
				}
			}
		}
	}

	done := func(text map[string]string) {
		wiz.md.Config.ExternalExecution.HttpUrl.Value = text[httpLabel]
		wiz.md.Config.ExternalExecution.WsUrl.Value = text[wsLabel]
		if wiz.md.Config.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
			wiz.consensusLocalModal.show()
		} else {
			wiz.consensusExternalSelectModal.show()
		}
	}

	back := func() {
		wiz.modeModal.show()
	}

	return newTextBoxWizardStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		70,
		"Execution Client (External)",
		[]string{httpLabel, wsLabel},
		[]int{wiz.md.Config.ExternalExecution.HttpUrl.MaxLength, wiz.md.Config.ExternalExecution.WsUrl.MaxLength},
		[]string{wiz.md.Config.ExternalExecution.HttpUrl.Regex, wiz.md.Config.ExternalExecution.WsUrl.Regex},
		show,
		done,
		back,
		"step-ec-external",
	)

}
