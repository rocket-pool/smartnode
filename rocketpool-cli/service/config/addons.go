package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/config"
)

// Constants
const addonPageID string = "addons"

// The addons page
type AddonsPage struct {
	home             *settingsHome
	page             *page
	layout           *standardLayout
	masterConfig     *config.RocketPoolConfig
	gwwPage          *AddonGwwPage
	gwwButton        *parameterizedFormItem
	rescueNodePage   *AddonRescueNodePage
	rescueNodeButton *parameterizedFormItem
	categoryList     *tview.List
	addonSubpages    []settingsPage
	content          tview.Primitive
}

// Create a new addons page
func NewAddonsPage(home *settingsHome) *AddonsPage {

	addonsPage := &AddonsPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	addonsPage.page = newPage(
		home.homePage,
		addonPageID,
		"Addons",
		"Manage custom services that can run alongside Rocket Pool, built by our community to enhance your Node Operator experience.",
		nil,
	)

	// Create the addon subpages
	addonsPage.gwwPage = NewAddonGwwPage(addonsPage, home.md.Config.GraffitiWallWriter)
	addonsPage.rescueNodePage = NewAddonRescueNodePage(addonsPage, home.md.Config.RescueNode)
	addonSubpages := []settingsPage{
		addonsPage.gwwPage,
		addonsPage.rescueNodePage,
	}
	addonsPage.addonSubpages = addonSubpages

	// Add the subpages to the main display
	for _, subpage := range addonSubpages {
		home.md.pages.AddPage(subpage.getPage().id, subpage.getPage().content, true, false)
	}

	addonsPage.createContent()
	addonsPage.page.content = addonsPage.layout.grid
	return addonsPage

}

// Get the underlying page
func (addonsPage *AddonsPage) getPage() *page {
	return addonsPage.page
}

// Creates the content for the fallback client settings page
func (addonsPage *AddonsPage) createContent() {

	addonsPage.layout = newStandardLayout()
	addonsPage.layout.createSettingFooter()

	// Create the category list
	categoryList := tview.NewList().
		SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			addonsPage.layout.descriptionBox.SetText(addonsPage.addonSubpages[index].getPage().description)
		})
	categoryList.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	categoryList.SetBorderPadding(0, 0, 1, 1)
	addonsPage.categoryList = categoryList

	// Set tab to switch to the save and quit buttons
	categoryList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			// Return to the home page
			addonsPage.home.md.setPage(addonsPage.home.homePage)
			return nil
		}
		return event
	})

	// Add all of the subpages to the list
	for _, subpage := range addonsPage.addonSubpages {
		categoryList.AddItem(subpage.getPage().title, "", 0, nil)
	}
	categoryList.SetSelectedFunc(func(i int, s1, s2 string, r rune) {
		addonsPage.addonSubpages[i].handleLayoutChanged()
		addonsPage.home.md.setPage(addonsPage.addonSubpages[i].getPage())
	})

	// Make it the content of the layout and set the default description text
	addonsPage.layout.setContent(categoryList, categoryList.Box, "Select an Addon")
	addonsPage.layout.descriptionBox.SetText(addonsPage.addonSubpages[0].getPage().description)

	// Make the footer
	//footer, height := addonsPage.createFooter()
	//layout.setFooter(footer, height)

	// Set the home page's content to be the standard layout's grid
	//addonsPage.content = layout.grid
}

// Handle a bulk redraw request
func (addonsPage *AddonsPage) handleLayoutChanged() {

}
