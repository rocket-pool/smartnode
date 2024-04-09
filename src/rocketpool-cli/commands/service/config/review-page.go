package config

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/v2/shared"
	snCfg "github.com/rocket-pool/smartnode/v2/shared/config"
)

// Constants
const reviewPageID string = "review-settings"

// The changed settings review page
type ReviewPage struct {
	md              *mainDisplay
	changedSettings []*config.ChangedSection
	page            *page
}

// Create a page to review any changes
func NewReviewPage(md *mainDisplay, oldConfig *snCfg.SmartNodeConfig, newConfig *snCfg.SmartNodeConfig) *ReviewPage {
	var changedSettings []*config.ChangedSection
	var totalAffectedContainers map[config.ContainerID]bool
	var changeNetworks bool
	var containersToRestart []config.ContainerID

	// Create the visual list for all of the changed settings
	changeBox := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)
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
		changedSettings, totalAffectedContainers, changeNetworks = newConfig.GetChanges(oldConfig)

		// Add changed containers if this is an update
		if md.isUpdate {
			totalAffectedContainers[config.ContainerID_Daemon] = true
			builder.WriteString(fmt.Sprintf("Updated to Smart Node v%s (will affect several containers)\n\n", shared.RocketPoolVersion))
		}

		// Get the map of changed settings by section name
		if len(changedSettings) > 0 {
			for _, change := range changedSettings {
				addChangesToDescription(change, "", &builder)
			}
		}

		// Print the list of containers to restart
		if builder.String() == "" {
			builder.WriteString("<No changes>")
		} else {
			builder.WriteString("The following containers must be restarted for these changes to take effect:")
			for container := range totalAffectedContainers {
				suffix := snCfg.GetContainerName(container)
				containerName := oldConfig.GetDockerArtifactName(suffix)
				builder.WriteString(fmt.Sprintf("\n\t%s", containerName))
				containersToRestart = append(containersToRestart, container)
			}
		}
	}

	changeBox.SetText(builder.String())

	// Create the layout
	width := 86

	// Create the main text view
	descriptionText := "Please review your changes below.\nScroll through them using the arrow keys, and press Enter when you're ready to save them."
	lines := tview.WordWrap(descriptionText, width-4)
	textViewHeight := len(lines) + 1
	textView := tview.NewTextView().
		SetText(descriptionText).
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true).
		SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	textView.SetBorderPadding(0, 0, 1, 1)

	var buttonGrid *tview.Flex

	if len(errors) > 0 {
		buttonGrid = tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(tview.NewBox().
				SetBackgroundColor(tview.Styles.ContrastBackgroundColor), 0, 1, false)
	} else {
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
			md.ContainersToRestart = containersToRestart
			if changeNetworks && !md.isNew {
				md.ChangeNetworks = true
			}
			md.app.Stop()
		})
		saveButton.SetBackgroundColorActivated(tcell.Color46)
		saveButton.SetLabelColorActivated(tcell.ColorBlack)

		buttonGrid = tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(tview.NewBox().
				SetBackgroundColor(tview.Styles.ContrastBackgroundColor), 0, 1, false).
			AddItem(saveButton, len(saveButton.GetLabel())+2, 0, true).
			AddItem(tview.NewBox().
				SetBackgroundColor(tview.Styles.ContrastBackgroundColor), 0, 1, false)
	}

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

// Add all of the changed parameters to the description builder
func addChangesToDescription(section *config.ChangedSection, titlePrefix string, description *strings.Builder) {
	// Get the full section name, including the title
	var sectionName string
	if titlePrefix == "" {
		sectionName = section.Name
	} else {
		sectionName = fmt.Sprintf("%s > %s", titlePrefix, section.Name)
	}

	// Handle the parameters
	if len(section.Settings) > 0 {
		description.WriteString(fmt.Sprintf("{%s}\n", sectionName))
		for _, setting := range section.Settings {
			description.WriteString(fmt.Sprintf("\t%s: %s => %s\n", setting.Name, setting.OldValue, setting.NewValue))
		}
		description.WriteString("\n")
	}

	// Handle the subsections
	for _, subsection := range section.Subsections {
		addChangesToDescription(subsection, sectionName, description)
	}
}
