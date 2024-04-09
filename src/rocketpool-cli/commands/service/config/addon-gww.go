package config

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/rocket-pool/node-manager-core/config"
	gww "github.com/rocket-pool/smartnode/v2/addons/graffiti_wall_writer"
	snCfg "github.com/rocket-pool/smartnode/v2/shared/config"
)

// The page wrapper for the Graffiti Wall Writer addon config
type AddonGwwPage struct {
	addonsPage   *AddonsPage
	page         *page
	layout       *standardLayout
	masterConfig *snCfg.SmartNodeConfig
	gwwConfig    *gww.GraffitiWallWriterConfig
	gww          *gww.GraffitiWallWriter
	enabledBox   *parameterizedFormItem
	otherParams  []*parameterizedFormItem
}

// Creates a new page for the Graffiti Wall Writer addon settings
func NewAddonGwwPage(addonsPage *AddonsPage, gwwConfig *gww.GraffitiWallWriterConfig) *AddonGwwPage {
	gww := gww.NewGraffitiWallWriter(gwwConfig)
	configPage := &AddonGwwPage{
		addonsPage:   addonsPage,
		masterConfig: addonsPage.home.md.Config,
		gwwConfig:    gwwConfig,
		gww:          gww,
	}
	configPage.createContent()

	configPage.page = newPage(
		addonsPage.page,
		"settings-addon-gww",
		gww.GetName(),
		gww.GetDescription(),
		configPage.layout.grid,
	)

	return configPage
}

// Get the underlying page
func (configPage *AddonGwwPage) getPage() *page {
	return configPage.page
}

// Creates the content for the GWW settings page
func (configPage *AddonGwwPage) createContent() {
	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Network, fmt.Sprintf("%s Settings", configPage.gww.GetName()))
	configPage.layout.setupEscapeReturnHomeHandler(configPage.addonsPage.home.md, configPage.addonsPage.page)

	// Get the parameters
	enabledParam := &configPage.gwwConfig.Enabled
	otherParams := []config.IParameter{}

	for _, param := range configPage.gwwConfig.GetParameters() {
		if param.GetCommon().ID != enabledParam.ID {
			otherParams = append(otherParams, param)
		}
	}

	// Set up the form items
	configPage.enabledBox = createParameterizedCheckbox(enabledParam)
	configPage.otherParams = createParameterizedFormItems(otherParams, configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enabledBox)
	configPage.layout.mapParameterizedFormItems(configPage.otherParams...)

	// Set up the setting callbacks
	configPage.enabledBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if enabledParam.Value == checked {
			return
		}
		enabledParam.Value = checked
		configPage.handleEnableChanged()
	})

	// Do the initial draw
	configPage.handleEnableChanged()
}

// Handle all of the form changes when the Use Fallback EC box has changed
func (configPage *AddonGwwPage) handleEnableChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enabledBox.item)

	// Only add the supporting stuff if external clients are enabled
	if !configPage.gwwConfig.Enabled.Value {
		return
	}
	configPage.layout.addFormItems(configPage.otherParams)
	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *AddonGwwPage) handleLayoutChanged() {
	configPage.handleEnableChanged()
}
