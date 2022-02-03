package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Creates a new page for the Smartnode settings
func createSettingSmartnodePage(home *settingsHome) *page {

	content := createSettingSmartnodeContent(home)

	return newPage(
		home.homePage,
		"settings-smartnode",
		"Smartnode and TX Fees",
		"Select this to configure the settings for the Smartnode itself, including the defaults and limits on transaction fees.",
		content,
	)

}

// Creates the content for the Smartnode settings page
func createSettingSmartnodeContent(home *settingsHome) tview.Primitive {

	layout := newStandardLayout()

	// PLACEHOLDER
	paramDescriptions := []string{
		"The Execution client you'd like to use. Probably have to describe each one when you open this dropdown and hover over them.",
		"Select this if you have an external Execution client that you want the Smartnode to use, instead of managing its own (\"Hybrid Mode\").",
		"Enter Geth's cache size, in MB.",
	}

	// Create the settings form
	form := tview.NewForm()
	a := tview.NewDropDown().
		SetLabel("Client").
		SetOptions([]string{"Geth", "Infura", "Pocket", "Custom"}, nil)
	a.SetFocusFunc(func() {
		layout.descriptionBox.SetText(paramDescriptions[0])
	})
	a.SetTextOptions(" ", " ", "", "", "")
	form.AddFormItem(a)

	b := tview.NewCheckbox().
		SetLabel("Externally managed?")
	b.SetFocusFunc(func() {
		layout.descriptionBox.SetText(paramDescriptions[1])
	})
	form.AddFormItem(b)

	c := tview.NewInputField().
		SetLabel("Geth Cache (MB)").
		SetText("1024")
	c.SetFocusFunc(func() {
		layout.descriptionBox.SetText(paramDescriptions[2])
	})
	form.AddFormItem(c)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			home.md.setPage(home.homePage)
			return nil
		}
		return event
	})

	// Make it the content of the layout and set the default description text
	layout.setContent(form, form.Box, "Execution Client (Eth1) Settings")
	layout.descriptionBox.SetText(paramDescriptions[0])

	// Make the footer
	footer, height := createSettingFooter()
	layout.setFooter(footer, height)

	// Return the standard layout's grid
	return layout.grid

}

// Create the footer, including the nav bar and the save / quit buttons
func createSettingFooter() (*tview.Flex, int) {

	// Nav bar
	navString1 := "Tab: Next Setting   Shift-Tab: Previous Setting"
	navTextView1 := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	fmt.Fprint(navTextView1, navString1)

	navString2 := "Space/Enter: Change Setting   Esc: Done, Return to Categories"
	navTextView2 := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	fmt.Fprint(navTextView2, navString2)

	navBar := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(navTextView1, len(navString1), 1, false).
			AddItem(tview.NewBox(), 0, 1, false),
			1, 1, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(navTextView2, len(navString2), 1, false).
			AddItem(tview.NewBox(), 0, 1, false),
			1, 1, false)

	return navBar, 2

}
