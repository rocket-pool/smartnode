package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/types/config"
)

const settingsHomeID string = "settings-home"

// This is a container for the primary settings category selection home screen.
type settingsHome struct {
	homePage         *page
	saveButton       *tview.Button
	wizardButton     *tview.Button
	smartnodePage    *SmartnodeConfigPage
	ecPage           *ExecutionConfigPage
	fallbackPage     *FallbackConfigPage
	ccPage           *ConsensusConfigPage
	mevBoostPage     *MevBoostConfigPage
	metricsPage      *MetricsConfigPage
	addonsPage       *AddonsPage
	categoryList     *tview.List
	settingsSubpages []settingsPage
	content          tview.Primitive
	md               *mainDisplay
}

// Creates a new SettingsHome instance and adds (and its subpages) it to the main display.
func newSettingsHome(md *mainDisplay) *settingsHome {

	homePage := newPage(nil, settingsHomeID, "Categories", "", nil)

	// Create the page and return it
	home := &settingsHome{
		md:       md,
		homePage: homePage,
	}

	// Create the settings subpages
	home.smartnodePage = NewSmartnodeConfigPage(home)
	home.ecPage = NewExecutionConfigPage(home)
	home.ccPage = NewConsensusConfigPage(home)
	home.fallbackPage = NewFallbackConfigPage(home)
	home.mevBoostPage = NewMevBoostConfigPage(home)
	home.metricsPage = NewMetricsConfigPage(home)
	home.addonsPage = NewAddonsPage(home)
	settingsSubpages := []settingsPage{
		home.smartnodePage,
		home.ecPage,
		home.ccPage,
		home.fallbackPage,
		home.mevBoostPage,
		home.metricsPage,
		home.addonsPage,
	}
	home.settingsSubpages = settingsSubpages

	// Add the subpages to the main display
	for _, subpage := range settingsSubpages {
		md.pages.AddPage(subpage.getPage().id, subpage.getPage().content, true, false)
	}
	home.createContent()
	homePage.content = home.content
	md.pages.AddPage(homePage.id, home.content, true, false)
	return home

}

// Create the content for this page
func (home *settingsHome) createContent() {

	layout := newStandardLayout()

	// Create the category list
	categoryList := tview.NewList().
		SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			if (home.md.Config.Smartnode.Network.Value == config.Network_Holesky || home.md.Config.Smartnode.Network.Value == config.Network_Devnet) && home.settingsSubpages[index].getPage().id == "settings-mev-boost" {
				// Disable MEV-Boost for Holesky
				layout.descriptionBox.SetText("MEV-Boost is currently disabled for the Holesky test network.")
			} else {
				layout.descriptionBox.SetText(home.settingsSubpages[index].getPage().description)
			}
		})
	categoryList.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	categoryList.SetBorderPadding(0, 0, 1, 1)
	home.categoryList = categoryList

	// Set tab to switch to the save and quit buttons
	categoryList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab {
			home.md.app.SetFocus(home.saveButton)
			return nil
		}
		return event
	})

	// Add all of the subpages to the list
	for _, subpage := range home.settingsSubpages {
		categoryList.AddItem(subpage.getPage().title, "", 0, nil)
	}
	categoryList.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		if (home.md.Config.Smartnode.Network.Value == config.Network_Holesky || home.md.Config.Smartnode.Network.Value == config.Network_Devnet) && home.settingsSubpages[i].getPage().id == "settings-mev-boost" {
			// Disable MEV-Boost for Holesky
			return
		} else {
			home.settingsSubpages[i].handleLayoutChanged()
			home.md.setPage(home.settingsSubpages[i].getPage())
		}
	})

	// Make it the content of the layout and set the default description text
	layout.setContent(categoryList, categoryList.Box, "Select a Category")
	layout.descriptionBox.SetText(home.settingsSubpages[0].getPage().description)

	// Make the footer
	footer, height := home.createFooter()
	layout.setFooter(footer, height)

	// Set the home page's content to be the standard layout's grid
	home.content = layout.grid

}

