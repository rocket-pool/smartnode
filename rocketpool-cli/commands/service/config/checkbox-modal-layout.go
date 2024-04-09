package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// A layout container that mimics a modal display with a series of checkboxes with descriptions and done/back buttons
type checkBoxModalLayout struct {
	// The parent application that owns this modal (for focus changes on vertical layouts)
	app *tview.Application

	width int

	borderGrid  *tview.Grid
	contentGrid *tview.Grid
	controlGrid *tview.Grid
	done        func(text map[string]bool)
	back        func()
	form        *Form
	buttonForm  *Form
	formFlex    *tview.Flex

	firstCheckbox *tview.Checkbox
	checkboxes    map[string]*tview.Checkbox

	selected int

	descriptionBox       *tview.TextView
	checkboxDescriptions []string

	page *page
}

// Creates a new CheckBoxModalLayout instance
func newCheckBoxModalLayout(app *tview.Application, title string, width int, text string) *checkBoxModalLayout {

	layout := &checkBoxModalLayout{
		app:   app,
		width: width,
	}

	layout.checkboxes = map[string]*tview.Checkbox{}

	// Create the button grid
	height := layout.createControlGrid()

	// Create the main text view
	textView := tview.NewTextView().
		SetText(text).
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true).
		SetTextColor(tview.Styles.PrimaryTextColor).
		SetDynamicColors(true)
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
		SetRows(1, 0, 1, height, 1).
		AddItem(spacer1, 0, 0, 1, 1, 0, 0, false).
		AddItem(textView, 1, 0, 1, 1, 0, 0, false).
		AddItem(spacer2, 2, 0, 1, 1, 0, 0, false).
		AddItem(layout.controlGrid, 3, 0, 1, 1, 0, 0, true).
		AddItem(spacer3, 4, 0, 1, 1, 0, 0, false)
	contentGrid.
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorder(true).
		SetTitle(" " + title + " ")
	layout.controlGrid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
		SetColumns(0, layout.width, 0)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	// Get the total content height, including spacers and borders
	lines := tview.WordWrap(text, layout.width-4)
	textViewHeight := len(lines) + 4
	borderGrid.SetRows(0, textViewHeight+height+3, 0, 2)

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

// Creates the grid for the layout's controls
func (layout *checkBoxModalLayout) createControlGrid() int {

	controlGrid := tview.NewGrid().
		SetRows(0, 1, 1, 1)
	controlGrid.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	formFlex := tview.NewFlex().
		SetDirection(tview.FlexRow)
	// Add a spacing row to make the first button line up with the description box
	spacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	formFlex.AddItem(spacer, 1, 1, false)
	layout.formFlex = formFlex

	// Create the form for the controls
	form := NewForm().
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetFieldBackgroundColor(tcell.ColorBlack)
	form.
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorderPadding(0, 0, 0, 0)
	layout.form = form

	formFlex.AddItem(layout.form, 0, 2, true)
	bottomFormSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	formFlex.AddItem(bottomFormSpacer, 0, 2, false)

	// Create the form for the Next button
	nextButtonForm := NewForm().
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetFieldBackgroundColor(tcell.ColorBlack)
	nextButtonForm.
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorderPadding(0, 0, 0, 0)
	nextButtonForm.AddButton("Next", func() {
		if layout.done != nil {
			settings := map[string]bool{}
			for label, checkBox := range layout.checkboxes {
				settings[label] = checkBox.IsChecked()
			}
			layout.done(settings)
		}
	}).
		SetButtonTextColor(tcell.ColorLightGray).
		SetButtonBackgroundActivatedColor(tcell.Color46).
		SetButtonTextActivatedColor(tcell.ColorBlack)

	// Set the listeners for the button
	button := nextButtonForm.GetButton(0)
	button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var nextSelection int
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyTab:
			nextSelection = 0
			if layout.descriptionBox != nil {
				layout.descriptionBox.SetText(layout.checkboxDescriptions[nextSelection])
			}
			layout.form.SetFocus(0)
			layout.app.SetFocus(layout.form)
			layout.selected = nextSelection
			return nil
		case tcell.KeyUp, tcell.KeyBacktab:
			nextSelection = len(layout.checkboxDescriptions) - 1
			if layout.descriptionBox != nil {
				layout.descriptionBox.SetText(layout.checkboxDescriptions[nextSelection])
			}
			layout.form.SetFocus(len(layout.checkboxDescriptions) - 1)
			layout.app.SetFocus(layout.form)
			layout.selected = nextSelection
			return nil
		default:
			return event
		}
	})
	layout.buttonForm = nextButtonForm

	// Create the columns, including the left and right spacers
	leftSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	rightSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	controlGrid.
		AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false).
		AddItem(formFlex, 0, 1, 1, 1, 0, 0, true).
		AddItem(rightSpacer, 0, 2, 1, 1, 0, 0, false)
	layout.controlGrid = controlGrid

	// Create the columns, including the left and right spacers
	layout.descriptionBox = tview.NewTextView().
		SetWordWrap(true).
		SetDynamicColors(true)
	layout.descriptionBox.SetBorder(true)
	layout.descriptionBox.SetBorderPadding(0, 0, 1, 1)
	layout.descriptionBox.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	layout.descriptionBox.SetTitle("Description")

	leftSpacer = tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	midSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	rightSpacer = tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

	controlGrid.SetRows(0, 1, 1)
	controlGrid.SetColumns(1, -3, 1, -5, 1)
	controlGrid.AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false)
	controlGrid.AddItem(formFlex, 0, 1, 1, 1, 0, 0, true)
	controlGrid.AddItem(midSpacer, 0, 2, 1, 1, 0, 0, false)
	controlGrid.AddItem(layout.descriptionBox, 0, 3, 1, 1, 0, 0, false)
	controlGrid.AddItem(rightSpacer, 0, 4, 1, 1, 0, 0, false)

	// Add spacers and the Next button
	topSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	controlGrid.AddItem(topSpacer, 1, 0, 1, 5, 0, 0, false)
	controlGrid.AddItem(layout.buttonForm, 2, 0, 1, 5, 0, 0, false)

	layout.controlGrid = controlGrid
	return 12 // TODO: dynamic sizing, may require rebuilding the entire modal

}

