package config

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

type newUserWizard struct {
	md                              *mainDisplay
	welcomeModal                    *page
	networkModal                    *choiceModalLayout
	executionModeModal              *choiceModalLayout
	executionLocalModal             *choiceModalLayout
	executionExternalModal          *textBoxModalLayout
	infuraModal                     *textBoxModalLayout
	fallbackInfuraModal             *textBoxModalLayout
	fallbackExecutionModal          *choiceModalLayout
	consensusModeModal              *choiceModalLayout
	consensusLocalModal             *choiceModalLayout
	consensusExternalSelectModal    *choiceModalLayout
	graffitiModal                   *textBoxModalLayout
	checkpointSyncProviderModal     *textBoxModalLayout
	doppelgangerDetectionModal      *choiceModalLayout
	lighthouseExternalSettingsModal *textBoxModalLayout
	prysmExternalSettingsModal      *textBoxModalLayout
	tekuExternalSettingsModal       *textBoxModalLayout
	externalGraffitiModal           *textBoxModalLayout
	metricsModal                    *choiceModalLayout
	finishedModal                   *page
}

func newNewUserWizard(md *mainDisplay) *newUserWizard {

	wiz := &newUserWizard{
		md: md,
	}

	wiz.createWelcomeModal()
	wiz.createNetworkModal()
	wiz.createExecutionModeModal()
	wiz.createLocalExecutionModal()
	wiz.createExternalExecutionModal()
	wiz.createInfuraModal()
	wiz.createFallbackExecutionModal()
	wiz.createFallbackInfuraModal()
	wiz.createConsensusModeModal()
	wiz.createExternalConsensusModal()
	wiz.createGraffitiModal()
	wiz.createCheckpointSyncProviderModal()
	wiz.createDoppelgangerModal()
	wiz.createLighthouseExternalSettingsModal()
	wiz.createPrysmExternalSettingsModal()
	wiz.createTekuExternalSettingsModal()
	wiz.createExternalGraffitiModal()
	wiz.createMetricsModal()
	wiz.createFinishedModal()

	return wiz

}

// ========================
// === 1: Welcome Modal ===
// ========================
func (wiz *newUserWizard) createWelcomeModal() {

	title := "[1/9] Welcome"
	modal := newChoiceModalLayout(
		wiz.md.app,
		title,
		60,
		shared.Logo+"\n\n"+

			"Welcome to the Smartnode configuration wizard!\n\n"+
			"Since this is your first time configuring the Smartnode, we'll walk you through the basic setup.\n\n",
		[]string{"Next", "Quit"}, nil, DirectionalModalHorizontal,
	)
	modal.done = func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.setPage(wiz.networkModal.page)
			wiz.networkModal.focus(0)
		} else if buttonIndex == 1 {
			wiz.md.app.Stop()
		}
	}

	page := newPage(nil, "new-user-welcome", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)

	wiz.welcomeModal = page

}

