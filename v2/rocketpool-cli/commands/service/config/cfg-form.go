package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/node-manager-core/config"
)

// A form item linked to a Parameter
type parameterizedFormItem struct {
	parameter config.IParameter
	item      tview.FormItem
}

func registerEnableCheckbox(param *config.Parameter[bool], checkbox *tview.Checkbox, form *Form, items []*parameterizedFormItem) {
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
func createParameterizedFormItems(params []config.IParameter, descriptionBox *tview.TextView) []*parameterizedFormItem {
	formItems := []*parameterizedFormItem{}
	for _, param := range params {
		item := getTypedFormItem(param, descriptionBox)
		formItems = append(formItems, item)
	}
	return formItems
}

// Create a form item binding for a parameter based on its type
func getTypedFormItem(param config.IParameter, descriptionBox *tview.TextView) *parameterizedFormItem {
	if len(param.GetOptions()) > 0 {
		return createParameterizedDropDown(param, descriptionBox)
	}
	if boolParam, ok := param.(*config.Parameter[bool]); ok {
		return createParameterizedCheckbox(boolParam)
	}
	if intParam, ok := param.(*config.Parameter[int]); ok {
		return createParameterizedIntField(intParam)
	}
	if uintParam, ok := param.(*config.Parameter[uint64]); ok {
		return createParameterizedUintField(uintParam)
	}
	if uint16Param, ok := param.(*config.Parameter[uint16]); ok {
		return createParameterizedUint16Field(uint16Param)
	}
	if stringParam, ok := param.(*config.Parameter[string]); ok {
		return createParameterizedStringField(stringParam)
	}
	if floatParam, ok := param.(*config.Parameter[float64]); ok {
		return createParameterizedFloatField(floatParam)
	}
	panic(fmt.Sprintf("param [%s] is not a supported type for form item binding", param.GetCommon().Name))
}

// Create a standard form checkbox
func createParameterizedCheckbox(param *config.Parameter[bool]) *parameterizedFormItem {
	item := tview.NewCheckbox().
		SetLabel(param.Name).
		SetChecked(param.Value).
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
func createParameterizedIntField(param *config.Parameter[int]) *parameterizedFormItem {
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
func createParameterizedUintField(param *config.Parameter[uint64]) *parameterizedFormItem {
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
				param.Value = value
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
func createParameterizedUint16Field(param *config.Parameter[uint16]) *parameterizedFormItem {
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
				param.Value = uint16(value)
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
func createParameterizedStringField(param *config.Parameter[string]) *parameterizedFormItem {
	item := tview.NewInputField().
		SetLabel(param.Name)
	item.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			item.SetText("")
		} else {
			param.Value = strings.TrimSpace(item.GetText())
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

// Create a standard float field
func createParameterizedFloatField(param *config.Parameter[float64]) *parameterizedFormItem {
	item := tview.NewInputField().
		SetLabel(param.Name).
		SetAcceptanceFunc(tview.InputFieldFloat)
	item.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			item.SetText("")
		} else {
			value, err := strconv.ParseFloat(item.GetText(), 64)
			if err != nil {
				// TODO: show error modal?
				item.SetText("")
			} else {
				param.Value = value
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

// Create a standard choice field
func createParameterizedDropDown(param config.IParameter, descriptionBox *tview.TextView) *parameterizedFormItem {
	// Create the list of options
	options := []string{}
	descriptions := []string{}
	values := []any{}
	for _, option := range param.GetOptions() {
		common := option.Common()
		options = append(options, common.Name)
		descriptions = append(descriptions, common.Description)
		values = append(values, option.GetValueAsAny())
	}
	item := NewDropDown().
		SetLabel(param.GetCommon().Name).
		SetOptions(options, func(text string, index int) {
			param.SetValue(values[index])
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
	list.SetSelectedBackgroundColor(tcell.Color46)
	list.SetSelectedTextColor(tcell.ColorBlack)
	list.SetBackgroundColor(tcell.ColorBlack)
	list.SetMainTextColor(tcell.ColorLightGray)

	return &parameterizedFormItem{
		parameter: param,
		item:      item,
	}
}
