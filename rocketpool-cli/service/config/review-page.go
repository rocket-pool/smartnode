package config

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// Constants
const reviewPageID string = "review-settings"

// A setting that has changed
type SettingsPair struct {
	Name               string
	OldValue           string
	NewValue           string
	AffectedContainers map[config.ContainerID]bool
}

// The changed settings review page
type ReviewPage struct {
	md              *mainDisplay
	changedSettings map[string][]SettingsPair
	page            *page
}

// Create a page to review any changes
func NewReviewPage(md *mainDisplay, oldConfig *config.RocketPoolConfig, newConfig *config.RocketPoolConfig) *ReviewPage {

	// Get the map of changed settings by category
	changedSettings := getChangedSettingsMap(oldConfig, newConfig)

	// Handle network changes (because it's a special case)
	changeNetworks := false
	if oldConfig.Smartnode.Network.Value != newConfig.Smartnode.Network.Value {
		changeNetworks = true
	}

	// Create a list of all of the container IDs that need to be restarted
	totalAffectedContainers := map[config.ContainerID]bool{}
	for _, settingList := range changedSettings {
		for _, setting := range settingList {
			for container := range setting.AffectedContainers {
				totalAffectedContainers[container] = true
			}
		}
	}

	// Create the visual list for all of the changed settings
	changeBox := tview.NewTextView().
		SetDynamicColors(true)
	changeBox.SetBorder(true)
	changeBox.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	changeBox.SetBorderPadding(0, 0, 1, 1)
	builder := strings.Builder{}

	if md.isUpdate || md.isMigration {
		totalAffectedContainers[config.ContainerID_Api] = true
		totalAffectedContainers[config.ContainerID_Node] = true
		totalAffectedContainers[config.ContainerID_Watchtower] = true

		if newConfig.ExecutionClientMode.Value.(config.Mode) == config.Mode_Local && newConfig.ExecutionClient.Value.(config.ExecutionClient) != config.ExecutionClient_Geth {
			totalAffectedContainers[config.ContainerID_Eth1] = true
		}
		if newConfig.FallbackExecutionClientMode.Value.(config.Mode) == config.Mode_Local {
			totalAffectedContainers[config.ContainerID_Eth1Fallback] = true
		}
		builder.WriteString(fmt.Sprintf("Updated to Smartnode v%s (will affect several containers)\n\n", shared.RocketPoolVersion))
	}

	for categoryName, changedSettingsList := range changedSettings {
		if len(changedSettingsList) > 0 {
			builder.WriteString(fmt.Sprintf("%s\n", categoryName))
			for _, pair := range changedSettingsList {
				builder.WriteString(fmt.Sprintf("\t%s: %s => %s\n", pair.Name, pair.OldValue, pair.NewValue))
			}
			builder.WriteString("\n")
		}
	}

	containersToRestart := []config.ContainerID{}
	if builder.String() == "" {
		builder.WriteString("<No changes>")
	} else {
		builder.WriteString("The following containers will be restarted for these changes to take effect:")
		for container, _ := range totalAffectedContainers {
			builder.WriteString(fmt.Sprintf("\n\t%v", container))
			containersToRestart = append(containersToRestart, container)
		}
	}
	changeBox.SetText(builder.String())

	// Create the layout
	width := 86

	// Create the main text view
	descriptionText := "Please review your changes below.\nScroll through them using the arrow keys, and press Enter when you're ready to save them and restart the relevant Docker containers."
	lines := tview.WordWrap(descriptionText, width-4)
	textViewHeight := len(lines) + 1
	textView := tview.NewTextView().
		SetText(descriptionText).
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true).
		SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	textView.SetBorderPadding(0, 0, 1, 1)

	// Create the save button
	saveButton := tview.NewButton("Save Settings and Restart Containers")
	saveButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			changeBox.InputHandler()(event, nil)
			return nil
		} else {
			return event
		}
	})
	// Save when selected
	saveButton.SetSelectedFunc(func() {
		if changeNetworks {
			// Network changes need to be handled specially
			modal := tview.NewModal().
				SetText("WARNING: You have requested to change networks.\n\nAll of your existing chain data, your node wallet, and your validator keys will be removed.\n\nPlease confirm you have backed up everything you want to keep, because it will be deleted once you save and run these changes!").
				AddButtons([]string{"Cancel!", "Ok, I'm Ready"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonIndex == 1 {
						md.ShouldSave = true
						md.ChangeNetworks = true
						md.ContainersToRestart = containersToRestart
						md.app.Stop()
					} else if buttonIndex == 0 {
						md.setPage(md.settingsHome.homePage)
						md.app.SetRoot(md.mainGrid, true)
					}
				})
			md.app.SetRoot(modal, true)
		} else {
			md.ShouldSave = true
			md.ContainersToRestart = containersToRestart
			md.app.Stop()
		}
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

	// Row spacers with the correct background color
	spacer1 := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacer2 := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacer3 := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacer4 := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacerL := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacerR := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	// The main content grid
	contentGrid := tview.NewGrid().
		SetRows(1, textViewHeight, 1, 0, 1, 1, 1).
		SetColumns(1, 0, 1).
		AddItem(spacer1, 0, 1, 1, 1, 0, 0, false).
		AddItem(textView, 1, 1, 1, 1, 0, 0, false).
		AddItem(spacer2, 2, 1, 1, 1, 0, 0, false).
		AddItem(changeBox, 3, 1, 1, 1, 0, 0, false).
		AddItem(spacer3, 4, 1, 1, 1, 0, 0, false).
		AddItem(buttonGrid, 5, 1, 1, 1, 0, 0, true).
		AddItem(spacer4, 6, 1, 1, 1, 0, 0, false).
		AddItem(spacerL, 0, 0, 7, 1, 0, 0, false).
		AddItem(spacerR, 0, 2, 7, 1, 0, 0, false)
	contentGrid.
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorder(true).
		SetTitle(" Review Changes ")
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

	page := newPage(nil, reviewPageID, "Review Settings", "", borderGrid)

	return &ReviewPage{
		md:              md,
		changedSettings: changedSettings,
		page:            page,
	}

}

// Get all of the changed settings between an old and new config
func getChangedSettingsMap(oldConfig *config.RocketPoolConfig, newConfig *config.RocketPoolConfig) map[string][]SettingsPair {
	changedSettings := map[string][]SettingsPair{}

	// Root settings
	oldRootParams := oldConfig.GetParameters()
	newRootParams := newConfig.GetParameters()
	changedSettings[oldConfig.Title] = getChangedSettings(oldRootParams, newRootParams, newConfig)

	// Subconfig settings
	oldSubconfigs := oldConfig.GetSubconfigs()
	for name, subConfig := range newConfig.GetSubconfigs() {
		oldParams := oldSubconfigs[name].GetParameters()
		newParams := subConfig.GetParameters()
		changedSettings[subConfig.GetConfigTitle()] = getChangedSettings(oldParams, newParams, newConfig)
	}

	return changedSettings
}

// Get all of the settings that have changed between the given parameter lists.
// Assumes the parameter lists represent identical parameters (e.g. they have the same number of elements and
// each element has the same ID).
func getChangedSettings(oldParams []*config.Parameter, newParams []*config.Parameter, newConfig *config.RocketPoolConfig) []SettingsPair {
	changedSettings := []SettingsPair{}

	for i, param := range newParams {
		oldValString := fmt.Sprint(oldParams[i].Value)
		newValString := fmt.Sprint(param.Value)
		if oldValString != newValString {
			changedSettings = append(changedSettings, SettingsPair{
				Name:               param.Name,
				OldValue:           oldValString,
				NewValue:           newValString,
				AffectedContainers: getAffectedContainers(param, newConfig),
			})
		}
	}

	return changedSettings
}

// Handles custom container overrides
func getAffectedContainers(param *config.Parameter, cfg *config.RocketPoolConfig) map[config.ContainerID]bool {

	affectedContainers := map[config.ContainerID]bool{}

	for _, container := range param.AffectsContainers {
		affectedContainers[container] = true
	}

	// Nimbus doesn't operate in split mode, so all of the VC parameters need to get redirected to the BN instead
	if cfg.ConsensusClientMode.Value.(config.Mode) == config.Mode_Local &&
		cfg.ConsensusClient.Value.(config.ConsensusClient) == config.ConsensusClient_Nimbus {
		for _, container := range param.AffectsContainers {
			if container == config.ContainerID_Validator {
				affectedContainers[config.ContainerID_Eth2] = true
			}
		}
	}
	return affectedContainers

}
