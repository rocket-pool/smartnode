package config

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/v2/addons/rescue_node"
	snCfg "github.com/rocket-pool/smartnode/v2/shared/config"
)

// The page wrapper for the Rescue Node addon config
type AddonRescueNodePage struct {
	addonsPage   *AddonsPage
	page         *page
	layout       *standardLayout
	masterConfig *snCfg.SmartNodeConfig
	rnConfig     *rescue_node.RescueNodeConfig
	rn           *rescue_node.RescueNode
	enabledBox   *parameterizedFormItem
	otherParams  []*parameterizedFormItem
}

// Creates a new page for the Graffiti Wall Writer addon settings
func NewAddonRescueNodePage(addonsPage *AddonsPage, rnConfig *rescue_node.RescueNodeConfig) *AddonRescueNodePage {
	rn := rescue_node.NewRescueNode(rnConfig)
	configPage := &AddonRescueNodePage{
		addonsPage:   addonsPage,
		masterConfig: addonsPage.home.md.Config,
		rnConfig:     rnConfig,
		rn:           rn,
	}
	configPage.createContent()

	configPage.page = newPage(
		addonsPage.page,
		"settings-addon-rescue-node",
		rn.GetName(),
		rn.GetDescription(),
		configPage.layout.grid,
	)

	return configPage
}

// Get the underlying page
func (configPage *AddonRescueNodePage) getPage() *page {
	return configPage.page
}

// Creates the content for the Rescue Node settings page
func (configPage *AddonRescueNodePage) createContent() {
	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Network, fmt.Sprintf("%s Settings", configPage.rn.GetName()))
	configPage.layout.setupEscapeReturnHomeHandler(configPage.addonsPage.home.md, configPage.addonsPage.page)

	// Get the parameters
	enabledParam := &configPage.rnConfig.Enabled
	otherParams := []config.IParameter{}
	for _, param := range configPage.rnConfig.GetParameters() {
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

// Handle all of the form changes when the Enabled box has changed
func (configPage *AddonRescueNodePage) handleEnableChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enabledBox.item)

	// Only add the supporting stuff if the rescue node is enabled
	if !configPage.rnConfig.Enabled.Value {
		return
	}
	configPage.layout.addFormItems(configPage.otherParams)
	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *AddonRescueNodePage) handleLayoutChanged() {
	configPage.handleEnableChanged()
}
