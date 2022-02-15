package config

import (
	"fmt"

	"github.com/rivo/tview"
)

// A layout container with the standard elements and design
type standardLayout struct {
	grid           *tview.Grid
	content        tview.Primitive
	descriptionBox *tview.TextView
	footer         tview.Primitive
}

// Creates a new StandardLayout instance, which includes the grid and description box preconstructed.
func newStandardLayout() *standardLayout {

	// Create the main display grid
	grid := tview.NewGrid().
		SetColumns(-5, 2, -3).
		SetRows(0, 1, 0).
		SetBorders(false)

	// Create the description box
	descriptionBox := tview.NewTextView()
	descriptionBox.SetBorder(true)
	descriptionBox.SetBorderPadding(0, 0, 1, 1)
	descriptionBox.SetTitle(" Description ")
	descriptionBox.SetWordWrap(true)
	descriptionBox.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	//descriptionBox.SetBorderColor(tcell.ColorDeepSkyBlue)
	//descriptionBox.SetTextColor(tcell.ColorDeepSkyBlue)

	grid.AddItem(descriptionBox, 0, 2, 1, 1, 0, 0, false)

	return &standardLayout{
		grid:           grid,
		descriptionBox: descriptionBox,
	}

}

// Sets the main content (the box on the left side of the screen) for this layout,
// applying the default styles to it.
func (layout *standardLayout) setContent(content tview.Primitive, contentBox *tview.Box, title string) {

	// Set the standard properties for the content (border and title)
	contentBox.SetBorder(true)
	contentBox.SetBorderPadding(1, 1, 1, 1)
	contentBox.SetTitle(fmt.Sprintf(" %s ", title))
	//contentBox.SetBorderColor(tcell.ColorGreen)

	// Add the content to the grid
	layout.content = content
	layout.grid.AddItem(content, 0, 0, 1, 1, 0, 0, true)
}

// Sets the footer for this layout.
func (layout *standardLayout) setFooter(footer tview.Primitive, height int) {

	if footer == nil {
		layout.grid.SetRows(0, 1)
	} else {
		// Add the footer to the grid
		layout.footer = footer
		layout.grid.SetRows(0, 1, height)
		layout.grid.AddItem(footer, 2, 0, 1, 3, 0, 0, false)
	}

}
