package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)


const settingsHomeId string = "settings-home"


// This is a container for the primary settings category selection home screen.
type settingsHome struct {
    homePage *page
    saveButton *tview.Button
    quitButton *tview.Button
    categoryList *tview.List
    settingsSubpages []*page
    content *tview.Box
    md *mainDisplay
}


// Creates a new SettingsHome instance and adds (and its subpages) it to the main display.
func newSettingsHome(md *mainDisplay) *settingsHome {
    
    homePage := newPage(nil, settingsHomeId, "Settings", "", nil)

    // Create the settings subpages
    settingsSubpages := []*page{
        createSettingSmartnodePage(homePage),
    }

    // Add the subpages to the main display
    for _, subpage := range settingsSubpages {
        md.pages.AddPage(subpage.id, subpage.content, true, false)
    }

    // Create the page and return it
    home := &settingsHome{
        settingsSubpages: settingsSubpages,
        md: md,
        homePage: homePage,
    }
    home.createContent()
    homePage.content = home.content
    md.pages.AddPage(settingsHomeId, home.content, true, false)
    return home

}


// Create the content for this page
func (home *settingsHome) createContent() {

    layout := newStandardLayout()

    // Create the category list
    categoryList := tview.NewList().
	    SetMainTextColor(tcell.ColorLightGreen).
        SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
            layout.descriptionBox.SetText(home.settingsSubpages[index].description)
        })
    home.categoryList = categoryList

    // Set tab to switch to the save and quit buttons
	categoryList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        if event.Key() == tcell.KeyTab {
            home.md.app.SetFocus(home.saveButton)
            return nil
        }
        return event
    })

    // Add all of the subpages to the list
    for _, subpage := range home.settingsSubpages {
        categoryList.AddItem(subpage.title, "", 0, func() {
            home.md.setPage(subpage)
        })
    }

    // Make it the content of the layout and set the default description text
    layout.setContent(categoryList.Box, "Select a Category")
    layout.descriptionBox.SetText(home.settingsSubpages[0].description)

    // Make the footer
    footer := home.createFooter()
    layout.setFooter(footer)

    // Set the home page's content to be the standard layout's grid
    home.content = layout.grid.Box

}


// Create the footer, including the nav bar and the save / quit buttons
func (home *settingsHome) createFooter() *tview.Box {

	// Nav bar
	navText := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)

	fmt.Fprintf(navText, "Arrow keys: Navigate   Enter: Select   Tab: Go to Save / Exit Buttons")

	navBar := tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(navText, 69, 1, false).
		AddItem(tview.NewBox(), 0, 1, false)

	// Save and Quit buttons
	saveButton := tview.NewButton("Save and Exit")
	quitButton := tview.NewButton("Quit without Saving")
    home.saveButton = saveButton
    home.quitButton = quitButton

	saveButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			home.md.app.SetFocus(home.categoryList)
			return nil
		} else if event.Key() == tcell.KeyRight ||
		event.Key() == tcell.KeyLeft {
			home.md.app.SetFocus(quitButton)
			return nil
		}
		return event
	})
	saveButton.SetSelectedFunc(func() {
        // TODO: SAVE
        home.md.app.Stop()
	})

	quitButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			home.md.app.SetFocus(home.categoryList)
			return nil
		} else if event.Key() == tcell.KeyRight ||
		event.Key() == tcell.KeyLeft {
			home.md.app.SetFocus(saveButton)
			return nil
		}
		return event
	})
	quitButton.SetSelectedFunc(func() {
		modal := tview.NewModal().
			SetText("Are you sure you want to quit?").
			AddButtons([]string{"Quit", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonIndex == 0 {
					home.md.app.Stop()
				} else if buttonIndex == 1 {
					home.md.app.SetRoot(home.md.pages, true)
				}
			})
        home.md.app.SetRoot(modal, true)
	})
	
	buttonBar := tview.NewFlex().
		AddItem(tview.NewBox(), 0, 3, false).
		AddItem(saveButton, 21, 1, false).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(quitButton, 21, 1, false).
		AddItem(tview.NewBox(), 0, 3, false)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(buttonBar, 1, 1, false).
		AddItem(tview.NewBox(), 1, 1, false).
		AddItem(navBar, 1, 1, false).
        Box

}