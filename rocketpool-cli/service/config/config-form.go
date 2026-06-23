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
	parameter     *cfgtypes.Parameter
	item          tview.FormItem
	onCommitError func(message string)
}

func (pfi *parameterizedFormItem) reportCommitError(message string) {
	if pfi.onCommitError != nil {
		pfi.onCommitError(message)
	}
}

func (pfi *parameterizedFormItem) clearCommitError() {
	if pfi.onCommitError != nil {
		pfi.onCommitError("")
	}
}

func (pfi *parameterizedFormItem) commit() {
	switch pfi.item.(type) {
	case *tview.Checkbox:
		pfi.parameter.Value = pfi.item.(*tview.Checkbox).IsChecked()
		pfi.clearCommitError()
	case *tview.InputField:
		var err error
		inputField := pfi.item.(*tview.InputField)
		switch pfi.parameter.Type {
		case cfgtypes.ParameterType_Int:
			pfi.parameter.Value, err = strconv.ParseInt(inputField.GetText(), 0, 0)
			if err != nil {
				pfi.reportCommitError(fmt.Sprintf("INVALID INTEGER VALUE FOR %s", pfi.parameter.Name))
				return
			}
		case cfgtypes.ParameterType_Uint:
			pfi.parameter.Value, err = strconv.ParseUint(inputField.GetText(), 0, 0)
			if err != nil {
				pfi.reportCommitError(fmt.Sprintf("INVALID UNSIGNED INTEGER VALUE FOR %s", pfi.parameter.Name))
				return
			}
		case cfgtypes.ParameterType_Uint16:
			pfi.parameter.Value, err = strconv.ParseUint(inputField.GetText(), 0, 16)
			if err != nil {
				pfi.reportCommitError(fmt.Sprintf("INVALID VALUE FOR %s (MUST BE 0–65535)", pfi.parameter.Name))
				return
			}
		case cfgtypes.ParameterType_String, cfgtypes.ParameterType_Float:
			pfi.parameter.Value = strings.TrimSpace(inputField.GetText())
		default:
			panic(fmt.Sprintf("Unknown parameter type for text field %v", pfi.parameter.Type))
		}
		pfi.clearCommitError()
	default:
		panic(fmt.Sprintf("Unknown form item type %v", pfi.item))
	}
}

// Create a list of form items based on a set of parameters
func createParameterizedFormItems(params []*cfgtypes.Parameter, layout *standardLayout) []*parameterizedFormItem {
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
			item = createParameterizedDropDown(param, layout.descriptionBox)
		default:
			panic(fmt.Sprintf("Unknown parameter type %v", param))
		}
		if layout != nil {
			item.onCommitError = layout.showCommitError
			layout.setItemNavCapture(item)
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
	item.SetChangedFunc(func(checked bool) {
		out.commit()
	})

	return out
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
