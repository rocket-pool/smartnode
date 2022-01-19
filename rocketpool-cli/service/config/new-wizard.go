package config

import (
	"github.com/rivo/tview"
)


type newUserWizard struct {
	md *mainDisplay
	welcomeModal *DirectionalModal
	executionModeModal *DirectionalModal
	executionDockerModal *DirectionalModal
	executionExternalModal *DirectionalModal
	consensusModeModal *tview.Modal
	consensusDockerModal *tview.Modal
	consensusExternalMoadl *tview.Modal
	finishedModal *tview.Modal
}


func newNewUserWizard(md *mainDisplay) *newUserWizard {

	wiz := &newUserWizard{
		md: md,
	}

	wiz.createWelcomeModal()
	wiz.createExecutionModeModal()

	return wiz

}


// Create the welcome modal
func (wiz *newUserWizard) createWelcomeModal() {

	welcomeModal := NewDirectionalModal(DirectionalModalHorizontal, wiz.md.app).
		SetText("Welcome to the Smartnode configuration wizard!\n\n" +
			"Since this is your first time configuring the Smartnode, we'll walk you through the basic setup.\n\n",
		).
		AddButtons([]string{"Next", "Quit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				wiz.md.app.SetRoot(wiz.executionModeModal, true)
			} else if buttonIndex == 1 {
				wiz.md.app.Stop()
			}
		})

	wiz.welcomeModal = welcomeModal

}


// Create the execution client mode selection modal
func (wiz *newUserWizard) createExecutionModeModal() {

	executionModeModal := NewDirectionalModal(DirectionalModalVertical, wiz.md.app).
		SetText("Let's start by choosing how you'd like to run your execution client (formerly eth1 client).\n\n" +
			"Would you like Rocket Pool to run and manage its own client, or would you like it to use an existing client you run and manage outside of Rocket Pool (formerly known as \"Hybrid Mode\")?",
		).
		AddButtons([]string{"Let Rocket Pool Manage its Own Client (Default)", "Use an Existing External Client (Hybrid Mode)", "Quit without Saving"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 0 {
				wiz.md.app.SetRoot(wiz.executionDockerModal, true)
			} else if buttonIndex == 1 {
				wiz.md.app.SetRoot(wiz.executionExternalModal, true)
			} else if buttonIndex == 2 {
				wiz.md.app.Stop()
			}
		})

	wiz.executionModeModal = executionModeModal
}