func (layout *checkBoxModalLayout) generateCheckboxes(labels []string, descriptions []string, settings []bool) {

	layout.form.Clear(true)
	layout.checkboxDescriptions = descriptions
	layout.descriptionBox.SetText(descriptions[0])

	// Create the controls
	for i := 0; i < len(labels); i++ {
		checkBox := tview.NewCheckbox().
			SetLabel(labels[i]).
			SetChecked(settings[i])
		layout.form.AddFormItem(checkBox)
		layout.checkboxes[labels[i]] = checkBox

		if i == 0 {
			layout.firstCheckbox = checkBox
		}
		i := i
		descriptions := descriptions
		checkBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyDown, tcell.KeyTab:
				var nextSelection int
				if layout.selected == len(descriptions)-1 {
					nextSelection = -1
					layout.selected = nextSelection
					layout.app.SetFocus(layout.buttonForm)
					return nil
				} else {
					nextSelection = i + 1
					layout.selected = nextSelection
					layout.descriptionBox.SetText(descriptions[nextSelection])
					return tcell.NewEventKey(tcell.KeyTab, 0, 0)
				}
			case tcell.KeyUp, tcell.KeyBacktab:
				var nextSelection int
				if layout.selected == 0 {
					nextSelection = -1
					layout.selected = nextSelection
					layout.app.SetFocus(layout.buttonForm)
					return nil
				} else {
					nextSelection = i - 1
					layout.selected = nextSelection
					layout.descriptionBox.SetText(descriptions[nextSelection])
					return tcell.NewEventKey(tcell.KeyBacktab, 0, 0)
				}
			default:
				return event
			}
		})
	}

	layout.formFlex.Clear()
	spacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	layout.formFlex.AddItem(spacer, 1, 1, false)
	layout.formFlex.AddItem(layout.form, len(labels)*2, 1, true)
	bottomFormSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	layout.formFlex.AddItem(bottomFormSpacer, 0, 1, false)

	/*
		// Calculate the total height
		height := len(labels)*2 + 4

		// Get the description box height
		for _, description := range descriptions {
			lines := tview.WordWrap(description, (layout.width-6)/2)
			if len(lines) > height {
				height = len(lines)
			}
		}

		layout.contentGrid.SetRows(1, 0, 1, height, 1)
	*/

}

// Focuses the modal
func (layout *checkBoxModalLayout) focus() {
	layout.form.SetFocus(0)
	layout.app.SetFocus(layout.firstCheckbox)
	layout.selected = 0
}
