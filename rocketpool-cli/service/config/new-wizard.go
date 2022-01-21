package config


type newUserWizard struct {
	md *mainDisplay
	welcomeModal *DirectionalModal
	networkModal *DirectionalModal
	executionModeModal *DirectionalModal
	executionDockerModal *DirectionalModal
	executionExternalModal *DirectionalModal
	consensusModeModal *DirectionalModal
	consensusDockerModal *DirectionalModal
	consensusExternalMoadl *DirectionalModal
	finishedModal *DirectionalModal
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

	wiz.welcomeModal = modal

}


// Create the network modal
func (wiz *newUserWizard) createNetworkModal() {

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

	wiz.networkModal = modal

}


// Create the execution client mode selection modal
func (wiz *newUserWizard) createExecutionModeModal() {

	modal := NewDirectionalModal(DirectionalModalVertical, wiz.md.app).
		SetText("Let's start by choosing how you'd like to run your execution client (formerly eth1 client).\n\n" +
			"Would you like Rocket Pool to run and manage its own client, or would you like it to use an existing client you run and manage outside of Rocket Pool (formerly known as \"Hybrid Mode\")?",
		).
		AddButtons([]string{
			"Let Rocket Pool Manage its Own Client (Default)", 
			"Use an Existing External Client (Hybrid Mode)", 
			"Quit without Saving",
		}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				wiz.md.app.SetRoot(wiz.executionDockerModal, true)
			} else if buttonIndex == 1 {
				wiz.md.app.SetRoot(wiz.executionExternalModal, true)
			} else if buttonIndex == 2 {
				wiz.md.app.Stop()
			}
		})

	wiz.executionModeModal = modal

}


// Create the execution client Docker modal
func (wiz *newUserWizard) createExecutionDockerModal() {

	modal := NewDirectionalModal(DirectionalModalVertical, wiz.md.app).
		SetText("Please select the Execution client you would like to use.\n\n" +
			"Highlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth1-clients to learn more about them.",
		).
		AddButtons([]string{
			"Geth",
			"Infura",
			"Pocket",
			"Quit without Saving",
			"[More stuff for required params here]",
		}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 3 {
				wiz.md.app.Stop()
			} else {
				wiz.md.app.SetRoot(wiz.consensusDockerModal, true)
			}
		})

	wiz.executionDockerModal = modal

}


// Create the consensus client Docker modal
func (wiz *newUserWizard) createConsensusDockerModal() {

	modal := NewDirectionalModal(DirectionalModalVertical, wiz.md.app).
		SetText("Please select the Consensus client you would like to use.\n\n" +
			"Highlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth2-clients to learn more about them.",
		).
		AddButtons([]string{
			"Lighthouse",
			"Nimbus (Recommended)",
			"Prysm",
			"Teku (Recommended)",
			"Quit without Saving",
			"[More stuff for required params here]",
		}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 4 {
				wiz.md.app.Stop()
			} else {
				wiz.md.app.SetRoot(wiz.finishedModal, true)
			}
		})

	wiz.consensusDockerModal = modal

}


// Create the finished modal
func (wiz *newUserWizard) createFinishedModal() {

	modal := NewDirectionalModal(DirectionalModalVertical, wiz.md.app).
		SetText("All done! You're ready to run.\n\n" +
			"If you'd like, you can review and change all of the Smartnode and client settings next or just save and exit.",
		).
		AddButtons([]string{
			"Review All Settings",
			"Save and Exit",
		}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				wiz.md.showMainGrid()
			} else {
				wiz.md.app.Stop()
			}
		})

	wiz.finishedModal = modal

}