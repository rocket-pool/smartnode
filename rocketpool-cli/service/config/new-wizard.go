package config

type newUserWizard struct {
	md                     *mainDisplay
	welcomeModal           *page
	networkModal           *page
	executionModeModal     *page
	executionDockerModal   *page
	executionExternalModal *DirectionalModal
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
		`______           _        _    ______           _
| ___ \         | |      | |   | ___ \         | |
| |_/ /___   ___| | _____| |_  | |_/ /__   ___ | |
|    // _ \ / __| |/ / _ \ __| |  __/ _ \ / _ \| |
| |\ \ (_) | (__|   <  __/ |_  | | | (_) | (_) | |
\_| \_\___/ \___|_|\_\___|\__| \_|  \___/ \___/|_|`+"\n\n"+

			"Welcome to the Smartnode configuration wizard!\n\n"+
			"Since this is your first time configuring the Smartnode, we'll walk you through the basic setup.\n\n",
		[]string{"Next", "Quit"}, nil, DirectionalModalHorizontal,
	)
	modal.done = func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.setPage(wiz.networkModal)
		} else if buttonIndex == 1 {
			wiz.md.app.Stop()
		}
	}

	page := newPage(nil, "new-user-welcome", "New User Wizard > [1/8] Welcome", "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)

	/*
		modal := NewDirectionalModal(DirectionalModalHorizontal, wiz.md.app).
			SetText(`
			______           _        _    ______           _
				| ___ \         | |      | |   | ___ \         | |
				| |_/ /___   ___| | _____| |_  | |_/ /__   ___ | |
				|    // _ \ / __| |/ / _ \ __| |  __/ _ \ / _ \| |
				| |\ \ (_) | (__|   <  __/ |_  | | | (_) | (_) | |
				\_| \_\___/ \___|_|\_\___|\__| \_|  \___/ \___/|_|` + "\n\n" +

				"Welcome to the Smartnode configuration wizard!\n\n" +
				"Since this is your first time configuring the Smartnode, we'll walk you through the basic setup.\n\n",
			).
			AddButtons([]string{"Next", "Quit", "this is a really long button to make the thing fit"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonIndex == 0 {
					wiz.md.app.SetRoot(wiz.networkModal, true)
				} else if buttonIndex == 1 {
					wiz.md.app.Stop()
				}
			})
	*/
	wiz.welcomeModal = page

}

// Create the network modal
func (wiz *newUserWizard) createNetworkModal() {

	modal := newModalLayout(
		wiz.md.app,
		40,
		"Which network would you like to use?",
		[]string{"The Prater Testnet", "The Ethereum Mainnet"},
		nil,
		DirectionalModalVertical)
	modal.done = func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 2 {
			wiz.md.app.Stop()
		} else {
			wiz.md.setPage(wiz.executionModeModal)
		}
	}

	page := newPage(nil, "new-user-network", "New User Wizard > [2/8] Network", "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)

	/*
		modal := NewDirectionalModal(DirectionalModalVertical, wiz.md.app).
			SetText("Which network would you like to use?").
			AddButtons([]string{"The Prater Testnet", "The Ethereum Mainnet", "Quit without Saving"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonIndex == 2 {
					wiz.md.app.Stop()
				} else {
					wiz.md.app.SetRoot(wiz.executionModeModal, true)
				}
			})
	*/
	wiz.networkModal = page

}

// Create the execution client mode selection modal
func (wiz *newUserWizard) createExecutionModeModal() {

	modal := newModalLayout(
		wiz.md.app,
		60,
		"Let's start by choosing how you'd like to run your Execution Client (formerly eth1 client).\n\n"+
			"Would you like Rocket Pool to run and manage its own client, or would you like it to use an existing client you run and manage outside of Rocket Pool (formerly known as \"Hybrid Mode\")?",
		[]string{
			"Let Rocket Pool Manage its Own Client (Default)",
			"Use an Existing External Client (Hybrid Mode)",
		},
		nil,
		DirectionalModalVertical)
	modal.done = func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.setPage(wiz.executionDockerModal)
		} else if buttonIndex == 1 {
			wiz.md.app.SetRoot(wiz.executionExternalModal, true)
		} else if buttonIndex == 2 {
			wiz.md.app.Stop()
		}
	}

	page := newPage(nil, "new-user-execution-mode", "New User Wizard > [3/8] Execution Client Mode", "", modal.borderGrid)
	wiz.md.pages.AddPage(page.id, page.content, true, false)

	wiz.executionModeModal = page

}

// Create the execution client Docker modal
func (wiz *newUserWizard) createExecutionDockerModal() {

	modal := newModalLayout(
		wiz.md.app,
		70,
		"Please select the Execution client you would like to use.\n\n"+
			"Highlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth1-clients to learn more about them.",
		[]string{
			"Geth",
			"Infura",
			"Pocket",
		},
		[]string{
			"Geth is one of the three original implementations of the Ethereum protocol. It is written in Go, fully open source and licensed under the GNU LGPL v3.",
			"Use infura.io as a light client for Eth 1.0. Not recommended for use in production.",
			"Use Pocket Network as a decentralized light client for Eth 1.0. Suitable for use in production.",
		},
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

	wiz.executionDockerModal = page

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
