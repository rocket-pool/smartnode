package config

import "github.com/rocket-pool/smartnode/shared/types/config"

func createMetricsStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "Would you like to enable the Smartnode's metrics monitoring system? This will monitor things such as hardware stats (CPU usage, RAM usage, free disk space), your minipool stats, stats about your node such as total RPL and ETH rewards, and much more. It also enables the Grafana dashboard to quickly and easily view these metrics (see https://docs.rocketpool.net/guides/node/grafana.html for an example).\n\nNone of this information will be sent to any remote servers for collection an analysis; this is purely for your own usage on your node."

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
			wiz.finishedModal.show()
		} else {
			wiz.mevModeModal.show()
		}
	}

	back := func() {
		wiz.useFallbackModal.show()
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
		"step-metrics",
	)

}
