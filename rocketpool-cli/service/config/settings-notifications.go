package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
//	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// The page wrapper for the notifications config
type NotificationsConfigPage struct {
	home                             *settingsHome
	page                             *page
	layout                           *standardLayout
	masterConfig                     *config.RocketPoolConfig
	enableNotificationsBox           *parameterizedFormItem
}

// Creates a new page for the notifications settings
func NewNotificationsConfigPage(home *settingsHome) *NotificationsConfigPage {

	configPage := &NotificationsConfigPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-notifications",
		"Notifications",
		"Select this to configure notifications for the Smartnode.",
		configPage.layout.grid,
	)

	return configPage

}

// Get the underlying page
func (configPage *NotificationsConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the monitoring / stats settings page
func (configPage *NotificationsConfigPage) createContent() {

	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "Notifications Settings")

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
	configPage.enableNotificationsBox = createParameterizedCheckbox(&configPage.masterConfig.EnableNotifications)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enableNotificationsBox)

	// Set up the setting callbacks
	configPage.enableNotificationsBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.EnableNotifications.Value == checked {
			return
		}
		configPage.masterConfig.EnableNotifications.Value = checked
		configPage.handleLayoutChanged()
	})

	// Do the initial draw
	configPage.handleLayoutChanged()
}

// Handle all of the form changes when the Enable Notifications box has changed
func (configPage *NotificationsConfigPage) handleLayoutChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enableNotificationsBox.item)

//	if configPage.masterConfig.EnableNotifications.Value == true {
//		configPage.layout.addFormItems([]*parameterizedFormItem{configPage.enableOdaoNotificationsBox, configPage.ecNotificationsPortBox, configPage.bnNotificationsPortBox, configPage.vcNotificationsPortBox, configPage.nodeNotificationsPortBox, configPage.exporterNotificationsPortBox, configPage.watchtowerNotificationsPortBox})
//	}

//	switch configPage.masterConfig.ConsensusClient.Value.(cfgtypes.ConsensusClient) {
//	case cfgtypes.ConsensusClient_Teku, cfgtypes.ConsensusClient_Lighthouse:
//		configPage.layout.form.AddFormItem(configPage.enableBitflyNodeNotificationsBox.item)
//		if configPage.masterConfig.EnableBitflyNodeNotifications.Value == true {
//			configPage.layout.addFormItems(configPage.bitflyNodeNotificationsItems)
//		}
//	}

	configPage.layout.refresh()
}
