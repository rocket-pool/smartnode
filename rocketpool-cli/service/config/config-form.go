package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// A form item linked to a Parameter
type parameterizedFormItem struct {
	parameter *cfgtypes.Parameter
	item      tview.FormItem
}

func (pfi *parameterizedFormItem) commit() {
	switch pfi.item.(type) {
	case *tview.Checkbox:
		pfi.parameter.Value = pfi.item.(*tview.Checkbox).IsChecked()
	case *tview.InputField:
		var err error
		inputField := pfi.item.(*tview.InputField)
		switch pfi.parameter.Type {
		case cfgtypes.ParameterType_Int:
			pfi.parameter.Value, err = strconv.ParseInt(inputField.GetText(), 0, 0)
			if err != nil {
				// TODO: show error modal?
				inputField.SetText("")
			}
		case cfgtypes.ParameterType_Uint:
			pfi.parameter.Value, err = strconv.ParseUint(inputField.GetText(), 0, 0)
			if err != nil {
				// TODO: show error modal?
				inputField.SetText("")
			}
		case cfgtypes.ParameterType_Uint16:
			pfi.parameter.Value, err = strconv.ParseUint(inputField.GetText(), 0, 16)
			if err != nil {
				// TODO: show error modal?
				inputField.SetText("")
			}
		case cfgtypes.ParameterType_String, cfgtypes.ParameterType_Float:
			pfi.parameter.Value = strings.TrimSpace(inputField.GetText())
		default:
			panic(fmt.Sprintf("Unknown parameter type for text field %v", pfi.parameter.Type))
		}
	default:
		panic(fmt.Sprintf("Unknown form item type %v", pfi.item))
	}
}

// Create a list of form items based on a set of parameters
func createParameterizedFormItems(params []*cfgtypes.Parameter, descriptionBox *tview.TextView) []*parameterizedFormItem {
	formItems := []*parameterizedFormItem{}
	for _, param := range params {
		var item *parameterizedFormItem
		switch param.Type {
		case cfgtypes.ParameterType_Bool:
			item = createParameterizedCheckbox(param)
		case cfgtypes.ParameterType_Int, cfgtypes.ParameterType_Uint, cfgtypes.ParameterType_Uint16:
			item = createParameterizedIntField(param)
		case cfgtypes.ParameterType_String, cfgtypes.ParameterType_Float:
			item = createParameterizedStringField(param)
		case cfgtypes.ParameterType_Choice:
			item = createParameterizedDropDown(param, descriptionBox)
		default:
			panic(fmt.Sprintf("Unknown parameter type %v", param))
		}
		formItems = append(formItems, item)
	}

	return formItems
}

// Create a standard form checkbox
func createParameterizedCheckbox(param *cfgtypes.Parameter) *parameterizedFormItem {
	item := tview.NewCheckbox().
		SetLabel(param.Name).
		SetChecked(param.Value == true)
	out := &parameterizedFormItem{
		parameter: param,
		item:      item,
	}
	item.SetInputCapture(navCapture)
	item.SetChangedFunc(func(checked bool) {
		out.commit()
	})

	return out
}

func navCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyDown, tcell.KeyTab:
		return tcell.NewEventKey(tcell.KeyTab, 0, 0)
	case tcell.KeyUp, tcell.KeyBacktab:
		return tcell.NewEventKey(tcell.KeyBacktab, 0, 0)
	}
	return event
}

// Create a standard int field
func createParameterizedIntField(param *cfgtypes.Parameter) *parameterizedFormItem {
	item := tview.NewInputField().
		SetLabel(param.Name).
		SetAcceptanceFunc(tview.InputFieldInteger)
	out := &parameterizedFormItem{
		parameter: param,
		item:      item,
	}
	item.SetInputCapture(navCapture)
	item.SetDoneFunc(func(key tcell.Key) {
		out.commit()
	})

	return out
}

// Create a standard string field
func createParameterizedStringField(param *cfgtypes.Parameter) *parameterizedFormItem {
	item := tview.NewInputField().
		SetLabel(param.Name)
	out := &parameterizedFormItem{
		parameter: param,
		item:      item,
	}
	item.SetDoneFunc(func(key tcell.Key) {
		out.commit()
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
	item.SetInputCapture(navCapture)

	return out
}

// Create a standard choice field
func createParameterizedDropDown(param *cfgtypes.Parameter, descriptionBox *tview.TextView) *parameterizedFormItem {
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
	item.SetInputCapture(navCapture)
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
