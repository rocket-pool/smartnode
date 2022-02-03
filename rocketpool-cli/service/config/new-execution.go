package config

import (
	"github.com/rivo/tview"
)

func createNewUserExecutionPage(app *tview.Application) *page {

	content := createNewUserExecutionContent()

	return newPage(
		nil,
		"new-user-execution",
		"Step 1: Execution Client",
		"",
		content,
	)

}

func createNewUserExecutionContent() tview.Primitive {

	layout := newStandardLayout()

	// Create the intro helper text
	intro := "Let's start by picking which Execution client (formerly \"eth1\" client) you'd like to use.\n" +
		"Choose one from the dropdown below by pressing [Space] or [Enter] while it's selected, and navigate between the options using the arrow keys.\n" +
		"As you hover on each choice, the Description box to the right will tell you a little about that client.\n" +
		"Press [Space] or [Enter] to choose a client while hovering on it."

	introTextView := tview.NewTextView().
		SetWordWrap(true).
		SetText(intro)

	form := tview.NewForm()

	// Create the external client toggles
	rpClientCheckbox := tview.NewCheckbox().
		SetLabel("I want Rocket Pool to manage my Execution client")
	rpClientCheckbox.SetFocusFunc(func() {
		layout.descriptionBox.SetText("Choose this if you don't already have your own Execution client running, and you want Rocket Pool to create and manage one for you.")
	})
	form.AddFormItem(rpClientCheckbox)

	externalClientCheckbox := tview.NewCheckbox().
		SetLabel("I want to use an existing Execution client that I manage outside of Rocket Pool")
	externalClientCheckbox.SetFocusFunc(func() {
		layout.descriptionBox.SetText("Choose this if you have your own Execution client installed and running on this machine or another one, and you want Rocket Pool to link to that instead of running its own client.")
	})
	form.AddFormItem(externalClientCheckbox)

	// Create the client choice dropdown
	clientDescriptions := []string{
		"Geth is one of the three original implementations of the Ethereum protocol. It is written in Go, fully open source and licensed under the GNU LGPL v3.",
		"Use infura.io as a light client for Eth 1.0. Not recommended for use in production.",
		"Use Pocket Network as a decentralized light client for Eth 1.0. Suitable for use in production.",
	}

	clientDropdown := tview.NewDropDown().
		SetLabel("Client").
		SetOptions([]string{"Geth", "Infura", "Pocket"}, nil)
	clientDropdown.SetFocusFunc(func() {
		layout.descriptionBox.SetText("Select which Execution client you'd like to use from this list.")
	})
	clientDropdown.SetSelectedFunc(func(text string, index int) {
		layout.descriptionBox.SetText(clientDescriptions[index])
	})
	clientDropdown.SetTextOptions(" ", " ", "", "", "")
	form.AddFormItem(clientDropdown)

	form.AddButton("Additional Options", func() {
		form.AddInputField("Placeholder1", "", 0, nil, nil)
		form.AddInputField("Placeholder2", "", 0, nil, nil)
		form.AddInputField("Placeholder3", "", 0, nil, nil)
		form.AddInputField("Placeholder4", "", 0, nil, nil)
		form.AddInputField("Placeholder5", "", 0, nil, nil)
		form.AddInputField("Placeholder6", "", 0, nil, nil)
		form.AddInputField("Placeholder7", "", 0, nil, nil)
		form.AddInputField("Placeholder8", "", 0, nil, nil)
		form.AddInputField("Placeholder9", "", 0, nil, nil)
		form.AddInputField("Placeholder0", "", 0, nil, nil)
	})

	contentFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(introTextView, 0, 3, false).
		AddItem(form, 0, 5, true)

	layout.setContent(contentFlex, contentFlex.Box, "Step 1: Execution Client")
	layout.setFooter(nil, 0)

	return layout.grid

}
