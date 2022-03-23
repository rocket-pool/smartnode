package config

type wizard struct {
	md *mainDisplay

	// Docker mode
	welcomeModal                    *choiceWizardStep
	networkModal                    *choiceWizardStep
	executionModeModal              *choiceWizardStep
	executionLocalModal             *choiceWizardStep
	executionExternalModal          *textBoxWizardStep
	infuraModal                     *textBoxWizardStep
	fallbackInfuraModal             *textBoxWizardStep
	fallbackExecutionModal          *choiceWizardStep
	consensusModeModal              *choiceWizardStep
	consensusLocalModal             *choiceWizardStep
	consensusExternalSelectModal    *choiceWizardStep
	graffitiModal                   *textBoxWizardStep
	checkpointSyncProviderModal     *textBoxWizardStep
	doppelgangerDetectionModal      *choiceWizardStep
	lighthouseExternalSettingsModal *textBoxWizardStep
	prysmExternalSettingsModal      *textBoxWizardStep
	tekuExternalSettingsModal       *textBoxWizardStep
	externalGraffitiModal           *textBoxWizardStep
	metricsModal                    *choiceWizardStep
	finishedModal                   *choiceWizardStep
	consensusLocalRandomModal       *choiceWizardStep
	consensusLocalRandomPrysmModal  *choiceWizardStep
	consensusLocalPrysmWarning      *choiceWizardStep
	consensusLocalTekuWarning       *choiceWizardStep
	externalDoppelgangerModal       *choiceWizardStep

	// Native mode
	nativeWelcomeModal  *choiceWizardStep
	nativeNetworkModal  *choiceWizardStep
	nativeEcModal       *textBoxWizardStep
	nativeCcModal       *choiceWizardStep
	nativeCcUrlModal    *textBoxWizardStep
	nativeDataModal     *textBoxWizardStep
	nativeMetricsModal  *choiceWizardStep
	nativeFinishedModal *choiceWizardStep
}

func newWizard(md *mainDisplay) *wizard {

	wiz := &wizard{
		md: md,
	}

	totalDockerSteps := 9
	totalNativeSteps := 8

	// Docker mode
	wiz.welcomeModal = createWelcomeStep(wiz, 1, totalDockerSteps)
	wiz.networkModal = createNetworkStep(wiz, 2, totalDockerSteps)
	wiz.executionModeModal = createEcModeStep(wiz, 3, totalDockerSteps)
	wiz.executionLocalModal = createLocalEcStep(wiz, 4, totalDockerSteps)
	wiz.executionExternalModal = createExternalEcStep(wiz, 4, totalDockerSteps)
	wiz.infuraModal = createInfuraStep(wiz, 4, totalDockerSteps)
	wiz.fallbackExecutionModal = createFallbackEcStep(wiz, 5, totalDockerSteps)
	wiz.fallbackInfuraModal = createFallbackInfuraStep(wiz, 5, totalDockerSteps)
	wiz.consensusModeModal = createCcModeStep(wiz, 6, totalDockerSteps)
	wiz.consensusExternalSelectModal = createExternalCcStep(wiz, 7, totalDockerSteps)
	wiz.consensusLocalPrysmWarning = createPrysmWarningStep(wiz, 7, totalDockerSteps)
	wiz.consensusLocalTekuWarning = createTekuWarningStep(wiz, 7, totalDockerSteps)
	wiz.graffitiModal = createGraffitiStep(wiz, 7, totalDockerSteps)
	wiz.checkpointSyncProviderModal = createCheckpointSyncStep(wiz, 7, totalDockerSteps)
	wiz.doppelgangerDetectionModal = createDoppelgangerStep(wiz, 7, totalDockerSteps)
	wiz.lighthouseExternalSettingsModal = createExternalLhStep(wiz, 7, totalDockerSteps)
	wiz.prysmExternalSettingsModal = createExternalPrysmStep(wiz, 7, totalDockerSteps)
	wiz.tekuExternalSettingsModal = createExternalTekuStep(wiz, 7, totalDockerSteps)
	wiz.externalGraffitiModal = createExternalGraffitiStep(wiz, 7, totalDockerSteps)
	wiz.externalDoppelgangerModal = createExternalDoppelgangerStep(wiz, 7, totalDockerSteps)
	wiz.metricsModal = createMetricsStep(wiz, 8, totalDockerSteps)
	wiz.finishedModal = createFinishedStep(wiz, 9, totalDockerSteps)

	// Native mode
	wiz.nativeWelcomeModal = createNativeWelcomeStep(wiz, 1, totalNativeSteps)
	wiz.nativeNetworkModal = createNativeNetworkStep(wiz, 2, totalNativeSteps)
	wiz.nativeEcModal = createNativeEcStep(wiz, 3, totalNativeSteps)
	wiz.nativeCcModal = createNativeCcStep(wiz, 4, totalNativeSteps)
	wiz.nativeCcUrlModal = createNativeCcUrlStep(wiz, 5, totalNativeSteps)
	wiz.nativeDataModal = createNativeDataStep(wiz, 6, totalNativeSteps)
	wiz.nativeMetricsModal = createNativeMetricsStep(wiz, 7, totalNativeSteps)
	wiz.nativeFinishedModal = createNativeFinishedStep(wiz, 8, totalNativeSteps)

	return wiz

}
