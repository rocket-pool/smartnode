package config

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// A layout container that mimics a modal display with a series of buttons and a description box
type choiceModalLayout struct {
	// The parent application that owns this modal (for focus changes on vertical layouts)
	app *tview.Application

	width int

	borderGrid  *tview.Grid
	contentGrid *tview.Grid
	buttonGrid  *tview.Grid
	done        func(buttonIndex int, buttonLabel string)
	back        func()

	// The forms embedded in the modal's frame for the buttons.
	forms []*Form

	// The currently selected form (for vertical layouts)
	selected int

	descriptionBox *tview.TextView

	buttonDescriptions []string

	direction int

	page *page
}

// Creates a new ChoiceModalLayout instance
func newChoiceModalLayout(app *tview.Application, title string, width int, text string, buttonLabels []string, buttonDescriptions []string, direction int) *choiceModalLayout {

	layout := &choiceModalLayout{
		app:                app,
		width:              width,
		buttonDescriptions: buttonDescriptions,
		direction:          direction,
	}

	// Create the button grid
	buttonGridHeight := layout.createButtonGrid(buttonLabels, buttonDescriptions)

	// Create the main text view
	textView := tview.NewTextView().
		SetText(text).
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true).
		SetDynamicColors(true).
		SetTextColor(tview.Styles.PrimaryTextColor)
	textView.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	textView.SetBorderPadding(0, 0, 1, 1)

	// Row spacers with the correct background color
	spacer1 := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacer2 := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	spacer3 := tview.NewBox().
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	// The main content grid
	contentGrid := tview.NewGrid().
		SetRows(1, 0, 1, buttonGridHeight, 1).
		AddItem(spacer1, 0, 0, 1, 1, 0, 0, false).
		AddItem(textView, 1, 0, 1, 1, 0, 0, false).
		AddItem(spacer2, 2, 0, 1, 1, 0, 0, false).
		AddItem(layout.buttonGrid, 3, 0, 1, 1, 0, 0, true).
		AddItem(spacer3, 4, 0, 1, 1, 0, 0, false)
	contentGrid.
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorder(true).
		SetTitle(" " + title + " ")
	layout.buttonGrid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if layout.back != nil {
				layout.back()
				return nil
			}
			return event
		default:
			return event
		}
	})

	// A grid with variable spaced borders that surrounds the fixed-size content grid
	borderGrid := tview.NewGrid().
		SetColumns(0, width, 0)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	// Get the total content height, including spacers and borders
	lines := tview.WordWrap(text, width-4)
	textViewHeight := len(lines) + 4
	borderGrid.SetRows(0, textViewHeight+buttonGridHeight+2, 0, 2)

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

	// Set the content and border for the layout
	layout.contentGrid = contentGrid
	layout.borderGrid = borderGrid
	return layout

}