// Create the footer, including the nav bar and the save / quit buttons
func (home *settingsHome) createFooter() (tview.Primitive, int) {

	// Nav bar
	navString1 := "Arrow keys: Navigate             Space/Enter: Select"
	navTextView1 := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	navBar1 := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(navTextView1, len(navString1), 1, false).
		AddItem(nil, 0, 1, false)
	fmt.Fprint(navTextView1, navString1)

	navString2 := "Tab: Go to the Buttons   Ctrl+C: Quit without Saving"
	navTextView2 := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	navBar2 := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(navTextView2, len(navString2), 1, false).
		AddItem(nil, 0, 1, false)
	fmt.Fprint(navTextView2, navString2)

	// Save and Quit buttons
	saveButton := tview.NewButton("Review Changes and Save")
	wizardButton := tview.NewButton("Open the Config Wizard")
	home.saveButton = saveButton
	home.wizardButton = wizardButton

	saveButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab {
			home.md.app.SetFocus(home.categoryList)
			return nil
		} else if event.Key() == tcell.KeyRight ||
			event.Key() == tcell.KeyLeft ||
			event.Key() == tcell.KeyUp ||
			event.Key() == tcell.KeyDown {
			home.md.app.SetFocus(wizardButton)
			return nil
		}
		return event
	})
	saveButton.SetSelectedFunc(func() {
		home.md.pages.RemovePage(reviewPageID)
		reviewPage := NewReviewPage(home.md, home.md.PreviousConfig, home.md.Config)
		home.md.pages.AddPage(reviewPage.page.id, reviewPage.page.content, true, true)
		home.md.setPage(reviewPage.page)
	})
	saveButton.SetBackgroundColorActivated(tcell.Color46)
	saveButton.SetLabelColorActivated(tcell.ColorBlack)

	wizardButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab {
			home.md.app.SetFocus(home.categoryList)
			return nil
		} else if event.Key() == tcell.KeyRight ||
			event.Key() == tcell.KeyLeft ||
			event.Key() == tcell.KeyUp ||
			event.Key() == tcell.KeyDown {
			home.md.app.SetFocus(saveButton)
			return nil
		}
		return event
	})
	wizardButton.SetSelectedFunc(func() {
		home.md.dockerWizard.welcomeModal.show()
	})
	wizardButton.SetBackgroundColorActivated(tcell.Color46)
	wizardButton.SetLabelColorActivated(tcell.ColorBlack)

	// Create overall layout for the footer
	buttonBar := tview.NewFlex().
		AddItem(nil, 0, 3, false).
		AddItem(saveButton, 25, 1, false).
		AddItem(nil, 0, 1, false).
		AddItem(wizardButton, 24, 1, false).
		AddItem(nil, 0, 3, false)

	footer := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(buttonBar, 1, 1, false).
		AddItem(nil, 1, 1, false).
		AddItem(navBar1, 1, 1, false).
		AddItem(navBar2, 1, 1, false)

	return footer, footer.GetItemCount()

}

// Refreshes the settings on all of the config pages to match the config's values
func (home *settingsHome) refresh() {
	/*
		if home.smartnodePage != nil {
			home.smartnodePage.layout.refresh()
		}*/

	if home.ecPage != nil {
		home.ecPage.layout.refresh()
	}

	if home.ccPage != nil {
		home.ccPage.layout.refresh()
	}

	if home.fallbackPage != nil {
		home.fallbackPage.layout.refresh()
	}

	if home.fallbackPage != nil {
		home.fallbackPage.layout.refresh()
	}

	if home.mevBoostPage != nil {
		home.mevBoostPage.layout.refresh()
	}

	if home.metricsPage != nil {
		home.metricsPage.layout.refresh()
	}
}
