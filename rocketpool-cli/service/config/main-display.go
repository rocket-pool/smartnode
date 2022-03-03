package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// This represents the primary TUI for the configuration command
type mainDisplay struct {
	navHeader           *tview.TextView
	pages               *tview.Pages
	app                 *tview.Application
	content             *tview.Box
	mainGrid            *tview.Grid
	newUserWizard       *newUserWizard
	settingsHome        *settingsHome
	PreviousConfig      *config.RocketPoolConfig
	Config              *config.RocketPoolConfig
	ShouldSave          bool
	ContainersToRestart []config.ContainerID
	ChangeNetworks      bool
}

// Creates a new MainDisplay instance.
func NewMainDisplay(app *tview.Application, config *config.RocketPoolConfig, isNew bool) *mainDisplay {

	// Create a copy of the original config for comparison purposes
	previousConfig := config.CreateCopy()

	// Create the main grid
	grid := tview.NewGrid().
		SetColumns(1, 0, 1).   // 1-unit border
		SetRows(1, 1, 1, 0, 1) // Also 1-unit border

	grid.SetBorder(true).
		SetTitle(fmt.Sprintf(" Rocket Pool Smartnode %s Configuration ", shared.RocketPoolVersion)).
		SetBorderColor(tcell.ColorOrange).
		SetTitleColor(tcell.ColorOrange).
		SetBackgroundColor(tcell.ColorBlack)

	// Create the navigation header
	navHeader := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	grid.AddItem(navHeader, 1, 1, 1, 1, 0, 0, false)

	// Create the page collection
	pages := tview.NewPages()
	grid.AddItem(pages, 3, 1, 1, 1, 0, 0, true)

	// Create the main display object
	md := &mainDisplay{
		navHeader:      navHeader,
		pages:          pages,
		app:            app,
		content:        grid.Box,
		mainGrid:       grid,
		PreviousConfig: previousConfig,
		Config:         config,
	}

	// Create all of the child elements
	md.settingsHome = newSettingsHome(md)
	md.newUserWizard = newNewUserWizard(md)

	if isNew {
		md.setPage(md.newUserWizard.welcomeModal)
	} else {
		md.setPage(md.settingsHome.homePage)
	}
	app.SetRoot(grid, true)
	return md

}

// Sets the current page that is on display.
func (md *mainDisplay) setPage(page *page) {
	md.navHeader.SetText(page.getHeader())
	md.pages.SwitchToPage(page.id)
}
