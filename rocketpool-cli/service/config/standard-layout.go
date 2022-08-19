package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// A layout container with the standard elements and design
type standardLayout struct {
	grid           *tview.Grid
	content        tview.Primitive
	descriptionBox *tview.TextView
	footer         tview.Primitive
	form           *Form
	parameters     map[tview.FormItem]*parameterizedFormItem
	cfg            cfgtypes.Config
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
	descriptionBox.SetDynamicColors(true)

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

// Create a standard form for this layout (for settings pages)
func (layout *standardLayout) createForm(networkParam *cfgtypes.Parameter, title string) {

	layout.parameters = map[tview.FormItem]*parameterizedFormItem{}

	// Create the form
	form := NewForm().
		SetFieldBackgroundColor(tcell.ColorBlack)
	form.
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorderPadding(0, 0, 0, 0)

	// Set up the selected parameter change callback to update the description box
	form.SetChangedFunc(func(index int) {
		if index < form.GetFormItemCount() {
			formItem := form.GetFormItem(index)
			param := layout.parameters[formItem].parameter
			defaultValue, _ := param.GetDefault(networkParam.Value.(cfgtypes.Network))
			descriptionText := fmt.Sprintf("Default: %v\n\n%s", defaultValue, param.Description)
			layout.descriptionBox.SetText(descriptionText)
			layout.descriptionBox.ScrollToBeginning()
		}
	})

	layout.form = form
	layout.setContent(form, form.Box, title)
	layout.createSettingFooter()
}

// Refreshes all of the form items to show the current configured values
func (layout *standardLayout) refresh() {

	for i := 0; i < layout.form.GetFormItemCount(); i++ {
		formItem := layout.form.GetFormItem(i)
		param := layout.parameters[formItem].parameter

		// Set the form item to the current value
		switch param.Type {
		case cfgtypes.ParameterType_Bool:
			formItem.(*tview.Checkbox).SetChecked(param.Value == true)

		case cfgtypes.ParameterType_Int, cfgtypes.ParameterType_Uint, cfgtypes.ParameterType_Uint16, cfgtypes.ParameterType_String, cfgtypes.ParameterType_Float:
			formItem.(*tview.InputField).SetText(fmt.Sprint(param.Value))

		case cfgtypes.ParameterType_Choice:
			for i := 0; i < len(param.Options); i++ {
				if param.Options[i].Value == param.Value {
					formItem.(*DropDown).SetCurrentOption(i)
				}
			}
		}
	}

	// Focus the first element
	layout.form.SetFocus(0)

}

// Create the footer, including the nav bar
func (layout *standardLayout) createSettingFooter() {

	// Nav bar
	navString1 := "Arrow keys: Navigate   Space/Enter: Change Setting"
	navTextView1 := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	fmt.Fprint(navTextView1, navString1)

	navString2 := "Esc: Go Back to Categories"
	navTextView2 := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	fmt.Fprint(navTextView2, navString2)

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

	layout.setFooter(navBar, 2)

}

// Add a collection of form items to this layout's form
func (layout *standardLayout) addFormItems(params []*parameterizedFormItem) {
	for _, param := range params {
		layout.form.AddFormItem(param.item)
	}
}

// Add a collection of "common" and "specific" form items to this layout's form, where some of the common
// items may not be valid and should be excluded
func (layout *standardLayout) addFormItemsWithCommonParams(commonParams []*parameterizedFormItem, specificParams []*parameterizedFormItem, unsupportedCommonParams []string) {

	// Add the common params if they aren't in the unsupported list
	for _, commonParam := range commonParams {
		isSupported := true
		for _, unsupportedParam := range unsupportedCommonParams {
			if commonParam.parameter.ID == unsupportedParam {
				isSupported = false
				break
			}
		}

		if isSupported {
			layout.form.AddFormItem(commonParam.item)
		}
	}

	// Add all of the specific params
	for _, specificParam := range specificParams {
		layout.form.AddFormItem(specificParam.item)
	}

}

func (layout *standardLayout) mapParameterizedFormItems(params ...*parameterizedFormItem) {
	for _, param := range params {
		layout.parameters[param.item] = param
	}
}
