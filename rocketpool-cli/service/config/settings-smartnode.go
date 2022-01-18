package config

import "github.com/rivo/tview"

const settingsSmartnodeId string = "settings-smartnode"

func createSettingSmartnodePage(parent *page) *page {

    content := tview.NewBox() // Placeholder

    return newPage(
        parent, 
        settingsSmartnodeId, 
        "Smartnode and TX Fees",
        "Select this to configure the settings for the Smartnode itself, including the default s and limits on transaction fees.",
        content,
    )

}