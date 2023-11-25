package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/types/addons"
)

// The page wrapper for the APCUPSD addon config
type AddonApcupsdPage struct {
	addonsPage     *AddonsPage
	page           *page
	layout         *standardLayout
	masterConfig   *config.RocketPoolConfig
	addon          addons.SmartnodeAddon
	enabledBox     *parameterizedFormItem
	modeBox        *parameterizedFormItem
	exporterImage  *parameterizedFormItem
	apcupsdImage   *parameterizedFormItem
	mountPoint     *parameterizedFormItem
	metricsPort    *parameterizedFormItem
	apcupsdAddress *parameterizedFormItem
}

// Creates a new page for the APCUPSD addon settings
func NewAddonApcupsdPage(addonsPage *AddonsPage, addon addons.SmartnodeAddon) *AddonApcupsdPage {

	configPage := &AddonApcupsdPage{
		addonsPage:   addonsPage,
		masterConfig: addonsPage.home.md.Config,
		addon:        addon,
	}
	configPage.createContent()

	configPage.page = newPage(
		addonsPage.page,
		"settings-addon-apcupsd",
		addon.GetName(),
		addon.GetDescription(),
		configPage.layout.grid,
	)

	return configPage

}

// Get the underlying page
func (configPage *AddonApcupsdPage) getPage() *page {
	return configPage.page
}

// Creates the content for the APCUPSD settings page
func (configPage *AddonApcupsdPage) createContent() {

	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, fmt.Sprintf("%s Settings", configPage.addon.GetName()))

	// Return to the home page after pressing Escape
	configPage.layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			// Close all dropdowns and break if one was open
			for _, param := range configPage.layout.parameters {
				dropDown, ok := param.item.(*DropDown)
				if ok && dropDown.open {
					dropDown.CloseList(configPage.addonsPage.home.md.app)
					return nil
				}
			}

			// Return to the home page
			configPage.addonsPage.home.md.setPage(configPage.addonsPage.page)
			return nil
		}
		return event
	})

	// Get the parameters
	enabledParam := configPage.addon.GetEnabledParameter()
	// TODO: Don't like how I reference these by index here. Is there a better way?
	modeParam := configPage.addon.GetConfig().GetParameters()[1]
	exporterImageParam := configPage.addon.GetConfig().GetParameters()[2]
	apcupsdImageParam := configPage.addon.GetConfig().GetParameters()[3]
	metricsPortParam := configPage.addon.GetConfig().GetParameters()[4]
	mountPointParam := configPage.addon.GetConfig().GetParameters()[5]
	networkAddress := configPage.addon.GetConfig().GetParameters()[6]

	// Set up the form items
	configPage.enabledBox = createParameterizedCheckbox(enabledParam)
	configPage.modeBox = createParameterizedDropDown(modeParam, configPage.layout.descriptionBox)
	configPage.exporterImage = createParameterizedStringField(exporterImageParam)
	configPage.apcupsdImage = createParameterizedStringField(apcupsdImageParam)
	configPage.metricsPort = createParameterizedStringField(metricsPortParam)
	configPage.mountPoint = createParameterizedStringField(mountPointParam)
	configPage.apcupsdAddress = createParameterizedStringField(networkAddress)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enabledBox)
	configPage.layout.mapParameterizedFormItems(configPage.modeBox)
	configPage.layout.mapParameterizedFormItems(configPage.exporterImage)
	configPage.layout.mapParameterizedFormItems(configPage.apcupsdImage)
	configPage.layout.mapParameterizedFormItems(configPage.metricsPort)
	configPage.layout.mapParameterizedFormItems(configPage.mountPoint)
	configPage.layout.mapParameterizedFormItems(configPage.apcupsdAddress)

	// Set up the setting callbacks
	configPage.enabledBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if enabledParam.Value == checked {
			return
		}
		enabledParam.Value = checked
		configPage.handleEnableChanged()
	})
	configPage.modeBox.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.modeBox.parameter.Value == configPage.modeBox.parameter.Options[index].Value {
			return
		}
		configPage.modeBox.parameter.Value = configPage.modeBox.parameter.Options[index].Value
		configPage.handleModeChanged()
	})

	// Do the initial draw
	configPage.handleDraw()

}

// Handle all of the form changes when the Enabled box has changed
func (configPage *AddonApcupsdPage) handleEnableChanged() {
	configPage.handleDraw()
}

// Handle all of the form changes when the Mode box has changed
func (configPage *AddonApcupsdPage) handleModeChanged() {
	configPage.handleDraw()
}

func (configPage *AddonApcupsdPage) handleDraw() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enabledBox.item)

	// Only add the supporting stuff if addon is enabled
	if configPage.addon.GetEnabledParameter().Value == false {
		return
	}
	configPage.addCommonFields()
	if configPage.modeBox.parameter.Value == configPage.modeBox.parameter.Options[0].Value {
		configPage.addContainerFields()
	} else {
		configPage.addNetworkFields()
	}
	configPage.layout.refresh()
}

func (configPage *AddonApcupsdPage) addCommonFields() {
	configPage.layout.form.AddFormItem(configPage.modeBox.item)
	configPage.layout.form.AddFormItem(configPage.exporterImage.item)
	configPage.layout.form.AddFormItem(configPage.metricsPort.item)
}
func (configPage *AddonApcupsdPage) addContainerFields() {
	configPage.layout.form.AddFormItem(configPage.apcupsdImage.item)
	configPage.layout.form.AddFormItem(configPage.mountPoint.item)
}

func (configPage *AddonApcupsdPage) addNetworkFields() {
	configPage.layout.form.AddFormItem(configPage.apcupsdAddress.item)
}

// Handle a bulk redraw request
func (configPage *AddonApcupsdPage) handleLayoutChanged() {
	configPage.handleEnableChanged()
}
