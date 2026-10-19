package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// EpbsPage groups shared builder settings and the existing PBS clients.
type EpbsPage struct {
	home            *settingsHome
	page            *page
	layout          *standardLayout
	categoryList    *tview.List
	builderPage     *BuilderConfigPage
	mevBoostPage    *MevBoostConfigPage
	commitBoostPage *CommitBoostConfigPage
	subpages        []settingsPage
}

func NewEpbsPage(home *settingsHome) *EpbsPage {
	epbsPage := &EpbsPage{home: home}
	epbsPage.page = newPage(
		home.homePage,
		"settings-epbs",
		"ePBS",
		"Configure builder settings shared by all consensus clients, and manage MEV-Boost and Commit-Boost.",
		nil,
	)
	epbsPage.builderPage = NewBuilderConfigPage(epbsPage)
	epbsPage.mevBoostPage = NewMevBoostConfigPage(epbsPage)
	epbsPage.commitBoostPage = NewCommitBoostConfigPage(epbsPage)
	epbsPage.subpages = []settingsPage{
		epbsPage.builderPage,
		epbsPage.mevBoostPage,
		epbsPage.commitBoostPage,
	}
	for _, subpage := range epbsPage.subpages {
		home.md.pages.AddPage(subpage.getPage().id, subpage.getPage().content, true, false)
	}
	epbsPage.createContent()
	epbsPage.page.content = epbsPage.layout.grid
	return epbsPage
}

func (epbsPage *EpbsPage) getPage() *page {
	return epbsPage.page
}

func (epbsPage *EpbsPage) createContent() {
	epbsPage.layout = newStandardLayout()
	epbsPage.layout.createSettingFooter()
	list := tview.NewList().SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		epbsPage.layout.descriptionBox.SetText(epbsPage.subpages[index].getPage().description)
	})
	list.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	list.SetBorderPadding(0, 0, 1, 1)
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			epbsPage.home.md.setPage(epbsPage.home.homePage)
			return nil
		}
		return event
	})
	for _, subpage := range epbsPage.subpages {
		list.AddItem(subpage.getPage().title, "", 0, nil)
	}
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		subpage := epbsPage.subpages[index]
		subpage.handleLayoutChanged()
		epbsPage.home.md.setPage(subpage.getPage())
	})
	epbsPage.categoryList = list
	epbsPage.layout.setContent(list, list.Box, "ePBS Settings")
	epbsPage.layout.descriptionBox.SetText(epbsPage.subpages[0].getPage().description)
}

func (epbsPage *EpbsPage) handleLayoutChanged() {
	for _, subpage := range epbsPage.subpages {
		subpage.handleLayoutChanged()
	}
}
