package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Constants
const addonPageID string = "addons"

// The addons page
type AddonsPage struct {
	md   *mainDisplay
	page *page
}

// Create a new addons page
func NewAddonsPage(md *mainDisplay) *AddonsPage {

	// Create the layout
	width := 86

	// Create the main text view
	descriptionText := "Coming soon!\n\nThis will be a page where you can configure custom, community-made Docker-based extensions to Rocket Pool. We'll populate this space with them as more and more are developed.\n\nIf you would like to develop an extension for Rocket Pool's Smartnode, please let us know on our Discord server (https://discord.gg/rocketpool) or on our Governance Forum (https://dao.rocketpool.net/)."
	lines := tview.WordWrap(descriptionText, width-4)
	textViewHeight := len(lines) + 1
	textView := tview.NewTextView().
		SetText(descriptionText).
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true).
		SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	textView.SetBorderPadding(0, 0, 1, 1)

	// Row spacers with the correct background color
	spacer1 := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacer2 := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacerL := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacerR := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	// The main content grid
	contentGrid := tview.NewGrid().
		SetRows(1, textViewHeight, 0).
		SetColumns(1, 0, 1).
		AddItem(spacer1, 0, 1, 1, 1, 0, 0, false).
		AddItem(textView, 1, 1, 1, 1, 0, 0, false).
		AddItem(spacer2, 2, 1, 1, 1, 0, 0, false).
		AddItem(spacerL, 0, 0, 3, 1, 0, 0, false).
		AddItem(spacerR, 0, 2, 3, 1, 0, 0, false)
	contentGrid.
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorder(true).
		SetTitle(" Addons ")
	contentGrid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			md.setPage(md.settingsHome.homePage)
			return nil
		default:
			return event
		}
	})

	// A grid with variable spaced borders that surrounds the fixed-size content grid
	borderGrid := tview.NewGrid().
		SetColumns(0, width, 0)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	// Get the total content height, including spacers and borders
	borderGrid.SetRows(1, 0, 1, 1, 1)

	// Create the nav footer text view
	navString1 := "Arrow keys: Navigate     Space/Enter: Select"
	navTextView1 := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	fmt.Fprint(navTextView1, navString1)

	navString2 := "Esc: Go Back     Ctrl+C: Quit without Saving"
	navTextView2 := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	fmt.Fprint(navTextView2, navString2)

	// Create the nav footer
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
	borderGrid.AddItem(navBar, 3, 1, 1, 1, 0, 0, true)

	page := newPage(nil, reviewPageID, "Addons", "Manage custom services that can run alongside Rocket Pool, built by our community to enhance your Node Operator experience.", borderGrid)

	return &AddonsPage{
		md:   md,
		page: page,
	}

}

// Get the underlying page
func (configPage *AddonsPage) getPage() *page {
	return configPage.page
}

// Handle a bulk redraw request
func (configPage *AddonsPage) handleLayoutChanged() {

}
