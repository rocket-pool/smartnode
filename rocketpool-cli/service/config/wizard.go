package config

type wizard struct {
	md *mainDisplay

	// Docker mode
	welcomeModal                    *choiceWizardStep
	networkModal                    *choiceWizardStep
	modeModal                       *choiceWizardStep
	executionLocalModal             *choiceWizardStep
	executionExternalModal          *textBoxWizardStep
	consensusLocalModal             *choiceWizardStep
	consensusExternalSelectModal    *choiceWizardStep
	graffitiModal                   *textBoxWizardStep
	checkpointSyncProviderModal     *textBoxWizardStep
	doppelgangerDetectionModal      *choiceWizardStep
	lighthouseExternalSettingsModal *textBoxWizardStep
	nimbusExternalSettingsModal     *textBoxWizardStep
	lodestarExternalSettingsModal   *textBoxWizardStep
	prysmExternalSettingsModal      *textBoxWizardStep
	tekuExternalSettingsModal       *textBoxWizardStep
	externalGraffitiModal           *textBoxWizardStep
	metricsModal                    *choiceWizardStep
	mevModeModal                    *choiceWizardStep
	localMevModal                   *checkBoxWizardStep
	externalMevModal                *textBoxWizardStep
	finishedModal                   *choiceWizardStep
	consensusLocalRandomModal       *choiceWizardStep
	consensusLocalRandomPrysmModal  *choiceWizardStep
	consensusLocalPrysmWarning      *choiceWizardStep
	consensusLocalTekuWarning       *choiceWizardStep
	externalDoppelgangerModal       *choiceWizardStep
	executionLocalRandomModal       *choiceWizardStep
	useFallbackModal                *choiceWizardStep
	fallbackNormalModal             *textBoxWizardStep
	fallbackPrysmModal              *textBoxWizardStep

	// Native mode
	nativeWelcomeModal     *choiceWizardStep
	nativeNetworkModal     *choiceWizardStep
	nativeEcModal          *textBoxWizardStep
	nativeCcModal          *choiceWizardStep
	nativeCcUrlModal       *textBoxWizardStep
	nativeDataModal        *textBoxWizardStep
	nativeUseFallbackModal *choiceWizardStep
	nativeFallbackModal    *textBoxWizardStep
	nativeMetricsModal     *choiceWizardStep
	nativeMevModal         *choiceWizardStep
	nativeFinishedModal    *choiceWizardStep
}

func newWizard(md *mainDisplay) *wizard {

	wiz := &wizard{
		md: md,
	}

	totalDockerSteps := 9
	totalNativeSteps := 10

	// Docker mode
	wiz.welcomeModal = createWelcomeStep(wiz, 1, totalDockerSteps)
	wiz.networkModal = createNetworkStep(wiz, 2, totalDockerSteps)
	wiz.modeModal = createModeStep(wiz, 3, totalDockerSteps)
	wiz.executionLocalModal = createLocalEcStep(wiz, 4, totalDockerSteps)
	wiz.executionExternalModal = createExternalEcStep(wiz, 4, totalDockerSteps)
	wiz.consensusLocalModal = createLocalCcStep(wiz, 5, totalDockerSteps)
	wiz.consensusExternalSelectModal = createExternalCcStep(wiz, 5, totalDockerSteps)
	wiz.consensusLocalPrysmWarning = createPrysmWarningStep(wiz, 5, totalDockerSteps)
	wiz.consensusLocalTekuWarning = createTekuWarningStep(wiz, 5, totalDockerSteps)
	wiz.graffitiModal = createGraffitiStep(wiz, 5, totalDockerSteps)
	wiz.checkpointSyncProviderModal = createCheckpointSyncStep(wiz, 5, totalDockerSteps)
	wiz.doppelgangerDetectionModal = createDoppelgangerStep(wiz, 5, totalDockerSteps)
	wiz.lighthouseExternalSettingsModal = createExternalLhStep(wiz, 5, totalDockerSteps)
	wiz.nimbusExternalSettingsModal = createExternalNimbusStep(wiz, 5, totalDockerSteps)
	wiz.lodestarExternalSettingsModal = createExternalLodestarStep(wiz, 5, totalDockerSteps)
	wiz.prysmExternalSettingsModal = createExternalPrysmStep(wiz, 5, totalDockerSteps)
	wiz.tekuExternalSettingsModal = createExternalTekuStep(wiz, 5, totalDockerSteps)
	wiz.externalGraffitiModal = createExternalGraffitiStep(wiz, 5, totalDockerSteps)
	wiz.externalDoppelgangerModal = createExternalDoppelgangerStep(wiz, 5, totalDockerSteps)
	wiz.useFallbackModal = createUseFallbackStep(wiz, 6, totalDockerSteps)
	wiz.fallbackNormalModal = createFallbackNormalStep(wiz, 6, totalDockerSteps)
	wiz.fallbackPrysmModal = createFallbackPrysmStep(wiz, 6, totalDockerSteps)
	wiz.metricsModal = createMetricsStep(wiz, 7, totalDockerSteps)
	wiz.mevModeModal = createMevModeStep(wiz, 8, totalDockerSteps)
	wiz.localMevModal = createLocalMevStep(wiz, 8, totalDockerSteps)
	wiz.externalMevModal = createExternalMevStep(wiz, 8, totalDockerSteps)
	wiz.finishedModal = createFinishedStep(wiz, 9, totalDockerSteps)

	// Native mode
	wiz.nativeWelcomeModal = createNativeWelcomeStep(wiz, 1, totalNativeSteps)
	wiz.nativeNetworkModal = createNativeNetworkStep(wiz, 2, totalNativeSteps)
	wiz.nativeEcModal = createNativeEcStep(wiz, 3, totalNativeSteps)
	wiz.nativeCcModal = createNativeCcStep(wiz, 4, totalNativeSteps)
	wiz.nativeCcUrlModal = createNativeCcUrlStep(wiz, 5, totalNativeSteps)
	wiz.nativeDataModal = createNativeDataStep(wiz, 6, totalNativeSteps)
	wiz.nativeUseFallbackModal = createNativeUseFallbackStep(wiz, 7, totalNativeSteps)
	wiz.nativeFallbackModal = createNativeFallbackStep(wiz, 7, totalNativeSteps)
	wiz.nativeMetricsModal = createNativeMetricsStep(wiz, 8, totalNativeSteps)
	wiz.nativeMevModal = createNativeMevStep(wiz, 9, totalNativeSteps)
	wiz.nativeFinishedModal = createNativeFinishedStep(wiz, 10, totalNativeSteps)

	return wiz

}
