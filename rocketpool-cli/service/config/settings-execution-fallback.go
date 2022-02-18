package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The page wrapper for the EC config
type FallbackExecutionConfigPage struct {
	home   *settingsHome
	page   *page
	layout *standardLayout
}

// Creates a new page for the fallback Execution client settings
func NewFallbackExecutionConfigPage(home *settingsHome) *FallbackExecutionConfigPage {

	configPage := &FallbackExecutionConfigPage{
		home: home,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-execution-fallback",
		"Execution Backup (Eth1 Fallback)",
		"Select this to choose your fallback / backup Execution Client (formerly called \"ETH1 fallback client\") that the Smartnode and Beacon client will use if your main Execution client ever goes offline.",
		configPage.layout.grid,
	)

	return configPage

}

// Creates the content for the fallback Execution client settings page
func (configPage *FallbackExecutionConfigPage) createContent() {

	// Create the layout
	masterConfig := configPage.home.md.config
	layout := newStandardLayout()
	configPage.layout = layout
	layout.createForm(&masterConfig.Smartnode.Network, "Execution Client (Eth1) Settings")

	// Return to the home page after pressing Escape
	layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			configPage.home.md.setPage(configPage.home.homePage)
			return nil
		}
		return event
	})

	// Set up the form items
	useFallbackEcBox := createParameterizedCheckbox(&masterConfig.UseFallbackExecutionClient)
	fallbackEcModeDropdown := createParameterizedDropDown(&masterConfig.FallbackExecutionClientMode, layout.descriptionBox)
	fallbackEcDropdown := createParameterizedDropDown(&masterConfig.FallbackExecutionClient, layout.descriptionBox)
	fallbackEcCommonItems := createParameterizedFormItems(masterConfig.FallbackExecutionCommon.GetParameters(), layout.descriptionBox)
	fallbackInfuraItems := createParameterizedFormItems(masterConfig.FallbackInfura.GetParameters(), layout.descriptionBox)
	fallbackPocketItems := createParameterizedFormItems(masterConfig.FallbackPocket.GetParameters(), layout.descriptionBox)
	fallbackExternalECItems := createParameterizedFormItems(masterConfig.FallbackExternalExecution.GetParameters(), layout.descriptionBox)

	layout.mapParameterizedFormItems(useFallbackEcBox, fallbackEcModeDropdown, fallbackEcDropdown)
	layout.mapParameterizedFormItems(fallbackEcCommonItems...)
	layout.mapParameterizedFormItems(fallbackInfuraItems...)
	layout.mapParameterizedFormItems(fallbackPocketItems...)
	layout.mapParameterizedFormItems(fallbackExternalECItems...)

	useFallbackEcBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if masterConfig.UseFallbackExecutionClient.Value == checked {
			return
		}

		masterConfig.UseFallbackExecutionClient.Value = checked

		layout.form.Clear(true)
		layout.form.AddFormItem(useFallbackEcBox.item)

		// Only add the supporting stuff if external clients are enabled
		if !checked {
			return
		}

		layout.form.AddFormItem(fallbackEcModeDropdown.item)
		selectedMode := masterConfig.FallbackExecutionClientMode.Value.(config.Mode)
		switch selectedMode {

		// Local (Docker mode)
		case config.Mode_Local:
			layout.form.AddFormItem(fallbackEcDropdown.item)
			selectedEc := masterConfig.FallbackExecutionClient.Value.(config.ExecutionClient)

			switch selectedEc {
			case config.ExecutionClient_Infura:
				layout.addFormItemsWithCommonParams(fallbackEcCommonItems, fallbackInfuraItems, masterConfig.FallbackInfura.UnsupportedCommonParams)
			case config.ExecutionClient_Pocket:
				layout.addFormItemsWithCommonParams(fallbackEcCommonItems, fallbackPocketItems, masterConfig.FallbackPocket.UnsupportedCommonParams)
			}

		// External (Hybrid mode)
		case config.Mode_External:
			for _, param := range fallbackExternalECItems {
				layout.form.AddFormItem(param.item)
			}
		}

		layout.refresh()
	})

	fallbackEcModeDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if masterConfig.FallbackExecutionClientMode.Value == masterConfig.FallbackExecutionClientMode.Options[index].Value {
			return
		}
		masterConfig.FallbackExecutionClientMode.Value = masterConfig.FallbackExecutionClientMode.Options[index].Value

		layout.form.Clear(true)
		layout.form.AddFormItem(useFallbackEcBox.item)
		layout.form.AddFormItem(fallbackEcModeDropdown.item)

		selectedMode := masterConfig.FallbackExecutionClientMode.Value.(config.Mode)
		switch selectedMode {

		// Local (Docker mode)
		case config.Mode_Local:
			layout.form.AddFormItem(fallbackEcDropdown.item)
			selectedEc := masterConfig.FallbackExecutionClient.Value.(config.ExecutionClient)

			switch selectedEc {
			case config.ExecutionClient_Infura:
				layout.addFormItemsWithCommonParams(fallbackEcCommonItems, fallbackInfuraItems, masterConfig.FallbackInfura.UnsupportedCommonParams)
			case config.ExecutionClient_Pocket:
				layout.addFormItemsWithCommonParams(fallbackEcCommonItems, fallbackPocketItems, masterConfig.FallbackPocket.UnsupportedCommonParams)
			}

		// External (Hybrid mode)
		case config.Mode_External:
			for _, param := range fallbackExternalECItems {
				layout.form.AddFormItem(param.item)
			}
		}

		layout.refresh()
	})

	fallbackEcDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if masterConfig.FallbackExecutionClient.Value == masterConfig.FallbackExecutionClient.Options[index].Value {
			return
		}
		masterConfig.FallbackExecutionClient.Value = masterConfig.FallbackExecutionClient.Options[index].Value

		layout.form.Clear(true)
		layout.form.AddFormItem(useFallbackEcBox.item)
		layout.form.AddFormItem(fallbackEcModeDropdown.item)
		layout.form.AddFormItem(fallbackEcDropdown.item)
		selectedEc := masterConfig.FallbackExecutionClient.Value.(config.ExecutionClient)

		switch selectedEc {
		case config.ExecutionClient_Infura:
			layout.addFormItemsWithCommonParams(fallbackEcCommonItems, fallbackInfuraItems, masterConfig.FallbackInfura.UnsupportedCommonParams)
		case config.ExecutionClient_Pocket:
			layout.addFormItemsWithCommonParams(fallbackEcCommonItems, fallbackPocketItems, masterConfig.FallbackPocket.UnsupportedCommonParams)
		}

		layout.refresh()

	})

	layout.form.Clear(true)
	layout.form.AddFormItem(useFallbackEcBox.item)

	// Only add the supporting stuff if external clients are enabled
	if masterConfig.UseFallbackExecutionClient.Value == false {
		return
	}

	layout.form.AddFormItem(fallbackEcModeDropdown.item)
	selectedMode := masterConfig.FallbackExecutionClientMode.Value.(config.Mode)
	switch selectedMode {

	// Local (Docker mode)
	case config.Mode_Local:
		layout.form.AddFormItem(fallbackEcDropdown.item)
		selectedEc := masterConfig.FallbackExecutionClient.Value.(config.ExecutionClient)

		switch selectedEc {
		case config.ExecutionClient_Infura:
			layout.addFormItemsWithCommonParams(fallbackEcCommonItems, fallbackInfuraItems, masterConfig.FallbackInfura.UnsupportedCommonParams)
		case config.ExecutionClient_Pocket:
			layout.addFormItemsWithCommonParams(fallbackEcCommonItems, fallbackPocketItems, masterConfig.FallbackPocket.UnsupportedCommonParams)
		}

	// External (Hybrid mode)
	case config.Mode_External:
		for _, param := range fallbackExternalECItems {
			layout.form.AddFormItem(param.item)
		}
	}

	layout.refresh()
}
