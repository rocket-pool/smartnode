package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The page wrapper for the EC config
type ExecutionConfigPage struct {
	home   *settingsHome
	page   *page
	layout *standardLayout
}

// Creates a new page for the Execution client settings
func NewExecutionConfigPage(home *settingsHome) *ExecutionConfigPage {

	configPage := &ExecutionConfigPage{
		home: home,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-execution",
		"Execution Client (ETH1)",
		"Select this to choose your Execution client (formerly called \"ETH1 client\") and configure its settings.",
		configPage.layout.grid,
	)

	return configPage

}

// Creates the content for the Execution client settings page
func (configPage *ExecutionConfigPage) createContent() {

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
	ecModeDropdown := createParameterizedFormItems([]*config.Parameter{&masterConfig.ExecutionClientMode}, layout.descriptionBox)[0]
	ecDropdown := createParameterizedFormItems([]*config.Parameter{&masterConfig.ExecutionClient}, layout.descriptionBox)[0]
	ecCommonItems := createParameterizedFormItems(masterConfig.ExecutionCommon.GetParameters(), layout.descriptionBox)
	gethItems := createParameterizedFormItems(masterConfig.Geth.GetParameters(), layout.descriptionBox)
	infuraItems := createParameterizedFormItems(masterConfig.Infura.GetParameters(), layout.descriptionBox)
	pocketItems := createParameterizedFormItems(masterConfig.Pocket.GetParameters(), layout.descriptionBox)
	externalECItems := createParameterizedFormItems(masterConfig.ExternalExecution.GetParameters(), layout.descriptionBox)

	layout.mapParameterizedFormItems(ecModeDropdown, ecDropdown)
	layout.mapParameterizedFormItems(ecCommonItems...)
	layout.mapParameterizedFormItems(gethItems...)
	layout.mapParameterizedFormItems(infuraItems...)
	layout.mapParameterizedFormItems(pocketItems...)
	layout.mapParameterizedFormItems(externalECItems...)

	ecModeDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		// TEMP, MAKE 2 CALLBACKS THAT GET TRIGGERED INSTEAD OF DOING THIS
		masterConfig.ExecutionClientMode.Value = masterConfig.ExecutionClientMode.Options[index].Value

		layout.form.Clear(true)
		layout.form.AddFormItem(ecModeDropdown.item)

		selectedMode := masterConfig.ExecutionClientMode.Options[index].Value.(config.Mode)
		switch selectedMode {

		// Local (Docker mode)
		case config.Mode_Local:
			layout.form.AddFormItem(ecDropdown.item)
			selectedEc := masterConfig.ExecutionClient.Value.(config.ExecutionClient)

			switch selectedEc {
			case config.ExecutionClient_Geth:
				layout.addFormItemsWithCommonParams(ecCommonItems, gethItems, masterConfig.Geth.UnsupportedCommonParams)
			case config.ExecutionClient_Infura:
				layout.addFormItemsWithCommonParams(ecCommonItems, infuraItems, masterConfig.Infura.UnsupportedCommonParams)
			case config.ExecutionClient_Pocket:
				layout.addFormItemsWithCommonParams(ecCommonItems, pocketItems, masterConfig.Pocket.UnsupportedCommonParams)
			}

		// External (Hybrid mode)
		case config.Mode_External:
			for _, param := range externalECItems {
				layout.form.AddFormItem(param.item)
			}
		}
	})

	ecDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		masterConfig.ExecutionClient.Value = masterConfig.ExecutionClient.Options[index].Value
		layout.form.Clear(true)
		layout.form.AddFormItem(ecModeDropdown.item)
		layout.form.AddFormItem(ecDropdown.item)
		selectedEc := masterConfig.ExecutionClient.Value.(config.ExecutionClient)

		switch selectedEc {
		case config.ExecutionClient_Geth:
			layout.addFormItemsWithCommonParams(ecCommonItems, gethItems, masterConfig.Geth.UnsupportedCommonParams)
		case config.ExecutionClient_Infura:
			layout.addFormItemsWithCommonParams(ecCommonItems, infuraItems, masterConfig.Infura.UnsupportedCommonParams)
		case config.ExecutionClient_Pocket:
			layout.addFormItemsWithCommonParams(ecCommonItems, pocketItems, masterConfig.Pocket.UnsupportedCommonParams)
		}
	})

	// TEMP
	layout.form.Clear(true)
	layout.form.AddFormItem(ecModeDropdown.item)

	selectedMode := masterConfig.ExecutionClientMode.Value.(config.Mode)
	switch selectedMode {

	// Local (Docker mode)
	case config.Mode_Local:
		layout.form.AddFormItem(ecDropdown.item)
		selectedEc := masterConfig.ExecutionClient.Value.(config.ExecutionClient)

		switch selectedEc {
		case config.ExecutionClient_Geth:
			layout.addFormItemsWithCommonParams(ecCommonItems, gethItems, masterConfig.Geth.UnsupportedCommonParams)
		case config.ExecutionClient_Infura:
			layout.addFormItemsWithCommonParams(ecCommonItems, infuraItems, masterConfig.Infura.UnsupportedCommonParams)
		case config.ExecutionClient_Pocket:
			layout.addFormItemsWithCommonParams(ecCommonItems, pocketItems, masterConfig.Pocket.UnsupportedCommonParams)
		}

	// External (Hybrid mode)
	case config.Mode_External:
		for _, param := range externalECItems {
			layout.form.AddFormItem(param.item)
		}
	}

	layout.refresh()

}
