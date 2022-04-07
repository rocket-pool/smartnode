package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The page wrapper for the IPFS config
type IpfsConfigPage struct {
	home          *settingsHome
	page          *page
	layout        *standardLayout
	masterConfig  *config.RocketPoolConfig
	enableIpfsBox *parameterizedFormItem
	ipfsItems     []*parameterizedFormItem
}

// Creates a new page for the IPFS settings
func NewIpfsConfigPage(home *settingsHome) *IpfsConfigPage {

	configPage := &IpfsConfigPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-ipfs",
		"IPFS",
		"Select this to configure the optional InterPlanetary File System (IPFS) node. This feature is used to store and host copies of the claims and proofs for all of Rocket Pool's rewards periods to help promote the network's decentralization.",
		configPage.layout.grid,
	)

	return configPage

}

// Creates the content for the IPFS settings page
func (configPage *IpfsConfigPage) createContent() {

	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "IPFS Settings")

	// Return to the home page after pressing Escape
	configPage.layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Return to the home page
		if event.Key() == tcell.KeyEsc {
			// Close all dropdowns and break if one was open
			for _, param := range configPage.layout.parameters {
				dropDown, ok := param.item.(*DropDown)
				if ok && dropDown.open {
					dropDown.CloseList(configPage.home.md.app)
					return nil
				}
			}

			configPage.home.md.setPage(configPage.home.homePage)
			return nil
		}
		return event
	})

	// Set up the form items
	configPage.enableIpfsBox = createParameterizedCheckbox(&configPage.masterConfig.EnableIpfs)
	configPage.ipfsItems = createParameterizedFormItems(configPage.masterConfig.Ipfs.GetParameters(), configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enableIpfsBox)
	configPage.layout.mapParameterizedFormItems(configPage.ipfsItems...)

	// Set up the setting callbacks
	configPage.enableIpfsBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.EnableIpfs.Value == checked {
			return
		}
		configPage.masterConfig.EnableIpfs.Value = checked
		configPage.handleEnableIpfsChanged()
	})

	// Do the initial draw
	configPage.handleEnableIpfsChanged()
}

// Handle all of the form changes when the Enable IPFS box has changed
func (configPage *IpfsConfigPage) handleEnableIpfsChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enableIpfsBox.item)

	// Only add the supporting stuff if IPFS is enabled
	if configPage.masterConfig.EnableIpfs.Value == false {
		return
	}

	configPage.layout.addFormItems(configPage.ipfsItems)

	configPage.layout.refresh()
}
