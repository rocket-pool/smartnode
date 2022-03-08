package config

import "github.com/rivo/tview"

type settingsPage struct {
	name        string
	description string
	pageId      string
	content     *tview.Box
}

type wizardStep interface {
	show()
}
