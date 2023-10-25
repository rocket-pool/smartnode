package config

import "github.com/rocket-pool/smartnode/shared/types/config"

func createNativeMetricsStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "Would you like to enable the daemon's metrics feature? This will allow you to access the Rocket Pool network's metrics and the metrics for your own node wallet in the Grafana dashboard."

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		if wiz.md.Config.EnableMetrics.Value == false {
			modal.focus(0)
		} else {
			modal.focus(1)
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 1 {
			wiz.md.Config.EnableMetrics.Value = true
		} else {
			wiz.md.Config.EnableMetrics.Value = false
		}
		if wiz.md.Config.Smartnode.Network.Value == config.Network_Holesky || wiz.md.Config.Smartnode.Network.Value == config.Network_Devnet {
			// Skip MEV for Holesky
			wiz.nativeFinishedModal.show()
		} else {
			wiz.nativeMevModal.show()
		}
	}

	back := func() {
		wiz.nativeUseFallbackModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"No", "Yes"},
		[]string{},
		76,
		"Metrics",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-native-metrics",
	)

}
