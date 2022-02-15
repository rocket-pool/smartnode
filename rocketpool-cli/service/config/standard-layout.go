package config

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// A layout container with the standard elements and design
type standardLayout struct {
	grid           *tview.Grid
	content        tview.Primitive
	descriptionBox *tview.TextView
	footer         tview.Primitive
	form           *Form
	cfg            config.Config
	defaults       []string
	formItems      map[string]tview.FormItem
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

func (layout *standardLayout) createFormForConfig(cfg config.Config, network config.Network, title string) {

	layout.cfg = cfg
	layout.formItems = map[string]tview.FormItem{}

	// Create the form
	form := NewForm().
		SetFieldBackgroundColor(tcell.ColorBlack)
	form.
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorderPadding(0, 0, 0, 0)

	// Set up the selected parameter change callback to update the description box
	params := layout.cfg.GetParameters()
	paramDescriptions := []string{}
	for _, param := range params {
		paramDescriptions = append(paramDescriptions, param.Description)
		layout.descriptionBox.ScrollToBeginning()
	}
	form.SetChangedFunc(func(index int) {
		descriptionText := fmt.Sprintf("Default: %s\n\n%s", layout.defaults[index], paramDescriptions[index])
		layout.descriptionBox.SetText(descriptionText)
		layout.descriptionBox.ScrollToBeginning()
	})

	// Set up the form items
	for _, param := range params {
		var item tview.FormItem
		switch param.Type {
		case config.ParameterType_Bool:
			item = layout.createCheckbox(param, network)
		case config.ParameterType_Int:
			item = layout.createIntField(param, network)
		case config.ParameterType_Uint:
			item = layout.createUintField(param, network)
		case config.ParameterType_Uint16:
			item = layout.createUint16Field(param, network)
		case config.ParameterType_String:
			item = layout.createStringField(param, network)
		case config.ParameterType_Choice:
			item = layout.createChoiceDropDown(param, network)
		}

		form.AddFormItem(item)
		layout.formItems[param.ID] = item

	}

	layout.form = form
	layout.setContent(form, form.Box, title)
	layout.createSettingFooter()
	layout.refresh(network)
}

// Create a standard form checkbox
func (layout *standardLayout) createCheckbox(param *config.Parameter, network config.Network) *tview.Checkbox {
	item := tview.NewCheckbox().
		SetLabel(param.Name).
		SetChecked(param.Value == true).
		SetChangedFunc(func(checked bool) {
			param.Value = checked
		})
	return item
}

// Create a standard int field
func (layout *standardLayout) createIntField(param *config.Parameter, network config.Network) *tview.InputField {
	item := tview.NewInputField().
		SetLabel(param.Name).
		SetAcceptanceFunc(tview.InputFieldInteger)
	item.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			item.SetText("")
		} else {
			value, err := strconv.ParseInt(item.GetText(), 0, 0)
			if err != nil {
				// TODO: show error modal?
				item.SetText("")
			} else {
				param.Value = int(value)
			}
		}
	})
	return item
}

// Create a standard uint field
func (layout *standardLayout) createUintField(param *config.Parameter, network config.Network) *tview.InputField {
	item := tview.NewInputField().
		SetLabel(param.Name).
		SetAcceptanceFunc(tview.InputFieldInteger)
	item.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			item.SetText("")
		} else {
			value, err := strconv.ParseUint(item.GetText(), 0, 0)
			if err != nil {
				// TODO: show error modal?
				item.SetText("")
			} else {
				param.Value = uint(value)
			}
		}
	})
	return item
}

// Create a standard uint16 field
func (layout *standardLayout) createUint16Field(param *config.Parameter, network config.Network) *tview.InputField {
	item := tview.NewInputField().
		SetLabel(param.Name).
		SetAcceptanceFunc(tview.InputFieldInteger)
	item.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			item.SetText("")
		} else {
			value, err := strconv.ParseUint(item.GetText(), 0, 16)
			if err != nil {
				// TODO: show error modal?
				item.SetText("")
			} else {
				param.Value = uint(value)
			}
		}
	})
	return item
}

// Create a standard string field
func (layout *standardLayout) createStringField(param *config.Parameter, network config.Network) *tview.InputField {
	item := tview.NewInputField().
		SetLabel(param.Name).
		SetAcceptanceFunc(tview.InputFieldInteger)
	item.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			item.SetText("")
		} else {
			param.Value = item.GetText()
		}
	})
	return item
}

// Create a standard choice field
func (layout *standardLayout) createChoiceDropDown(param *config.Parameter, network config.Network) *DropDown {
	// Create the list of options
	options := []string{}
	descriptions := []string{}
	values := []interface{}{}
	for _, option := range param.Options {
		options = append(options, option.Name)
		descriptions = append(descriptions, option.Description)
		values = append(values, option.Value)
	}
	item := NewDropDown().
		SetLabel(param.Name).
		SetOptions(options, func(text string, index int) {
			param.Value = values[index]
		}).
		SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			layout.descriptionBox.SetText(descriptions[index])
		})
	item.SetTextOptions(" ", " ", "", "", "")
	return item
}

// Refreshes all of the form items to show the current configured values
func (layout *standardLayout) refresh(network config.Network) {

	layout.defaults = []string{}
	params := layout.cfg.GetParameters()
	for _, param := range params {
		// Recreate the default text for this parameter
		defaultValue, _ := param.GetDefault(network)
		defaultValueString := fmt.Sprint(defaultValue)

		// Set the form item to the current value
		formItem := layout.formItems[param.ID]
		switch param.Type {
		case config.ParameterType_Bool:
			formItem.(*tview.Checkbox).SetChecked(param.Value == true)
			layout.defaults = append(layout.defaults, defaultValueString)

		case config.ParameterType_Int, config.ParameterType_Uint, config.ParameterType_Uint16, config.ParameterType_String:
			formItem.(*tview.InputField).SetText(fmt.Sprint(param.Value))
			layout.defaults = append(layout.defaults, defaultValueString)

		case config.ParameterType_Choice:
			for i := 0; i < len(param.Options); i++ {
				if param.Options[i].Value == param.Value {
					formItem.(*DropDown).SetCurrentOption(i)
				}
				if param.Options[i].Value == defaultValue {
					layout.defaults = append(layout.defaults, param.Options[i].Name)
				}
			}
		}
	}

	// Focus the first element
	layout.form.SetFocus(0)

}

// Create the footer, including the nav bar and the save / quit buttons
func (layout *standardLayout) createSettingFooter() {

	// Nav bar
	navString1 := "Tab: Next Setting   Shift-Tab: Previous Setting"
	navTextView1 := tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(false).
		SetWrap(false)
	fmt.Fprint(navTextView1, navString1)

	navString2 := "Space/Enter: Change Setting   Esc: Done, Return to Categories"
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
