package config

import (
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The page wrapper for the alerting config
type AlertingConfigPage struct {
	mainDisplay       *mainDisplay
	homePage          *page
	page              *page
	layout            *standardLayout
	masterConfig      *config.RocketPoolConfig
	alertmanagerItems []*parameterizedFormItem
}

func NewAlertingConfigPage(home *settingsHome) *AlertingConfigPage {
	configPage := &AlertingConfigPage{
		mainDisplay:  home.md,
		homePage:     home.homePage,
		masterConfig: home.md.Config,
	}

	configPage.createContent()
	configPage.initPage(false)

	return configPage
}

func NewAlertingConfigPageForNative(home *settingsNativeHome) *AlertingConfigPage {
	configPage := &AlertingConfigPage{
		mainDisplay:  home.md,
		homePage:     home.homePage,
		masterConfig: home.md.Config,
	}

	configPage.createContent()
	configPage.initPage(true)

	return configPage
}

func (configPage *AlertingConfigPage) initPage(isNative bool) {
	id := "settings-alerting"
	if isNative {
		id = "settings-alerting-native"
	}
	configPage.page = newPage(
		configPage.homePage,
		id,
		"Monitoring / Alerting",
		"Select this to configure the alerting of the Smartnode. Requires metrics to be enabled.",
		configPage.layout.grid,
	)
}

func (configPage *AlertingConfigPage) getPage() *page {
	return configPage.page
}

// Creates the UI form items of the alerting config page.
func (configPage *AlertingConfigPage) createContent() {
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "Alerting Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.mainDisplay, configPage.homePage)

	// Set up the UI components
	configPage.alertmanagerItems = createParameterizedFormItems(configPage.masterConfig.Alertmanager.GetParameters(), configPage.layout.descriptionBox)

	// Map the config parameters to the UI form items:
	configPage.layout.mapParameterizedFormItems(configPage.alertmanagerItems...)

	// Do the initial draw
	configPage.handleLayoutChanged()
}

// Handle all of the form changes when the Enable Metrics box has changed
func (configPage *AlertingConfigPage) handleLayoutChanged() {
	configPage.layout.form.Clear(true)
	if configPage.masterConfig.EnableMetrics.Value == true {
		configPage.layout.addFormItems(configPage.alertmanagerItems)
	}
	configPage.layout.refresh()
}
