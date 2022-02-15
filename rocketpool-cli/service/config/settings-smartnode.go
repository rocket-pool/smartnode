package config

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

type SmartnodeConfigPage struct {
	home     *settingsHome
	page     *page
	layout   *standardLayout
	paramMap map[string]tview.FormItem
}

// Creates a new page for the Smartnode settings
func NewSmartnodeConfigPage(home *settingsHome) *SmartnodeConfigPage {

	configPage := &SmartnodeConfigPage{
		home:     home,
		paramMap: map[string]tview.FormItem{},
	}

	configPage.createSettingSmartnodeContent()
	configPage.page = newPage(
		home.homePage,
		"settings-smartnode",
		"Smartnode and TX Fees",
		"Select this to configure the settings for the Smartnode itself, including the defaults and limits on transaction fees.",
		configPage.layout.grid,
	)

	return configPage

}

// Creates the content for the Smartnode settings page
func (configPage *SmartnodeConfigPage) createSettingSmartnodeContent() {

	layout := newStandardLayout()
	form := NewForm().
		SetFieldBackgroundColor(tcell.ColorBlack)
	form.
		SetBackgroundColor(tview.Styles.ContrastBackgroundColor).
		SetBorderPadding(0, 0, 0, 0)

	masterConfig := configPage.home.md.config
	for _, param := range masterConfig.Smartnode.GetParameters() {
		switch param.Type {
		// Boolean (checkbox)
		case config.ParameterType_Bool:
			item := tview.NewCheckbox().
				SetLabel(param.Name).
				SetChecked(param.Value == true).
				SetChangedFunc(func(checked bool) {
					param.Value = checked
				})
			item.SetFocusFunc(func() {
				defaultValue, err := param.GetDefault(masterConfig)
				if err != nil {
					layout.descriptionBox.SetText(fmt.Sprintf("<Error creating Smartnode view: %s>", err.Error()))
				} else {
					layout.descriptionBox.SetText(fmt.Sprintf("Default: %t\n\n%s", defaultValue, param.Description))
				}
			})
			form.AddFormItem(item)
			configPage.paramMap[param.ID] = item
			item.SetChecked(param.Value == true)

		// Int (textbox)
		case config.ParameterType_Int:
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
						param.Value = value
					}
				}
			})
			item.SetFocusFunc(func() {
				layout.descriptionBox.SetText(param.Description)
			})
			form.AddFormItem(item)
			configPage.paramMap[param.ID] = item
			item.SetText(fmt.Sprintf("%d", param.Value))

		// Uint (textbox)
		case config.ParameterType_Uint:
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
			item.SetFocusFunc(func() {
				layout.descriptionBox.SetText(param.Description)
			})
			form.AddFormItem(item)
			configPage.paramMap[param.ID] = item
			item.SetText(fmt.Sprintf("%d", param.Value))

		// Uint16 (textbox)
		case config.ParameterType_Uint16:
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
						param.Value = value
					}
				}
			})
			item.SetFocusFunc(func() {
				layout.descriptionBox.SetText(param.Description)
			})
			form.AddFormItem(item)
			configPage.paramMap[param.ID] = item
			item.SetText(fmt.Sprintf("%d", param.Value))

		// String (textbox)
		case config.ParameterType_String:
			item := tview.NewInputField().
				SetLabel(param.Name)
			item.SetDoneFunc(func(key tcell.Key) {
				if key == tcell.KeyEscape {
					item.SetText("")
				} else {
					param.Value = item.GetText()
				}
			})
			item.SetFocusFunc(func() {
				layout.descriptionBox.SetText(param.Description)
			})
			form.AddFormItem(item)
			configPage.paramMap[param.ID] = item
			item.SetText(param.Value.(string))

		// Choice (dropdown)
		case config.ParameterType_Choice:
			// Create the list of options
			options := []string{}
			descriptions := []string{}
			values := []interface{}{}
			for _, option := range param.Options {
				options = append(options, option.Name)
				descriptions = append(descriptions, option.Description)
				values = append(values, option.Value)
			}

			// Create the dropdown
			item := NewDropDown().
				SetLabel(param.Name).
				SetOptions(options, func(text string, index int) {
					param.Value = values[index]
				}).
				SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
					layout.descriptionBox.SetText(descriptions[index])
				})
			item.SetFocusFunc(func() {
				layout.descriptionBox.SetText(param.Description)
			})
			item.SetTextOptions(" ", " ", "", "", "")
			form.AddFormItem(item)
			configPage.paramMap[param.ID] = item
			for i := 0; i < len(values); i++ {
				if values[i] == param.Value {
					item.SetCurrentOption(i)
					break
				}
			}

		}

	}

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			configPage.home.md.setPage(configPage.home.homePage)
			return nil
		}
		return event
	})

	// Make it the content of the layout and set the default description text
	layout.setContent(form, form.Box, "Smartnode and TX Fee Settings")

	paramDescriptions := []string{}
	for _, param := range masterConfig.Smartnode.GetParameters() {
		paramDescriptions = append(paramDescriptions, param.Description)
		layout.descriptionBox.ScrollToBeginning()
	}
	layout.descriptionBox.SetText(paramDescriptions[0])
	form.SetChangedFunc(func(index int) {
		layout.descriptionBox.SetText(paramDescriptions[index])
		layout.descriptionBox.ScrollToBeginning()
	})

	// Make the footer
	footer, height := createSettingFooter()
	layout.setFooter(footer, height)

	// Return the standard layout's grid
	configPage.layout = layout

}

// Create the footer, including the nav bar and the save / quit buttons
func createSettingFooter() (*tview.Flex, int) {

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

	return navBar, 2

}