// =========================
// === 2: Select Network ===
// =========================
func (wiz *newUserWizard) createNetworkModal() {

	title := "[2/9] Network"

	// Create the button names and descriptions from the config
	networks := wiz.md.Config.Smartnode.Network.Options
	networkNames := []string{}
	networkDescriptions := []string{}
	for _, network := range networks {
		networkNames = append(networkNames, network.Name)
		networkDescriptions = append(networkDescriptions, network.Description)
	}

	// Create the modal
	modal := newChoiceModalLayout(
		wiz.md.app,
		title,
		70,
		"Let's start by choosing which network you'd like to use.\n\n",
		networkNames,
		networkDescriptions,
		DirectionalModalVertical)

	// Set up the callbacks
	modal.done = func(buttonIndex int, buttonLabel string) {
		newNetwork := networks[buttonIndex].Value.(config.Network)
		wiz.md.Config.ChangeNetwork(newNetwork)
		wiz.md.setPage(wiz.executionModeModal.page)
		wiz.executionModeModal.focus(0)
	}
	modal.back = func() {
		wiz.md.setPage(wiz.welcomeModal)
	}

	// Create the page
	wiz.networkModal = modal
	page := newPage(nil, "new-user-network", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page
}

// ================================
// === 3: Select Execution Mode ===
// ================================
func (wiz *newUserWizard) createExecutionModeModal() {

	title := "[3/9] Execution Client Mode"

	// Create the button names and descriptions from the config
	modes := wiz.md.Config.ExecutionClientMode.Options
	modeNames := []string{}
	modeDescriptions := []string{}
	for _, mode := range modes {
		modeNames = append(modeNames, mode.Name)
		modeDescriptions = append(modeDescriptions, mode.Description)
	}

	// Create the modal
	modal := newChoiceModalLayout(
		wiz.md.app,
		title,
		76,
		"Now let's decide which mode you'd like to use for the Execution client (formerly eth1 client).\n\n"+
			"Would you like Rocket Pool to run and manage its own client, or would you like it to use an existing client you run and manage outside of Rocket Pool (also known as \"Hybrid Mode\")?",

		modeNames,
		modeDescriptions,
		DirectionalModalVertical)

	// Set up the callbacks
	modal.done = func(buttonIndex int, buttonLabel string) {
		wiz.md.Config.ExecutionClientMode.Value = modes[buttonIndex].Value
		switch modes[buttonIndex].Value {
		case config.Mode_Local:
			wiz.md.setPage(wiz.executionLocalModal.page)
			wiz.executionLocalModal.focus(0)
		case config.Mode_External:
			wiz.md.setPage(wiz.executionExternalModal.page)
			wiz.executionExternalModal.focus()
		default:
			panic(fmt.Sprintf("Unknown execution client mode %s", modes[buttonIndex].Value))
		}
	}
	modal.back = func() {
		wiz.md.setPage(wiz.networkModal.page)
		wiz.networkModal.focus(0)
	}

	// Create the page
	wiz.executionModeModal = modal
	page := newPage(nil, "new-user-execution-mode", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ========================================
// === 4a: Select Local Exection Client ===
// ========================================
func (wiz *newUserWizard) createLocalExecutionModal() {

	title := "[4/9] Execution Client > Selection"

	// Create the button names and descriptions from the config
	clients := wiz.md.Config.ExecutionClient.Options
	clientNames := []string{}
	clientDescriptions := []string{}
	for _, client := range clients {
		clientNames = append(clientNames, client.Name)
		clientDescriptions = append(clientDescriptions, client.Description)
	}

	// Create the modal
	modal := newChoiceModalLayout(
		wiz.md.app,
		title,
		76,
		"Please select the Execution client you would like to use.\n\n"+
			"Highlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth1-clients to learn more about them.",
		clientNames,
		clientDescriptions,
		DirectionalModalVertical)

	// Set up the callbacks
	modal.done = func(buttonIndex int, buttonLabel string) {
		selectedClient := clients[buttonIndex].Value.(config.ExecutionClient)
		wiz.md.Config.ExecutionClient.Value = selectedClient
		switch selectedClient {
		case config.ExecutionClient_Geth:
			// Geth doesn't have any required parameters so move on
			wiz.md.setPage(wiz.fallbackExecutionModal.page)
			wiz.fallbackExecutionModal.focus(0)
		case config.ExecutionClient_Infura:
			// Switch to the Infura dialog
			wiz.md.setPage(wiz.infuraModal.page)
			wiz.infuraModal.focus()
		case config.ExecutionClient_Pocket:
			// Pocket doesn't have any required parameters so move on
			wiz.md.setPage(wiz.fallbackExecutionModal.page)
			wiz.fallbackExecutionModal.focus(0)
		}
	}
	modal.back = func() {
		wiz.md.setPage(wiz.executionModeModal.page)
		wiz.executionModeModal.focus(0)
	}

	// Create the page
	wiz.executionLocalModal = modal
	page := newPage(nil, "new-user-execution-local", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ==========================================
// === 4b: Setup External Exection Client ===
// ==========================================
func (wiz *newUserWizard) createExternalExecutionModal() {

	title := "[4/9] Execution Client (External)"

	// Create the labels
	httpLabel := wiz.md.Config.ExternalExecution.HttpUrl.Name
	wsLabel := wiz.md.Config.ExternalExecution.WsUrl.Name

	// Create the modal
	modal := newTextBoxModalLayout(
		wiz.md.app,
		title,
		70,
		"Please enter the URL of the HTTP-based RPC API and the URL of the Websocket-based RPC API for your existing client.\n\n"+
			"For example: `http://192.168.1.45:8545` and `ws://192.168.1.45:8546`",
		[]string{httpLabel, wsLabel},
		[]string{})

	// Set up the callbacks
	modal.done = func(text map[string]string) {
		wiz.md.Config.ExternalExecution.HttpUrl.Value = text[httpLabel]
		wiz.md.Config.ExternalExecution.WsUrl.Value = text[wsLabel]
		wiz.md.setPage(wiz.fallbackExecutionModal.page)
		wiz.fallbackExecutionModal.focus(0)
	}
	modal.back = func() {
		wiz.md.setPage(wiz.executionModeModal.page)
		wiz.executionModeModal.focus(0)
	}

	// Create the page
	wiz.executionExternalModal = modal
	page := newPage(nil, "new-user-execution-external", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ========================
// === 4c: Local Infura ===
// ========================
func (wiz *newUserWizard) createInfuraModal() {

	title := "[4/9] Execution Client > Infura"

	// Create the labels
	projectIdLabel := wiz.md.Config.Infura.ProjectID.Name

	// Create the modal
	modal := newTextBoxModalLayout(
		wiz.md.app,
		title,
		70,
		"Please enter the Project ID for your Infura Ethereum project. You can find this on the Infura website, in your Ethereum project settings.",
		[]string{projectIdLabel},
		[]string{})

	// Set up the callbacks
	modal.done = func(text map[string]string) {
		wiz.md.Config.Infura.ProjectID.Value = text[projectIdLabel]
		wiz.md.setPage(wiz.fallbackExecutionModal.page)
		wiz.fallbackExecutionModal.focus(0)
	}
	modal.back = func() {
		wiz.md.setPage(wiz.executionLocalModal.page)
		wiz.executionLocalModal.focus(0)
	}

	// Create the page
	wiz.infuraModal = modal
	page := newPage(nil, "new-user-execution-infura", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// =====================================
// === 5a: Fallback Execution Client ===
// =====================================
func (wiz *newUserWizard) createFallbackExecutionModal() {

	title := "[5/9] Fallback Execution Client"

	// Create the button names and descriptions from the config
	clients := wiz.md.Config.FallbackExecutionClient.Options
	clientNames := []string{"None"}
	clientDescriptions := []string{"Do not use a fallback client."}
	for _, client := range clients {
		clientNames = append(clientNames, client.Name)
		clientDescriptions = append(clientDescriptions, client.Description)
	}

	// Create the modal
	modal := newChoiceModalLayout(
		wiz.md.app,
		title,
		70,
		"If you would like to add a fallback Execution client, please choose it below.\n\nThe Smartnode will temporarily use this instead of your main Execution client if the main client ever fails.\nIt will switch back to the main client when it starts working again.",
		clientNames,
		clientDescriptions,
		DirectionalModalVertical,
	)

	// Set up the callbacks
	modal.done = func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.Config.UseFallbackExecutionClient.Value = false
			wiz.md.setPage(wiz.consensusModeModal.page)
			wiz.consensusModeModal.focus(0)
		} else {
			wiz.md.Config.UseFallbackExecutionClient.Value = true
			selectedClient := clients[buttonIndex-1].Value.(config.ExecutionClient)
			wiz.md.Config.FallbackExecutionClient.Value = selectedClient
			switch selectedClient {
			case config.ExecutionClient_Infura:
				// Switch to the Infura dialog
				wiz.md.setPage(wiz.fallbackInfuraModal.page)
				wiz.fallbackInfuraModal.focus()
			default:
				wiz.md.setPage(wiz.consensusModeModal.page)
				wiz.consensusModeModal.focus(0)
			}
		}
	}
	modal.back = func() {
		wiz.md.setPage(wiz.executionModeModal.page)
		wiz.executionModeModal.focus(0)
	}

	// Create the page
	wiz.fallbackExecutionModal = modal
	page := newPage(nil, "new-user-fallback-execution", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ===========================
// === 5b: Fallback Infura ===
// ===========================
func (wiz *newUserWizard) createFallbackInfuraModal() {

	title := "[5/9] Fallback Execution Client > Infura"

	// Create the labels
	projectIdLabel := wiz.md.Config.FallbackInfura.ProjectID.Name

	// Create the modal
	modal := newTextBoxModalLayout(
		wiz.md.app,
		title,
		70,
		"Please enter the Project ID for your Infura Ethereum project. You can find this on the Infura website, in your Ethereum project settings.",
		[]string{projectIdLabel},
		[]string{})

	// Set up the callbacks
	modal.done = func(text map[string]string) {
		wiz.md.Config.FallbackInfura.ProjectID.Value = text[projectIdLabel]
		wiz.md.setPage(wiz.consensusModeModal.page)
	}
	modal.back = func() {
		wiz.md.setPage(wiz.fallbackExecutionModal.page)
		wiz.fallbackExecutionModal.focus(0)
	}

	// Create the page
	wiz.fallbackInfuraModal = modal
	page := newPage(nil, "new-user-fallback-execution-infura", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ================================
// === 6: Select Consensus Mode ===
// ================================
func (wiz *newUserWizard) createConsensusModeModal() {

	title := "[6/9] Consensus Client Mode"

	// Create the button names and descriptions from the config
	modes := wiz.md.Config.ConsensusClientMode.Options
	modeNames := []string{}
	modeDescriptions := []string{}
	for _, mode := range modes {
		modeNames = append(modeNames, mode.Name)
		modeDescriptions = append(modeDescriptions, mode.Description)
	}

	// Create the modal
	modal := newChoiceModalLayout(
		wiz.md.app,
		title,
		76,
		"Next, let's decide which mode you'd like to use for the Consensus client (formerly eth2 client).\n\n"+
			"Would you like Rocket Pool to run and manage its own client, or would you like it to use an existing client you run and manage outside of Rocket Pool (also known as \"Hybrid Mode\")?",

		modeNames,
		modeDescriptions,
		DirectionalModalVertical)

	// Set up the callbacks
	modal.done = func(buttonIndex int, buttonLabel string) {
		wiz.md.Config.ConsensusClientMode.Value = modes[buttonIndex].Value
		switch modes[buttonIndex].Value {
		case config.Mode_Local:
			wiz.createLocalConsensusModal()
			wiz.md.setPage(wiz.consensusLocalModal.page)
			wiz.consensusLocalModal.focus(0)
		case config.Mode_External:
			wiz.md.setPage(wiz.consensusExternalSelectModal.page)
		default:
			panic(fmt.Sprintf("Unknown execution client mode %s", modes[buttonIndex].Value))
		}
	}
	modal.back = func() {
		wiz.md.setPage(wiz.fallbackExecutionModal.page)
		wiz.fallbackExecutionModal.focus(0)
	}

	// Create the page
	wiz.consensusModeModal = modal
	page := newPage(nil, "new-user-consensus-mode", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// =========================================
// === 7a: Select Local Consensus Client ===
// =========================================
func (wiz *newUserWizard) createLocalConsensusModal() {

	title := "[7/9] Consensus Client > Selection"

	// Get the list of clients
	goodClients, badClients := wiz.md.Config.GetCompatibleConsensusClients()

	// Create the button names and descriptions from the config
	clients := wiz.md.Config.ConsensusClient.Options
	clientNames := []string{}
	clientDescriptions := []string{}
	for _, client := range goodClients {
		clientNames = append(clientNames, client.Name)
		clientDescriptions = append(clientDescriptions, client.Description)
	}

	incompatibleClientWarning := ""
	if len(badClients) > 0 {
		incompatibleClientWarning = fmt.Sprintf("\n\n[orange]NOTE: The following clients are incompatible with your choice of Execution and/or fallback Execution clients: %s", strings.Join(badClients, ", "))
	}

	// Create the modal
	modal := newChoiceModalLayout(
		wiz.md.app,
		title,
		76,
		fmt.Sprintf("Please select the Consensus client you would like to use.\n\n"+
			"Highlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth2-clients to learn more about them.%s", incompatibleClientWarning),
		clientNames,
		clientDescriptions,
		DirectionalModalVertical)

	// Set up the callbacks
	modal.done = func(buttonIndex int, buttonLabel string) {
		selectedClient := clients[buttonIndex].Value.(config.ConsensusClient)
		wiz.md.Config.ConsensusClient.Value = selectedClient
		wiz.md.setPage(wiz.graffitiModal.page)
		wiz.graffitiModal.focus()
	}
	modal.back = func() {
		wiz.md.setPage(wiz.consensusModeModal.page)
		wiz.consensusModeModal.focus(0)
	}

	// Create the page
	wiz.consensusLocalModal = modal
	page := newPage(nil, "new-user-consensus-local", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ============================================
// === 7b: Select External Consensus Client ===
// ============================================
func (wiz *newUserWizard) createExternalConsensusModal() {

	title := "[7/9] Consensus Client (External) > Selection"

	// Create the button names and descriptions from the config
	clients := wiz.md.Config.ExternalConsensusClient.Options
	clientNames := []string{}
	for _, client := range clients {
		clientNames = append(clientNames, client.Name)
	}

	// Create the modal
	modal := newChoiceModalLayout(
		wiz.md.app,
		title,
		70,
		"Which Consensus client are you externally managing? Each of them has small behavioral differences, so we'll need to know which one you're using in order to connect to it properly.\n\n[orange]Note: if your client is not listed here, it isn't compatible with external management mode (Hybrid Mode).",
		clientNames,
		[]string{},
		DirectionalModalVertical)

	// Set up the callbacks
	modal.done = func(buttonIndex int, buttonLabel string) {
		selectedClient := clients[buttonIndex].Value.(config.ConsensusClient)
		wiz.md.Config.ExternalConsensusClient.Value = selectedClient
		switch selectedClient {
		case config.ConsensusClient_Lighthouse:
			wiz.md.setPage(wiz.lighthouseExternalSettingsModal.page)
			wiz.lighthouseExternalSettingsModal.focus()
		case config.ConsensusClient_Prysm:
			wiz.md.setPage(wiz.prysmExternalSettingsModal.page)
			wiz.prysmExternalSettingsModal.focus()
		case config.ConsensusClient_Teku:
			wiz.md.setPage(wiz.tekuExternalSettingsModal.page)
			wiz.tekuExternalSettingsModal.focus()
		}
	}
	modal.back = func() {
		wiz.md.setPage(wiz.consensusModeModal.page)
		wiz.consensusModeModal.focus(0)
	}

	// Create the page
	wiz.consensusExternalSelectModal = modal
	page := newPage(nil, "new-user-consensus-external", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ====================
// === 7c: Graffiti ===
// ====================
func (wiz *newUserWizard) createGraffitiModal() {

	title := "[7/9] Consensus Client > Graffiti"

	// Create the labels
	graffitiLabel := wiz.md.Config.ConsensusCommon.Graffiti.Name

	// Create the modal
	modal := newTextBoxModalLayout(
		wiz.md.app,
		title,
		70,
		"If you would like to add a short custom message to each block that your minipools propose (called the block's \"graffiti\"), please enter it here. The graffiti is limited to 16 characters max.",
		[]string{graffitiLabel},
		[]string{})

	// Set up the callbacks
	modal.done = func(text map[string]string) {
		wiz.md.Config.ConsensusCommon.Graffiti.Value = text[graffitiLabel]
		// Get the selected client
		client, err := wiz.md.Config.GetSelectedConsensusClientConfig()
		if err != nil {
			wiz.md.app.Stop()
			fmt.Printf("Error setting the consensus client graffiti: %s", err.Error())
		}

		// Check to see if it supports checkpoint sync or doppelganger detection
		unsupportedParams := client.(config.LocalConsensusConfig).GetUnsupportedCommonParams()
		supportsCheckpointSync := true
		supportsDoppelganger := true
		for _, param := range unsupportedParams {
			if param == config.CheckpointSyncUrlID {
				supportsCheckpointSync = false
			} else if param == config.DoppelgangerDetectionID {
				supportsDoppelganger = false
			}
		}

		// Move to the next appropriate dialog
		if supportsCheckpointSync {
			wiz.md.setPage(wiz.checkpointSyncProviderModal.page)
			wiz.checkpointSyncProviderModal.focus()
		} else if supportsDoppelganger {
			wiz.md.setPage(wiz.doppelgangerDetectionModal.page)
			wiz.doppelgangerDetectionModal.focus(0)
		} else {
			wiz.md.setPage(wiz.metricsModal.page)
		}
	}
	modal.back = func() {
		wiz.md.setPage(wiz.consensusModeModal.page)
		wiz.consensusModeModal.focus(0)
	}

	// Create the page
	wiz.graffitiModal = modal
	page := newPage(nil, "new-user-consensus-graffiti", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ===========================
// === 7d: Checkpoint Sync ===
// ===========================
func (wiz *newUserWizard) createCheckpointSyncProviderModal() {

	title := "[7/9] Consensus Client > Checkpoint Sync"

	// Create the labels
	checkpointSyncLabel := wiz.md.Config.ConsensusCommon.CheckpointSyncProvider.Name

	// Create the modal
	modal := newTextBoxModalLayout(
		wiz.md.app,
		title,
		76,
		"Your client supports Checkpoint Sync. This powerful feature allows it to copy the most recent state from a separate Consensus client that you trust, so you don't have to wait for it to sync from scratch - you can start using it instantly!\n\n"+
			"Take a look at our documentation for an example of how to use it:\nhttps://docs.rocketpool.net/guides/node/docker.html#eth2-checkpoint-syncing-with-infura\n\n"+
			"If you would like to use Checkpoint Sync, please provide the provider URL here. If you don't want to use it, leave it blank.",
		[]string{checkpointSyncLabel},
		[]string{})

	// Set up the callbacks
	modal.done = func(text map[string]string) {
		wiz.md.Config.ConsensusCommon.CheckpointSyncProvider.Value = text[checkpointSyncLabel]
		// Get the selected client
		client, err := wiz.md.Config.GetSelectedConsensusClientConfig()
		if err != nil {
			wiz.md.app.Stop()
			fmt.Printf("Error setting the consensus client checkpoint sync provider: %s", err.Error())
		}

		// Check to see if it supports doppelganger detection
		unsupportedParams := client.(config.LocalConsensusConfig).GetUnsupportedCommonParams()
		supportsDoppelganger := true
		for _, param := range unsupportedParams {
			if param == config.DoppelgangerDetectionID {
				supportsDoppelganger = false
			}
		}

		// Move to the next appropriate dialog
		if supportsDoppelganger {
			wiz.md.setPage(wiz.doppelgangerDetectionModal.page)
			wiz.doppelgangerDetectionModal.focus(0)
		} else {
			wiz.md.setPage(wiz.metricsModal.page)
		}
	}
	modal.back = func() {
		wiz.md.setPage(wiz.graffitiModal.page)
		wiz.graffitiModal.focus()
	}

	// Create the page
	wiz.checkpointSyncProviderModal = modal
	page := newPage(nil, "new-user-consensus-checkpoint", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ==================================
// === 7e: Doppelganger Detection ===
// ==================================
func (wiz *newUserWizard) createDoppelgangerModal() {

	title := "[7/9] Consensus Client > Doppelganger Protection"

	// Create the modal
	modal := newChoiceModalLayout(
		wiz.md.app,
		title,
		76,
		"Your client supports Doppelganger Protection. This feature can prevent your minipools from being slashed (penalized for a lot of ETH and removed from the Beacon Chain) if you accidentally run your validator keys on multiple machines at the same time.\n\n"+
			"If enabled, whenever your validator client restarts, it will intentionally miss 2-3 attestations (for each minipool). If all of them are missed successfully, you can be confident that you are safe to start attesting.\n\n"+
			"Would you like to enable Doppelganger Protection?",
		[]string{"Yes", "No"},
		[]string{},
		DirectionalModalHorizontal)

	// Set up the callbacks
	modal.done = func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.Config.ConsensusCommon.DoppelgangerDetection.Value = true
		} else {
			wiz.md.Config.ConsensusCommon.DoppelgangerDetection.Value = false
		}
		wiz.md.setPage(wiz.metricsModal.page)
	}
	modal.back = func() {
		wiz.md.setPage(wiz.graffitiModal.page)
		wiz.graffitiModal.focus()
	}

	// Create the page
	wiz.doppelgangerDetectionModal = modal
	page := newPage(nil, "new-user-consensus-doppelganger", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ===============================
// === 7f: External Lighthouse ===
// ===============================
func (wiz *newUserWizard) createLighthouseExternalSettingsModal() {

	title := "[7/9] Consensus Client (External) > Settings"

	// Create the labels
	httpUrlLabel := wiz.md.Config.ExternalLighthouse.HttpUrl.Name

	// Create the modal
	modal := newTextBoxModalLayout(
		wiz.md.app,
		title,
		70,
		"Please provide the URL of your Lighthouse client's HTTP API (for example: `http://192.168.1.40:5052`).\n\nNote that if you're running it on the same machine as the Smartnode, you cannot use `localhost` or `127.0.0.1`; you must use your machine's LAN IP address.",
		[]string{httpUrlLabel},
		[]string{})

	// Set up the callbacks
	modal.done = func(text map[string]string) {
		wiz.md.Config.ExternalLighthouse.HttpUrl.Value = text[httpUrlLabel]
		wiz.md.setPage(wiz.externalGraffitiModal.page)
	}
	modal.back = func() {
		wiz.md.setPage(wiz.consensusModeModal.page)
		wiz.consensusModeModal.focus(0)
	}

	// Create the page
	wiz.lighthouseExternalSettingsModal = modal
	page := newPage(nil, "new-user-consensus-external-lighthouse", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ==========================
// === 7g: External Prysm ===
// ==========================
func (wiz *newUserWizard) createPrysmExternalSettingsModal() {

	title := "[7/9] Consensus Client (External) > Settings"

	// Create the labels
	httpUrlLabel := wiz.md.Config.ExternalPrysm.HttpUrl.Name
	jsonRpcUrlLabel := wiz.md.Config.ExternalPrysm.JsonRpcUrl.Name

	// Create the modal
	modal := newTextBoxModalLayout(
		wiz.md.app,
		title,
		70,
		"Please provide the URL of your Prysm client's HTTP API (for example: `http://192.168.1.40:5052`) and the URL of its JSON RPC API (e.g., `http://192.168.1.40:5053`).\n\nNote that if you're running it on the same machine as the Smartnode, you cannot use `localhost` or `127.0.0.1`; you must use your machine's LAN IP address.",
		[]string{httpUrlLabel, jsonRpcUrlLabel},
		[]string{})

	// Set up the callbacks
	modal.done = func(text map[string]string) {
		wiz.md.Config.ExternalPrysm.HttpUrl.Value = text[httpUrlLabel]
		wiz.md.Config.ExternalPrysm.JsonRpcUrl.Value = text[jsonRpcUrlLabel]
		wiz.md.setPage(wiz.externalGraffitiModal.page)
	}
	modal.back = func() {
		wiz.md.setPage(wiz.consensusModeModal.page)
		wiz.consensusModeModal.focus(0)
	}

	// Create the page
	wiz.prysmExternalSettingsModal = modal
	page := newPage(nil, "new-user-consensus-external-prysm", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// =========================
// === 7f: External Teku ===
// =========================
func (wiz *newUserWizard) createTekuExternalSettingsModal() {

	title := "[7/9] Consensus Client (External) > Settings"

	// Create the labels
	httpUrlLabel := wiz.md.Config.ExternalTeku.HttpUrl.Name

	// Create the modal
	modal := newTextBoxModalLayout(
		wiz.md.app,
		title,
		70,
		"Please provide the URL of your Teku client's HTTP API (for example: `http://192.168.1.40:5052`).\n\nNote that if you're running it on the same machine as the Smartnode, you cannot use `localhost` or `127.0.0.1`; you must use your machine's LAN IP address.",
		[]string{httpUrlLabel},
		[]string{})

	// Set up the callbacks
	modal.done = func(text map[string]string) {
		wiz.md.Config.ExternalTeku.HttpUrl.Value = text[httpUrlLabel]
		wiz.md.setPage(wiz.externalGraffitiModal.page)
	}
	modal.back = func() {
		wiz.md.setPage(wiz.consensusModeModal.page)
		wiz.consensusModeModal.focus(0)
	}

	// Create the page
	wiz.tekuExternalSettingsModal = modal
	page := newPage(nil, "new-user-consensus-external-teku", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// =============================
// === 7g: External Graffiti ===
// =============================
func (wiz *newUserWizard) createExternalGraffitiModal() {

	title := "[7/9] Consensus Client (External) > Graffiti"

	// Create the labels - use the vanilla graffiti name
	graffitiLabel := wiz.md.Config.ConsensusCommon.Graffiti.Name

	// Create the modal
	modal := newTextBoxModalLayout(
		wiz.md.app,
		title,
		70,
		"If you would like to add a short custom message to each block that your minipools propose (called the block's \"graffiti\"), please enter it here. The graffiti is limited to 16 characters max.",
		[]string{graffitiLabel},
		[]string{})

	// Set up the callbacks
	modal.done = func(text map[string]string) {
		// Get the selected client
		switch wiz.md.Config.ExternalConsensusClient.Value.(config.ConsensusClient) {
		case config.ConsensusClient_Lighthouse:
			wiz.md.Config.ExternalLighthouse.Graffiti.Value = text[graffitiLabel]
		case config.ConsensusClient_Prysm:
			wiz.md.Config.ExternalPrysm.Graffiti.Value = text[graffitiLabel]
		case config.ConsensusClient_Teku:
			wiz.md.Config.ExternalTeku.Graffiti.Value = text[graffitiLabel]
		}
		wiz.md.setPage(wiz.metricsModal.page)
		wiz.metricsModal.focus(0)
	}
	modal.back = func() {
		wiz.md.setPage(wiz.consensusModeModal.page)
		wiz.consensusModeModal.focus(0)
	}

	// Create the page
	wiz.externalGraffitiModal = modal
	page := newPage(nil, "new-user-consensus-external-graffiti", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ==================
// === 8: Metrics ===
// ==================
func (wiz *newUserWizard) createMetricsModal() {

	title := "[8/9] Metrics"

	// Create the modal
	modal := newChoiceModalLayout(
		wiz.md.app,
		title,
		76,
		"Would you like to enable the Smartnode's metrics monitoring system? This will monitor things such as hardware stats (CPU usage, RAM usage, free disk space), your minipool stats, stats about your node such as total RPL and ETH rewards, and much more. It also enables the Grafana dashboard to quickly and easily view these metrics (see https://docs.rocketpool.net/guides/node/grafana.html for an example).\n\nNone of this information will be sent to any remote servers for collection an analysis; this is purely for your own usage on your node.",
		[]string{"Yes", "No"},
		[]string{},
		DirectionalModalHorizontal)

	// Set up the callbacks
	modal.done = func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.Config.EnableMetrics.Value = true
		} else {
			wiz.md.Config.EnableMetrics.Value = false
		}
		wiz.md.setPage(wiz.finishedModal)
	}
	modal.back = func() {
		wiz.md.setPage(wiz.consensusModeModal.page)
		wiz.consensusModeModal.focus(0)
	}

	// Create the page
	wiz.metricsModal = modal
	page := newPage(nil, "new-user-metrics", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// ===================
// === 9: Finished ===
// ===================
func (wiz *newUserWizard) createFinishedModal() {

	title := "[9/9] Finished"

	modal := newChoiceModalLayout(
		wiz.md.app,
		title,
		40,
		"All done! You're ready to run.\n\n"+
			"If you'd like, you can review and change all of the Smartnode and client settings next or just save and exit.",
		[]string{
			"Review All Settings",
			"Save and Exit",
		},
		nil,
		DirectionalModalVertical)
	modal.done = func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.setPage(wiz.md.settingsHome.homePage)
		} else {
			wiz.md.ShouldSave = true
			wiz.md.app.Stop()
		}
	}

	page := newPage(nil, "new-user-finished", "New User Wizard > "+title, "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)

	wiz.finishedModal = page

}
