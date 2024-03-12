package config

func createNativeMevStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	helperText := "NOTE: Rocket Pool expects our node operators to run MEV-Boost to prevent cheating and gain supplemental rewards from block proposals. In Native mode, you are responsible for setting up MEV-Boost yourself.\nMEV-Boost is currently opt-out, [orange]but will become required for all node operators in the future. [white]Please set MEV-Boost up when you are done configuring the Smartnode unless you explicitly intend to opt-out for now.\n\n[lime]Please read our guide to learn more about MEV:\nhttps://docs.rocketpool.net/guides/node/mev.html"

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0)
	}

	done := func(buttonIndex int, buttonLabel string) {
		wiz.nativeFinishedModal.show()
	}

	back := func() {
		wiz.nativeMetricsModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		[]string{"Ok"},
		[]string{},
		90,
		"MEV-Boost",
		DirectionalModalHorizontal,
		show,
		done,
		back,
		"step-native-mev",
	)

}
