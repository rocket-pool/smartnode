package config

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

type newUserWizard struct {
	md                     *mainDisplay
	welcomeModal           *page
	networkModal           *modalLayout
	executionModeModal     *modalLayout
	executionLocalModal    *page
	executionExternalModal *page
	consensusModeModal     *DirectionalModal
	consensusDockerModal   *page
	consensusExternalMoadl *DirectionalModal
	finishedModal          *page
}

func newNewUserWizard(md *mainDisplay) *newUserWizard {

	wiz := &newUserWizard{
		md: md,
	}

	wiz.createWelcomeModal()
	wiz.createNetworkModal()
	wiz.createExecutionModeModal()
	wiz.createExecutionDockerModal()
	wiz.createConsensusDockerModal()
	wiz.createFinishedModal()

	return wiz

}

// Create the welcome modal
func (wiz *newUserWizard) createWelcomeModal() {

	modal := newModalLayout(
		wiz.md.app,
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

	page := newPage(nil, "new-user-welcome", "New User Wizard > [1/8] Welcome", "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)

	wiz.welcomeModal = page

}

// Create the network modal
func (wiz *newUserWizard) createNetworkModal() {

	// Create the button names and descriptions from the config
	networks := wiz.md.config.Smartnode.Network.Options
	networkNames := []string{}
	networkDescriptions := []string{}
	for _, network := range networks {
		networkNames = append(networkNames, network.Name)
		networkDescriptions = append(networkDescriptions, network.Description)
	}

	// Create the modal
	modal := newModalLayout(
		wiz.md.app,
		70,
		"Let's start by choosing which network you'd like to use.\n\n",
		networkNames,
		networkDescriptions,
		DirectionalModalVertical)

	// Set up the callbacks
	modal.done = func(buttonIndex int, buttonLabel string) {
		wiz.md.config.Smartnode.Network.Value = networks[buttonIndex].Value
		wiz.md.setPage(wiz.executionModeModal.page)
		wiz.executionModeModal.focus(0)
	}

	// Create the page
	wiz.networkModal = modal
	page := newPage(nil, "new-user-network", "New User Wizard > [2/8] Network", "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page
}

// Create the execution client mode selection modal
func (wiz *newUserWizard) createExecutionModeModal() {

	// Create the button names and descriptions from the config
	modes := wiz.md.config.ExecutionClientMode.Options
	modeNames := []string{}
	modeDescriptions := []string{}
	for _, mode := range modes {
		modeNames = append(modeNames, mode.Name)
		modeDescriptions = append(modeDescriptions, mode.Description)
	}

	// Create the modal
	modal := newModalLayout(
		wiz.md.app,
		76,
		"Now let's decide which mode you'd like to use for the Execution client (formerly eth1 client).\n\n"+
			"Would you like Rocket Pool to run and manage its own client, or would you like it to use an existing client you run and manage outside of Rocket Pool (also known as \"Hybrid Mode\")?",
		/*[]string{
			"Let Rocket Pool Manage its Own Client (Default)",
			"Use an Existing External Client (Hybrid Mode)",
		},*/
		modeNames,
		modeDescriptions,
		DirectionalModalVertical)

	// Set up the callbacks
	modal.done = func(buttonIndex int, buttonLabel string) {
		wiz.md.config.ExecutionClientMode.Value = modes[buttonIndex].Value
		switch modes[buttonIndex].Value {
		case config.Mode_Local:
			wiz.md.setPage(wiz.executionLocalModal)
		case config.Mode_External:
			wiz.md.setPage(wiz.executionExternalModal)
		default:
			panic(fmt.Sprintf("Unknown execution client mode %s", modes[buttonIndex].Value))
		}
	}

	// Create the page
	wiz.executionModeModal = modal
	page := newPage(nil, "new-user-execution-mode", "New User Wizard > [3/8] Execution Client Mode", "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

}

// Create the execution client Docker modal
func (wiz *newUserWizard) createExecutionDockerModal() {

	modal := newModalLayout(
		wiz.md.app,
		90,
		"Please select the Execution client you would like to use.\n\n"+
			"Highlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth1-clients to learn more about them.",
		[]string{
			"Geth",
			"Infura",
			"Pocket",
		},
		[]string{},
		DirectionalModalVertical)
	modal.done = func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 3 {
			wiz.md.app.Stop()
		} else {
			wiz.md.setPage(wiz.consensusDockerModal)
		}
	}

	page := newPage(nil, "new-user-execution-docker", "New User Wizard > [4/8] Execution Client", "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)

	wiz.executionLocalModal = page

}

// Create the consensus client Docker modal
func (wiz *newUserWizard) createConsensusDockerModal() {

	modal := newModalLayout(
		wiz.md.app,
		80,
		"Please select the Consensus client you would like to use.\n\n"+
			"Highlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth2-clients to learn more about them.",
		[]string{
			"Lighthouse",
			"Nimbus",
			"Prysm",
			"Teku",
		},
		[]string{
			"Lighthouse is a Consensus client with a heavy focus on speed and security. The team behind it, Sigma Prime, is an information security and software engineering firm who have funded Lighthouse along with the Ethereum Foundation, Consensys, and private individuals. Lighthouse is built in Rust and offered under an Apache 2.0 License.",
			"Nimbus is a Consensus client implementation that strives to be as lightweight as possible in terms of resources used. This allows it to perform well on embedded systems, resource-restricted devices -- including Raspberry Pis and mobile devices -- and multi-purpose servers.",
			"Prysm is a Go implementation of a Consensus client with a focus on usability, security, and reliability. Prysm is developed by Prysmatic Labs, a company with the sole focus on the development of their client. Prysm is written in Go and released under a GPL-3.0 license.",
			"Teku is a Java-based Consensus client designed & built to meet institutional needs and security requirements. PegaSys is an arm of ConsenSys dedicated to building enterprise-ready clients and tools for interacting with the core Ethereum platform. Teku is Apache 2 licensed.",
		},
		DirectionalModalVertical)
	modal.done = func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 4 {
			wiz.md.app.Stop()
		} else {
			wiz.md.setPage(wiz.finishedModal)
		}
	}

	page := newPage(nil, "new-user-consensus-docker", "New User Wizard > [6/8] Consensus Client", "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)

	wiz.consensusDockerModal = page

}

// Create the finished modal
func (wiz *newUserWizard) createFinishedModal() {

	modal := newModalLayout(
		wiz.md.app,
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
			wiz.md.app.Stop()
		}
	}

	page := newPage(nil, "new-user-finished", "New User Wizard > [8/8] Finished", "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)

	wiz.finishedModal = page

}
