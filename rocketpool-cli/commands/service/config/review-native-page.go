package config

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/node-manager-core/config"
	snCfg "github.com/rocket-pool/smartnode/v2/shared/config"
)

// Constants
const reviewNativePageID string = "review-native-settings"

// Create a page to review any changes
func NewReviewNativePage(md *mainDisplay, oldConfig *snCfg.SmartNodeConfig, newConfig *snCfg.SmartNodeConfig) *ReviewPage {
	var changedSettings []*config.ChangedSection

	// Create the visual list for all of the changed settings
	changeBox := tview.NewTextView().
		SetDynamicColors(true)
	changeBox.SetBorder(true)
	changeBox.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	changeBox.SetBorderPadding(0, 0, 1, 1)

	builder := strings.Builder{}
	errors := newConfig.Validate()
	if len(errors) > 0 {
		builder.WriteString("[orange]WARNING: Your configuration encountered errors. You must correct the following in order to save it:\n\n")
		for _, err := range errors {
			builder.WriteString(fmt.Sprintf("%s\n\n", err))
		}
	} else {
		// Get the map of changed settings by category
		changedSettings, _, _ = newConfig.GetChanges(oldConfig)

		// Get the map of changed settings by section name
		if len(changedSettings) > 0 {
			for _, change := range changedSettings {
				addChangesToDescription(change, "", &builder)
			}
		}

		if builder.String() == "" {
			builder.WriteString("<No changes>")
		}
	}
	changeBox.SetText(builder.String())

	// Create the layout
	width := 86

	// Create the save button
	saveButton := tview.NewButton("Save Settings")
	saveButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			changeBox.InputHandler()(event, nil)
			return nil
		}
		return event
	})
	// Save when selected
	saveButton.SetSelectedFunc(func() {
		md.ShouldSave = true
		md.app.Stop()
	})
	saveButton.SetBackgroundColorActivated(tcell.Color46)
	saveButton.SetLabelColorActivated(tcell.ColorBlack)

	buttonGrid := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox().
			SetBackgroundColor(tview.Styles.ContrastBackgroundColor), 0, 1, false).
		AddItem(saveButton, len(saveButton.GetLabel())+2, 0, true).
		AddItem(tview.NewBox().
			SetBackgroundColor(tview.Styles.ContrastBackgroundColor), 0, 1, false)

	return reviewPage(md, md.settingsNativeHome.homePage, width, changeBox, buttonGrid, changedSettings)
}
