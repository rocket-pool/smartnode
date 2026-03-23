package config

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

var emailDetailParameters = map[string]interface{}{
	"emailToAddress":  nil,
	"emailFromAddress": nil,
	"emailSmarthost":  nil,
	"emailAuthUser":   nil,
	"emailAuthPass":   nil,
}

var alertingParametersNativeMode map[string]interface{} = map[string]interface{}{
	"enableAlerting":                           nil,
	"showAlertsOnCLI":                          nil,
	"nativeModeHost":                           nil,
	"nativeModePort":                           nil,
	"discordWebhookURL":                        nil,
	"enableEmailAlerts":                        nil,
	"pushoverToken":                            nil,
	"pushoverUserKey":                          nil,
	"alertEnabled_FeeRecipientChanged":         nil,
	"alertEnabled_MinipoolBondReduced":         nil,
	"alertEnabled_MinipoolBalanceDistributed":  nil,
	"alertEnabled_MinipoolPromoted":            nil,
	"alertEnabled_MinipoolStaked":              nil,
	"alertEnabled_ExecutionClientSyncComplete": nil,
	"alertEnabled_BeaconClientSyncComplete":    nil,
	"alertEnabled_PortConnectivityCheck":       nil,
	"alertEnabled_LowETHBalance":               nil,
	"lowETHBalanceThreshold":                   nil,
}

var alertingParametersDockerMode map[string]interface{} = map[string]interface{}{
	"enableAlerting":                           nil,
	"showAlertsOnCLI":                          nil,
	"port":                                     nil,
	"openPort":                                 nil,
	"containerTag":                             nil,
	"discordWebhookURL":                        nil,
	"enableEmailAlerts":                        nil,
	"pushoverToken":                            nil,
	"pushoverUserKey":                          nil,
	"alertEnabled_ClientSyncStatusBeacon":      nil,
	"alertEnabled_ClientSyncStatusExecution":   nil,
	"alertEnabled_UpcomingSyncCommittee":       nil,
	"alertEnabled_ActiveSyncCommittee":         nil,
	"alertEnabled_UpcomingProposal":            nil,
	"alertEnabled_RecentProposal":              nil,
	"alertEnabled_LowDiskSpaceWarning":         nil,
	"alertEnabled_LowDiskSpaceCritical":        nil,
	"alertEnabled_OSUpdatesAvailable":          nil,
	"alertEnabled_RPUpdatesAvailable":          nil,
	"alertEnabled_FeeRecipientChanged":         nil,
	"alertEnabled_MinipoolBondReduced":         nil,
	"alertEnabled_MinipoolBalanceDistributed":  nil,
	"alertEnabled_MinipoolPromoted":            nil,
	"alertEnabled_MinipoolStaked":              nil,
	"alertEnabled_ExecutionClientSyncComplete": nil,
	"alertEnabled_BeaconClientSyncComplete":    nil,
	"alertEnabled_PortConnectivityCheck":       nil,
	"alertEnabled_LowETHBalance":               nil,
	"lowETHBalanceThreshold":                   nil,
}

// The page wrapper for the alerting config
type AlertingConfigPage struct {
	mainDisplay          *mainDisplay
	homePage             *page
	page                 *page
	layout               *standardLayout
	masterConfig         *config.RocketPoolConfig
	alertingEnabledItem  parameterizedFormItem
	emailEnabledItem     *parameterizedFormItem
	emailDetailItems     []*parameterizedFormItem
	otherItems           []*parameterizedFormItem
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
	allItems := createParameterizedFormItems(configPage.masterConfig.Alertmanager.GetParameters(), configPage.layout.descriptionBox)

	// Map the config parameters to the UI form items:
	configPage.layout.mapParameterizedFormItems(allItems...)

	var enableAlertingBox *tview.Checkbox
	var enableEmailBox *tview.Checkbox
	for _, item := range allItems {
		if item.parameter.ID == "enableAlerting" {
			configPage.alertingEnabledItem = *item
			enableAlertingBox = item.item.(*tview.Checkbox)
			continue
		}
		if item.parameter.ID == "enableEmailAlerts" {
			configPage.emailEnabledItem = item
			enableEmailBox = item.item.(*tview.Checkbox)
			continue
		}
		if _, isEmailDetail := emailDetailParameters[item.parameter.ID]; isEmailDetail {
			configPage.emailDetailItems = append(configPage.emailDetailItems, item)
			continue
		}
		_, isNativeParameter := alertingParametersNativeMode[item.parameter.ID]
		_, isDockerParameter := alertingParametersDockerMode[item.parameter.ID]
		if (configPage.masterConfig.IsNativeMode && isNativeParameter) || (!configPage.masterConfig.IsNativeMode && isDockerParameter) {
			configPage.otherItems = append(configPage.otherItems, item)
		}
	}

	if enableAlertingBox != nil {
		enableAlertingBox.SetChangedFunc(func(checked bool) {
			if configPage.masterConfig.Alertmanager.EnableAlerting.Value == checked {
				return
			}
			configPage.masterConfig.Alertmanager.EnableAlerting.Value = checked
			configPage.handleLayoutChanged()
		})
	} else {
		fmt.Println("Error: enableAlerting checkbox not found in alertmanagerItems")
	}

	if enableEmailBox != nil {
		enableEmailBox.SetChangedFunc(func(checked bool) {
			if configPage.masterConfig.Alertmanager.EnableEmailAlerts.Value == checked {
				return
			}
			configPage.masterConfig.Alertmanager.EnableEmailAlerts.Value = checked
			configPage.handleLayoutChanged()
		})
	}

	// Do the initial draw
	configPage.handleLayoutChanged()
}

func (configPage *AlertingConfigPage) handleLayoutChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.addFormItems([]*parameterizedFormItem{&configPage.alertingEnabledItem})
	if configPage.masterConfig.Alertmanager.EnableAlerting.Value == true {
		configPage.layout.addFormItems(configPage.otherItems)
		if configPage.emailEnabledItem != nil {
			configPage.layout.addFormItems([]*parameterizedFormItem{configPage.emailEnabledItem})
			if configPage.masterConfig.Alertmanager.EnableEmailAlerts.Value == true {
				configPage.layout.addFormItems(configPage.emailDetailItems)
			}
		}
	}
	configPage.layout.refresh()
}
