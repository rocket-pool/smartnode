package config

import (
	"github.com/rivo/tview"
)


func createNewUserWelcomeModal(md *mainDisplay) *tview.Modal {

	// TODO - this is not the right place to put this, temporary only!
	page := createNewUserExecutionPage(md.app)
	md.pages.AddPage(page.id, page.content, true, false)

	return tview.NewModal().
		SetText("Welcome to the Smartnode configuration wizard!\n\n" +
			"Since this is your first time configuring the Smartnode, we'll walk you through the basic setup.",
		).
		AddButtons([]string{"Ok", "Quit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 1 {
				md.app.Stop()
			} else if buttonIndex == 0 {
				md.showMainGrid()
				md.setPage(page)
				md.app.SetFocus(page.content)
			}
		})

}