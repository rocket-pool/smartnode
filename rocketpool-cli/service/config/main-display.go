package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// This represents the primary TUI for the configuration command
type mainDisplay struct {
	navHeader           *tview.TextView
	pages               *tview.Pages
	app                 *tview.Application
	content             *tview.Box
	mainGrid            *tview.Grid
	dockerWizard        *wizard
	settingsHome        *settingsHome
	settingsNativeHome  *settingsNativeHome
	isNew               bool
	isUpdate            bool
	isNative            bool
	previousWidth       int
	previousHeight      int
	PreviousConfig      *config.RocketPoolConfig
	Config              *config.RocketPoolConfig
	ShouldSave          bool
	ContainersToRestart []cfgtypes.ContainerID
	ChangeNetworks      bool
}

const (
	// Background for non-interactive elements
	NonInteractiveBackgroundColor tcell.Color = tcell.ColorBlack
)

// Creates a new MainDisplay instance.
func NewMainDisplay(app *tview.Application, previousConfig *config.RocketPoolConfig, config *config.RocketPoolConfig, isNew bool, isUpdate bool, isNative bool) *mainDisplay {

	// Create a copy of the original config for comparison purposes
	if previousConfig == nil {
		previousConfig = config.CreateCopy()
	}

	// Create the main grid
	grid := tview.NewGrid().
		SetColumns(1, 0, 1).   // 1-unit border
		SetRows(1, 1, 1, 0, 1) // Also 1-unit border

	grid.SetBackgroundColor(NonInteractiveBackgroundColor)

	grid.SetBorder(true).
		SetTitle(fmt.Sprintf(" Rocket Pool Smart Node %s Configuration ", shared.RocketPoolVersion())).
		SetBorderColor(tcell.ColorOrange).
		SetTitleColor(tcell.ColorOrange).
		SetBackgroundColor(NonInteractiveBackgroundColor)

	// Padding
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 0, 2, 1, 1, 0, 0, false)

	// Create the navigation header
	navHeader := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	grid.AddItem(navHeader, 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 1, 0, 1, 1, 0, 0, false)
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 1, 2, 1, 1, 0, 0, false)

	// Padding
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 2, 0, 1, 1, 0, 0, false)
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 2, 1, 1, 1, 0, 0, false)
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 2, 2, 1, 1, 0, 0, false)

	// Create the page collection
	pages := tview.NewPages()
	grid.AddItem(pages, 3, 1, 1, 1, 0, 0, true)
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 3, 0, 1, 1, 0, 0, false)
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 3, 2, 1, 1, 0, 0, false)

	// Padding
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 4, 0, 1, 1, 0, 0, false)
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 4, 1, 1, 1, 0, 0, false)
	grid.AddItem(tview.NewBox().SetBackgroundColor(NonInteractiveBackgroundColor), 4, 2, 1, 1, 0, 0, false)

	// Create the resize warning
	resizeWarning := tview.NewTextView().
		SetText("Your terminal is too small to run the service configuration app.\n\nPlease resize your terminal window and make it larger to see the app properly.").
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true).
		SetTextColor(tview.Styles.PrimaryTextColor)
	resizeWarning.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	resizeWarning.SetBorderPadding(0, 0, 1, 1)

	// Create the main display object
	md := &mainDisplay{
		navHeader:      navHeader,
		pages:          pages,
		app:            app,
		content:        grid.Box,
		mainGrid:       grid,
		isNew:          isNew,
		isUpdate:       isUpdate,
		isNative:       isNative,
		PreviousConfig: previousConfig,
		Config:         config,
	}

	// Create all of the child elements
	md.settingsHome = newSettingsHome(md)
	md.settingsNativeHome = newSettingsNativeHome(md)
	md.dockerWizard = newWizard(md)

	// Set up the resize warning
	md.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		x, y := screen.Size()
		if x == md.previousWidth && y == md.previousHeight {
			return false
		}
		if x < 112 || y < 32 {
			grid.RemoveItem(pages)
			grid.AddItem(resizeWarning, 3, 1, 1, 1, 0, 0, false)
		} else {
			grid.RemoveItem(resizeWarning)
			grid.AddItem(pages, 3, 1, 1, 1, 0, 0, true)
		}
		md.previousWidth = x
		md.previousHeight = y

		return false
	})

	if isNew {
		if isNative {
			md.dockerWizard.nativeWelcomeModal.show()
		} else {
			md.dockerWizard.welcomeModal.show()
		}
	} else {
		if isNative {
			md.setPage(md.settingsNativeHome.homePage)
		} else {
			md.setPage(md.settingsHome.homePage)
		}
	}
	app.SetRoot(grid, true)
	return md

}

// Sets the current page that is on display.
func (md *mainDisplay) setPage(page *page) {
	md.navHeader.SetText(page.getHeader())
	md.pages.SwitchToPage(page.id)
}
