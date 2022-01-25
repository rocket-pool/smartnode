package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// A layout container that mimics a modal display
type modalLayout struct {
    // The parent application that owns this modal (for focus changes on vertical layouts)
    app *tview.Application

    borderGrid *tview.Grid
	contentGrid *tview.Grid
    buttonGrid *tview.Grid
	done func(buttonIndex int, buttonLabel string)

	// The forms embedded in the modal's frame for the buttons.
	forms []*tview.Form

    // The currently selected form (for vertical layouts)
    selected int

    descriptionBox *tview.TextView
}


// Creates a new ModalLayout instance
func newModalLayout(app *tview.Application, width int, text string, buttonLabels []string, buttonDescriptions []string, direction int) (*modalLayout) {

    layout := &modalLayout{
        app: app,
    }

    // Create the button grid
    buttonGridHeight := layout.createButtonGrid(buttonLabels, buttonDescriptions, direction)

    // Create the main text view
    textView := tview.NewTextView().
        SetText(text).
        SetTextAlign(tview.AlignCenter).
        SetWordWrap(true).
        SetTextColor(tview.Styles.PrimaryTextColor)
    textView.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

    // Row spacers with the correct background color 
    spacer1 := tview.NewBox().
        SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
    spacer2 := tview.NewBox().
        SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

    // The main content grid
    contentGrid := tview.NewGrid().
        SetRows(0, 1, buttonGridHeight, 1).
        AddItem(textView, 0, 0, 1, 1, 0, 0, false).
        AddItem(spacer1, 1, 0, 1, 1, 0, 0, false).
        AddItem(layout.buttonGrid, 2, 0, 1, 1, 0, 0, true).
        AddItem(spacer2, 3, 0, 1, 1, 0, 0, false)
    contentGrid.
        SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
        SetBorder(true)

    // A grid with variable spaced borders that surrounds the fixed-size content grid
    borderGrid := tview.NewGrid().
        SetColumns(0, width, 0)
    borderGrid.AddItem(contentGrid, 1, 1, 1, 1, 0, 0, true)

	// Get the total content height, including spacers and borders
    lines := tview.WordWrap(text, width)
	textViewHeight := len(lines) + 2
    borderGrid.SetRows(0, textViewHeight + buttonGridHeight + 2, 0, 1)

    // Create the nav footer text view
    navString1 := "Arrow keys: Navigate   Space/Enter: Select"
    navTextView1 := tview.NewTextView().
        SetDynamicColors(false).
        SetRegions(false).
        SetWrap(false)
    fmt.Fprint(navTextView1, navString1)

    navString2 := "Esc: Quit without Saving"
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
func (layout *modalLayout) createButtonGrid(buttonLabels []string, buttonDescriptions []string, direction int) int {

    buttonGrid := tview.NewGrid().
        SetRows(0)
    buttonGrid.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)

    // This tracks the length of the buttons themselves
	buttonsWidth := 0
    height := 0

    // Self-explanatory horizontal buttons without a description box
    if direction == DirectionalModalHorizontal {

        // Create the form for the buttons
        form := tview.NewForm().
            SetButtonsAlign(tview.AlignCenter).
            SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
            SetButtonTextColor(tview.Styles.PrimaryTextColor)
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

        height = 1

    // Vertical buttons that may come with descriptions
    } else if direction == DirectionalModalVertical {

        formsFlex := tview.NewFlex().
            SetDirection(tview.FlexRow)
        for index, label := range buttonLabels {
            func(i int, l string) {

                // Create a new form for this button
                form := tview.NewForm().
                    SetButtonsAlign(tview.AlignCenter).
                    SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
                    SetButtonTextColor(tview.Styles.PrimaryTextColor)
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
                    case tcell.KeyUp, tcell.KeyLeft, tcell.KeyBacktab:
                        var nextSelection int
                        if layout.selected == 0 {
                            nextSelection = len(layout.forms)-1
                        } else {
                            nextSelection = i - 1
                        }
                        if layout.descriptionBox != nil {
                            layout.descriptionBox.SetText(buttonDescriptions[nextSelection])
                        }
                        layout.app.SetFocus(layout.forms[nextSelection])
                        layout.selected = nextSelection
                    }
                    return event
                })

                // Calculate the width of this button 
                buttonWidth := tview.TaggedStringWidth(label) + 4 + 2
                if buttonWidth > buttonsWidth {
                    buttonsWidth = buttonWidth
                }

                // Add the form to the layout's list of forms
                layout.forms = append(layout.forms, form)
                formsFlex.AddItem(form, 1, 1, true)
                spacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
                formsFlex.AddItem(spacer, 1, 1, false)
            }(index, label)
        }

        // Create the columns, including the left and right spacers
        buttonsWidth -= 2
        if len(buttonDescriptions) == 0 {
            buttonGrid.SetColumns(0, buttonsWidth, 0)
            leftSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
            rightSpacer := tview.NewBox().SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
            buttonGrid.AddItem(leftSpacer, 0, 0, 1, 1, 0, 0, false)
            buttonGrid.AddItem(formsFlex, 0, 1, 1, 1, 0, 0, true)
            buttonGrid.AddItem(rightSpacer, 0, 2, 1, 1, 0, 0, false)
        } else {
            // If this layout comes with button descriptions, include the description box
            layout.descriptionBox = tview.NewTextView()
            buttonGrid.SetColumns(1, buttonsWidth, 1, 0, 1)
            buttonGrid.AddItem(formsFlex, 0, 1, 1, 1, 0, 0, true)
            buttonGrid.AddItem(layout.descriptionBox, 0, 3, 1, 1, 0, 0, false)
        }


        height = len(layout.forms) * 2 - 1

    }

    layout.buttonGrid = buttonGrid

    return height
}

