package config

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// A form item linked to a Parameter
type parameterizedFormItem struct {
	parameter *config.Parameter
	item      tview.FormItem
}

func registerEnableCheckbox(param *config.Parameter, checkbox *tview.Checkbox, form *Form, items []*parameterizedFormItem) {
	checkbox.SetChangedFunc(func(checked bool) {
		param.Value = checked

		if !checked {
			form.Clear(true)
			form.AddFormItem(checkbox)
		} else {
			for _, item := range items {
				form.AddFormItem(item.item)
			}
		}
	})
}

// Create a list of form items based on a set of parameters
func createParameterizedFormItems(params []*config.Parameter, descriptionBox *tview.TextView) []*parameterizedFormItem {
	formItems := []*parameterizedFormItem{}
	for _, param := range params {
		var item *parameterizedFormItem
		switch param.Type {
		case config.ParameterType_Bool:
			item = createParameterizedCheckbox(param)
		case config.ParameterType_Int:
			item = createParameterizedIntField(param)
		case config.ParameterType_Uint:
			item = createParameterizedUintField(param)
		case config.ParameterType_Uint16:
			item = createParameterizedUint16Field(param)
		case config.ParameterType_String:
			item = createParameterizedStringField(param)
		case config.ParameterType_Choice:
			item = createParameterizedDropDown(param, descriptionBox)
		case config.ParameterType_Float:
			item = createParameterizedStringField(param)
		default:
			panic(fmt.Sprintf("Unknown parameter type %v", param))
		}
		formItems = append(formItems, item)
	}

	return formItems
}

// Create a standard form checkbox
func createParameterizedCheckbox(param *config.Parameter) *parameterizedFormItem {
	item := tview.NewCheckbox().
		SetLabel(param.Name).
		SetChecked(param.Value == true).
		SetChangedFunc(func(checked bool) {
			param.Value = checked
		})
	item.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyTab:
			return tcell.NewEventKey(tcell.KeyTab, 0, 0)
		case tcell.KeyUp, tcell.KeyBacktab:
			return tcell.NewEventKey(tcell.KeyBacktab, 0, 0)
		default:
			return event
		}
	})

	return &parameterizedFormItem{
		parameter: param,
		item:      item,
	}
}

// Create a standard int field
func createParameterizedIntField(param *config.Parameter) *parameterizedFormItem {
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
	item.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyTab:
			return tcell.NewEventKey(tcell.KeyTab, 0, 0)
		case tcell.KeyUp, tcell.KeyBacktab:
			return tcell.NewEventKey(tcell.KeyBacktab, 0, 0)
		default:
			return event
		}
	})

	return &parameterizedFormItem{
		parameter: param,
		item:      item,
	}
}

// Create a standard uint field
func createParameterizedUintField(param *config.Parameter) *parameterizedFormItem {
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
				param.Value = int(value)
			}
		}
	})
	item.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyTab:
			return tcell.NewEventKey(tcell.KeyTab, 0, 0)
		case tcell.KeyUp, tcell.KeyBacktab:
			return tcell.NewEventKey(tcell.KeyBacktab, 0, 0)
		default:
			return event
		}
	})

	return &parameterizedFormItem{
		parameter: param,
		item:      item,
	}
}

// Create a standard uint16 field
func createParameterizedUint16Field(param *config.Parameter) *parameterizedFormItem {
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
				param.Value = int(value)
			}
		}
	})
	item.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyTab:
			return tcell.NewEventKey(tcell.KeyTab, 0, 0)
		case tcell.KeyUp, tcell.KeyBacktab:
			return tcell.NewEventKey(tcell.KeyBacktab, 0, 0)
		default:
			return event
		}
	})

	return &parameterizedFormItem{
		parameter: param,
		item:      item,
	}
}

// Create a standard string field
func createParameterizedStringField(param *config.Parameter) *parameterizedFormItem {
	item := tview.NewInputField().
		SetLabel(param.Name)
	item.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			item.SetText("")
		} else {
			param.Value = item.GetText()
		}
	})
	item.SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
		if param.MaxLength > 0 {
			if len(textToCheck) > param.MaxLength {
				return false
			}
		}
		// TODO: regex support
		return true
	})
	item.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyTab:
			return tcell.NewEventKey(tcell.KeyTab, 0, 0)
		case tcell.KeyUp, tcell.KeyBacktab:
			return tcell.NewEventKey(tcell.KeyBacktab, 0, 0)
		default:
			return event
		}
	})

	return &parameterizedFormItem{
		parameter: param,
		item:      item,
	}
}

// Create a standard choice field
func createParameterizedDropDown(param *config.Parameter, descriptionBox *tview.TextView) *parameterizedFormItem {
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
			descriptionBox.SetText(descriptions[index])
		})
	item.SetTextOptions(" ", " ", "", "", "")
	item.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyTab:
			return tcell.NewEventKey(tcell.KeyTab, 0, 0)
		case tcell.KeyUp, tcell.KeyBacktab:
			return tcell.NewEventKey(tcell.KeyBacktab, 0, 0)
		default:
			return event
		}
	})
	list := item.GetList()
	list.SetSelectedBackgroundColor(tcell.ColorLightGreen)
	list.SetSelectedTextColor(tcell.ColorDarkGreen)
	list.SetBackgroundColor(tcell.ColorBlack)
	list.SetMainTextColor(tcell.ColorLightGray)

	return &parameterizedFormItem{
		parameter: param,
		item:      item,
	}
}
