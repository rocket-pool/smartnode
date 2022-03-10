package config

type wizard struct {
	md                              *mainDisplay
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
}

func newWizard(md *mainDisplay) *wizard {

	wiz := &wizard{
		md: md,
	}

	totalSteps := 9

	wiz.welcomeModal = createWelcomeStep(wiz, 1, totalSteps)
	wiz.networkModal = createNetworkStep(wiz, 2, totalSteps)
	wiz.executionModeModal = createEcModeStep(wiz, 3, totalSteps)
	wiz.executionLocalModal = createLocalEcStep(wiz, 4, totalSteps)
	wiz.executionExternalModal = createExternalEcStep(wiz, 4, totalSteps)
	wiz.infuraModal = createInfuraStep(wiz, 4, totalSteps)
	wiz.fallbackExecutionModal = createFallbackEcStep(wiz, 5, totalSteps)
	wiz.fallbackInfuraModal = createFallbackInfuraStep(wiz, 5, totalSteps)
	wiz.consensusModeModal = createCcModeStep(wiz, 6, totalSteps)
	wiz.consensusExternalSelectModal = createExternalCcStep(wiz, 7, totalSteps)
	wiz.consensusLocalPrysmWarning = createPrysmWarningStep(wiz, 7, totalSteps)
	wiz.consensusLocalTekuWarning = createTekuWarningStep(wiz, 7, totalSteps)
	wiz.graffitiModal = createGraffitiStep(wiz, 7, totalSteps)
	wiz.checkpointSyncProviderModal = createCheckpointSyncStep(wiz, 7, totalSteps)
	wiz.doppelgangerDetectionModal = createDoppelgangerStep(wiz, 7, totalSteps)
	wiz.lighthouseExternalSettingsModal = createExternalLhStep(wiz, 7, totalSteps)
	wiz.prysmExternalSettingsModal = createExternalPrysmStep(wiz, 7, totalSteps)
	wiz.tekuExternalSettingsModal = createExternalTekuStep(wiz, 7, totalSteps)
	wiz.externalGraffitiModal = createExternalGraffitiStep(wiz, 7, totalSteps)
	wiz.externalDoppelgangerModal = createExternalDoppelgangerStep(wiz, 7, totalSteps)
	wiz.metricsModal = createMetricsStep(wiz, 8, totalSteps)
	wiz.finishedModal = createFinishedStep(wiz, 9, totalSteps)

	return wiz

}
