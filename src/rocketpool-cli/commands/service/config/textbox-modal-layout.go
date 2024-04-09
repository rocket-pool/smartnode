package config

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// A layout container that mimics a modal display with a series of textboxes and done/back buttons
type textBoxModalLayout struct {
	// The parent application that owns this modal (for focus changes on vertical layouts)
	app *tview.Application

	width int

	borderGrid  *tview.Grid
	contentGrid *tview.Grid
	controlGrid *tview.Grid
	done        func(text map[string]string)
	back        func()
	form        *Form

	firstTextbox *tview.InputField
	textboxes    map[string]*tview.InputField

	page *page
}

// Creates a new TextBoxModalLayout instance
func newTextBoxModalLayout(app *tview.Application, title string, width int, text string, labels []string, maxLengths []int, regexes []string) *textBoxModalLayout {

	layout := &textBoxModalLayout{
		app:       app,
		width:     width,
		textboxes: map[string]*tview.InputField{},
	}

	// Create the button grid
	height := layout.createControlGrid(labels, maxLengths, regexes)

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
		SetColumns(0, width, 0)
	borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	// Get the total content height, including spacers and borders
	lines := tview.WordWrap(text, width-4)
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
func (layout *textBoxModalLayout) createControlGrid(labels []string, maxLengths []int, regexes []string) int {

	controlGrid := tview.NewGrid().
		SetRows(0).
		SetColumns(-1, -5, -1)
	controlGrid.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

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

	// Create the controls and add listeners
	for i := 0; i < len(labels); i++ {
		textbox := tview.NewInputField().
			SetLabel(labels[i])
		maxLength := maxLengths[i]
		textbox.SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
			if maxLength > 0 {
				if len(textToCheck) > maxLength {
					return false
				}
			}
			// TODO: regex support
			return true
		})
		form.AddFormItem(textbox)
		layout.textboxes[labels[i]] = textbox

		if layout.firstTextbox == nil {
			layout.firstTextbox = textbox
		}
	}
	form.AddButton("Next", func() {
		if layout.done != nil {
			text := map[string]string{}
			for label, textbox := range layout.textboxes {
				text[label] = strings.TrimSpace(textbox.GetText())
			}
			layout.done(text)
		}
	}).
		SetButtonTextColor(tcell.ColorLightGray).
		SetButtonBackgroundActivatedColor(tcell.Color46).
		SetButtonTextActivatedColor(tcell.ColorBlack)

	// Create the columns, including the left and right spacers
	leftSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	rightSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	controlGrid.
		AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false).
		AddItem(form, 0, 1, 1, 1, 0, 0, true).
		AddItem(rightSpacer, 0, 2, 1, 1, 0, 0, false)
	layout.controlGrid = controlGrid

	return len(labels)*2 + 1
}

// Focuses the textbox
func (layout *textBoxModalLayout) focus() {
	layout.app.SetFocus(layout.firstTextbox)
	layout.form.SetFocus(0)
}