// Creates the grid for the layout's buttons and optional description text.
func (layout *choiceModalLayout) createButtonGrid(buttonLabels []string, buttonDescriptions []string) int {

	buttonGrid := tview.NewGrid().
		SetRows(0)
	buttonGrid.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	// This tracks the length of the buttons themselves
	buttonsWidth := 0
	height := 0

	// Self-explanatory horizontal buttons without a description box
	if layout.direction == DirectionalModalHorizontal {

		// Create the form for the buttons
		form := NewForm().
			SetButtonsAlign(tview.AlignCenter).
			SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
			SetButtonTextColor(tcell.ColorLightGray).
			SetButtonBackgroundActivatedColor(tcell.Color46).
			SetButtonTextActivatedColor(tcell.ColorBlack)
		form.
			SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
			SetBorderPadding(0, 0, 0, 0)

		// Create the buttons and add listeners
		for index, label := range buttonLabels {
			func(i int, l string) {
				form.AddButton(label, func() {
					if layout.done != nil {
						layout.done(i, l)
					}
				})
				button := form.GetButton(form.GetButtonCount() - 1)
				button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					switch event.Key() {
					case tcell.KeyDown, tcell.KeyRight:
						return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
					case tcell.KeyUp, tcell.KeyLeft:
						return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
					}
					return event
				})

				// Add the width of this button (including borders)
				buttonsWidth += tview.TaggedStringWidth(label) + 4 + 2
			}(index, label)
		}

		// Create the columns, including the left and right spacers
		buttonsWidth -= 2
		buttonGrid.SetColumns(0, buttonsWidth, 0)
		leftSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		rightSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		buttonGrid.
			AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false).
			AddItem(form, 0, 1, 1, 1, 0, 0, true).
			AddItem(rightSpacer, 0, 2, 1, 1, 0, 0, false)
		layout.forms = append(layout.forms, form)
		height = 1

		// Vertical buttons that may come with descriptions
	} else if layout.direction == DirectionalModalVertical {

		formsFlex := tview.NewFlex().
			SetDirection(tview.FlexRow)
		if len(buttonDescriptions) > 0 {
			// Add a spacing row to make the first button line up with the description box
			spacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
			formsFlex.AddItem(spacer, 1, 1, false)
		}

		// Adjust the labels so they all have the same length
		sizedButtonLabels := layout.getSizedButtonLabels(buttonLabels)
		buttonsWidth := tview.TaggedStringWidth(sizedButtonLabels[0]) + 4 + 2

		for index, label := range sizedButtonLabels {
			func(i int, l string) {

				// Create a new form for this button
				form := NewForm().
					SetButtonsAlign(tview.AlignCenter).
					SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
					SetButtonTextColor(tcell.ColorLightGray).
					SetButtonBackgroundActivatedColor(tcell.Color46).
					SetButtonTextActivatedColor(tcell.ColorBlack)
				form.SetBackgroundColor(tview.Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
				form.AddButton(label, func() {
					if layout.done != nil {
						layout.done(i, l)
					}
				})

				// Set the listeners for the button
				button := form.GetButton(form.GetButtonCount() - 1)
				button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					switch event.Key() {
					case tcell.KeyDown, tcell.KeyRight, tcell.KeyTab:
						var nextSelection int
						if layout.selected == len(layout.forms)-1 {
							nextSelection = 0
						} else {
							nextSelection = i + 1
						}
						if layout.descriptionBox != nil {
							layout.descriptionBox.SetText(buttonDescriptions[nextSelection])
						}
						layout.app.SetFocus(layout.forms[nextSelection])
						layout.selected = nextSelection
						return tcell.NewEventKey(tcell.KeyDown, 0, 0)
					case tcell.KeyUp, tcell.KeyLeft, tcell.KeyBacktab:
						var nextSelection int
						if layout.selected == 0 {
							nextSelection = len(layout.forms) - 1
						} else {
							nextSelection = i - 1
						}
						if layout.descriptionBox != nil {
							layout.descriptionBox.SetText(buttonDescriptions[nextSelection])
						}
						layout.app.SetFocus(layout.forms[nextSelection])
						layout.selected = nextSelection
						return tcell.NewEventKey(tcell.KeyUp, 0, 0)
					default:
						return event
					}
				})

				// Add the form to the layout's list of forms
				layout.forms = append(layout.forms, form)
				formsFlex.AddItem(form, 1, 1, true)
				spacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
				formsFlex.AddItem(spacer, 1, 1, false)
			}(index, label)
		}

		bottomFormSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
		formsFlex.AddItem(bottomFormSpacer, 0, 2, false)

		// Create the columns, including the left and right spacers
		buttonsWidth -= 2
		if len(buttonDescriptions) == 0 {
			buttonGrid.SetColumns(0, buttonsWidth, 0)

			leftSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
			rightSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

			buttonGrid.AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false)
			buttonGrid.AddItem(formsFlex, 0, 1, 1, 1, 0, 0, true)
			buttonGrid.AddItem(rightSpacer, 0, 2, 1, 1, 0, 0, false)

			height = len(layout.forms)*2 - 1
		} else {
			// If this layout comes with button descriptions, include the description box
			layout.descriptionBox = tview.NewTextView().
				SetWordWrap(true).
				SetDynamicColors(true)
			layout.descriptionBox.SetBorder(true)
			layout.descriptionBox.SetBorderPadding(0, 0, 1, 1)
			layout.descriptionBox.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
			layout.descriptionBox.SetTitle("Description")

			leftSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
			midSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
			rightSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

			buttonGrid.SetColumns(1, -2, 1, -5, 1)
			buttonGrid.AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false)
			buttonGrid.AddItem(formsFlex, 0, 1, 1, 1, 0, 0, true)
			buttonGrid.AddItem(midSpacer, 0, 2, 1, 1, 0, 0, false)
			buttonGrid.AddItem(layout.descriptionBox, 0, 3, 1, 1, 0, 0, false)
			buttonGrid.AddItem(rightSpacer, 0, 4, 1, 1, 0, 0, false)

			height = len(layout.forms)*2 + 1

			// Get the description box height
			for _, description := range buttonDescriptions {
				lines := tview.WordWrap(description, (layout.width-6)/2)
				if len(lines) > height {
					height = len(lines)
				}
			}
		}

	}

	layout.buttonGrid = buttonGrid

	return height
}

// Pads each of the button labels with spaces so they all have the same length while staying centered.
func (layout *choiceModalLayout) getSizedButtonLabels(buttonLabels []string) []string {

	// Get the longest label
	maxLabelSize := 0
	for _, label := range buttonLabels {
		if len(label) > maxLabelSize {
			maxLabelSize = tview.TaggedStringWidth(label)
		}
	}

	// Pad each label
	sizedButtonLabels := []string{}
	for _, label := range buttonLabels {
		length := tview.TaggedStringWidth(label)
		leftPad := (maxLabelSize - length) / 2
		rightPad := maxLabelSize - length - leftPad

		sizedLabel := strings.Repeat(" ", leftPad) + label + strings.Repeat(" ", rightPad)
		sizedButtonLabels = append(sizedButtonLabels, sizedLabel)
	}

	return sizedButtonLabels

}

// Focuses the given button
func (layout *choiceModalLayout) focus(index int) {

	if layout.direction == DirectionalModalVertical {
		if layout.descriptionBox != nil {
			layout.descriptionBox.SetText(layout.buttonDescriptions[index])
		}
		if index < 0 || index > len(layout.forms)-1 {
			return
		}
		layout.app.SetFocus(layout.forms[index])
		layout.selected = index
	} else {
		if len(layout.forms) > 0 {
			if index < 0 || index > layout.forms[0].GetButtonCount()-1 {
				return
			}
			layout.forms[0].SetFocus(index)
			layout.app.SetFocus(layout.forms[0])
			layout.selected = index
		}
	}
}
