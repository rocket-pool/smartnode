package config

type wizard struct {
	md *mainDisplay

	// ===================
	// === Docker Mode ===
	// ===================

	// Step 1 - Welcome
	welcomeModal *choiceWizardStep

	// Step 2 - Network
	networkModal *choiceWizardStep

	// Step 3 - Client mode
	modeModal *choiceWizardStep

	// Step 4 - EC settings
	localEcModal            *choiceWizardStep
	localEcRandomModal      *choiceWizardStep
	externalEcSelectModal   *choiceWizardStep
	externalEcSettingsModal *textBoxWizardStep

	// Step 5 - BN settings
	localBnModal                *choiceWizardStep
	localBnRandomModal          *choiceWizardStep
	localBnPrysmWarning         *choiceWizardStep
	localBnTekuWarning          *choiceWizardStep
	checkpointSyncProviderModal *textBoxWizardStep
	externalBnSelectModal       *choiceWizardStep
	externalBnSettingsModal     *textBoxWizardStep
	externalPrysmSettingsModal  *textBoxWizardStep

	// Step 6 - VC settings
	graffitiModal              *textBoxWizardStep
	doppelgangerDetectionModal *choiceWizardStep

	// Step 7 - Fallback clients
	useFallbackModal    *choiceWizardStep
	fallbackNormalModal *textBoxWizardStep
	fallbackPrysmModal  *textBoxWizardStep

	// Step 8 - Metrics
	metricsModal *choiceWizardStep

	// Step 9 - MEV Boost
	mevModeModal     *choiceWizardStep
	localMevModal    *checkBoxWizardStep
	externalMevModal *textBoxWizardStep

	// Done
	finishedModal *choiceWizardStep

	// ===================
	// === Native Mode ===
	// ===================

	// Step 1 - Welcome
	nativeWelcomeModal *choiceWizardStep

	// Step 2 - Network
	nativeNetworkModal *choiceWizardStep

	// Step 3 - EC settings
	nativeEcModal    *choiceWizardStep
	nativeEcUrlModal *textBoxWizardStep

	// Step 4 - BN settings
	nativeBnModal    *choiceWizardStep
	nativeBnUrlModal *textBoxWizardStep

	// Step 5 - Fallback clients
	nativeUseFallbackModal *choiceWizardStep
	nativeFallbackModal    *textBoxWizardStep

	// Step 6 - Native stuff
	nativeDataModal *textBoxWizardStep

	// Step 7 - Metrics
	nativeMetricsModal *choiceWizardStep

	// Step 8 - MEV Boost
	nativeMevModal *choiceWizardStep

	// Done
	nativeFinishedModal *choiceWizardStep
}

func newWizard(md *mainDisplay) *wizard {
	wiz := &wizard{
		md: md,
	}

	// ===================
	// === Docker Mode ===
	// ===================
	totalDockerSteps := 10
	stepCount := 0

	// Step 1 - Welcome
	wiz.welcomeModal = createWelcomeStep(wiz, stepCount, totalDockerSteps)
	stepCount++

	// Step 2 - Network
	wiz.networkModal = createNetworkStep(wiz, stepCount, totalDockerSteps)
	stepCount++

	// Step 3 - Client mode
	wiz.modeModal = createModeStep(wiz, stepCount, totalDockerSteps)
	stepCount++

	// Step 4 - EC settings
	wiz.localEcModal = createLocalEcStep(wiz, stepCount, totalDockerSteps)
	wiz.externalEcSelectModal = createExternalEcSelectStep(wiz, stepCount, totalDockerSteps)
	wiz.externalEcSettingsModal = createExternalEcStep(wiz, stepCount, totalDockerSteps)
	stepCount++

	// Step 5 - BN settings
	wiz.localBnModal = createLocalCcStep(wiz, stepCount, totalDockerSteps)
	wiz.localBnPrysmWarning = createPrysmWarningStep(wiz, stepCount, totalDockerSteps)
	wiz.localBnTekuWarning = createTekuWarningStep(wiz, stepCount, totalDockerSteps)
	wiz.checkpointSyncProviderModal = createCheckpointSyncStep(wiz, stepCount, totalDockerSteps)
	wiz.externalBnSelectModal = createExternalBnStep(wiz, stepCount, totalDockerSteps)
	wiz.externalBnSettingsModal = createExternalBnSettingsStep(wiz, stepCount, totalDockerSteps)
	wiz.externalPrysmSettingsModal = createExternalPrysmSettingsStep(wiz, stepCount, totalDockerSteps)
	stepCount++

	// Step 6 - VC settings
	wiz.graffitiModal = createGraffitiStep(wiz, stepCount, totalDockerSteps)
	wiz.doppelgangerDetectionModal = createDoppelgangerStep(wiz, stepCount, totalDockerSteps)
	stepCount++

	// Step 7 - Fallback clients
	wiz.useFallbackModal = createUseFallbackStep(wiz, stepCount, totalDockerSteps)
	wiz.fallbackNormalModal = createFallbackNormalStep(wiz, stepCount, totalDockerSteps)
	wiz.fallbackPrysmModal = createFallbackPrysmStep(wiz, stepCount, totalDockerSteps)
	stepCount++

	// Step 8 - Metrics
	wiz.metricsModal = createMetricsStep(wiz, stepCount, totalDockerSteps)
	stepCount++

	// Step 9 - MEV Boost
	wiz.mevModeModal = createMevModeStep(wiz, stepCount, totalDockerSteps)
	wiz.localMevModal = createLocalMevStep(wiz, stepCount, totalDockerSteps)
	wiz.externalMevModal = createExternalMevStep(wiz, stepCount, totalDockerSteps)
	stepCount++

	// Done
	wiz.finishedModal = createFinishedStep(wiz, stepCount, totalDockerSteps)

	// ===================
	// === Native Mode ===
	// ===================
	totalNativeSteps := 10
	stepCount = 0

	// Step 1 - Welcome
	wiz.nativeWelcomeModal = createNativeWelcomeStep(wiz, stepCount, totalNativeSteps)
	stepCount++

	// Step 2 - Network
	wiz.nativeNetworkModal = createNativeNetworkStep(wiz, stepCount, totalNativeSteps)
	stepCount++

	// Step 3 - EC settings
	wiz.nativeEcModal = createNativeEcStep(wiz, stepCount, totalNativeSteps)
	wiz.nativeEcUrlModal = createNativeEcUrlStep(wiz, stepCount, totalNativeSteps)
	stepCount++

	// Step 4 - BN settings
	wiz.nativeBnModal = createNativeBnStep(wiz, stepCount, totalNativeSteps)
	wiz.nativeBnUrlModal = createNativeBnUrlStep(wiz, stepCount, totalNativeSteps)
	stepCount++

	// Step 5 - Fallback clients
	wiz.nativeUseFallbackModal = createNativeUseFallbackStep(wiz, stepCount, totalNativeSteps)
	wiz.nativeFallbackModal = createNativeFallbackStep(wiz, stepCount, totalNativeSteps)
	stepCount++

	// Step 6 - Native stuff
	wiz.nativeDataModal = createNativeDataStep(wiz, stepCount, totalNativeSteps)
	stepCount++

	// Step 7 - Metrics
	wiz.nativeMetricsModal = createNativeMetricsStep(wiz, stepCount, totalNativeSteps)
	stepCount++

	// Step 8 - MEV Boost
	wiz.nativeMevModal = createNativeMevStep(wiz, stepCount, totalNativeSteps)
	stepCount++

	// Done
	wiz.nativeFinishedModal = createNativeFinishedStep(wiz, stepCount, totalNativeSteps)

	return wiz
}
