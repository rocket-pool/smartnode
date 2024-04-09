package config

import (
	"fmt"

	"github.com/rivo/tview"
	nmc_ids "github.com/rocket-pool/node-manager-core/config/ids"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/config/ids"
)

var alertingParametersNativeMode map[string]interface{} = map[string]interface{}{
	ids.AlertmanagerEnableAlertingID:              nil,
	ids.AlertmanagerNativeModeHostID:              nil,
	ids.AlertmanagerNativeModePortID:              nil,
	ids.AlertmanagerDiscordWebhookUrlID:           nil,
	ids.AlertmanagerFeeRecipientChangedID:         nil,
	ids.AlertmanagerMinipoolBondReducedID:         nil,
	ids.AlertmanagerMinipoolBalanceDistributedID:  nil,
	ids.AlertmanagerMinipoolPromotedID:            nil,
	ids.AlertmanagerMinipoolStakedID:              nil,
	ids.AlertmanagerExecutionClientSyncCompleteID: nil,
	ids.AlertmanagerBeaconClientSyncCompleteID:    nil,
}

var alertingParametersDockerMode map[string]interface{} = map[string]interface{}{
	ids.AlertmanagerEnableAlertingID:              nil,
	nmc_ids.PortID:                                nil,
	nmc_ids.OpenPortID:                            nil,
	nmc_ids.ContainerTagID:                        nil,
	ids.AlertmanagerDiscordWebhookUrlID:           nil,
	ids.AlertmanagerClientSyncStatusBeaconID:      nil,
	ids.AlertmanagerUpcomingSyncCommitteeID:       nil,
	ids.AlertmanagerActiveSyncCommitteeID:         nil,
	ids.AlertmanagerUpcomingProposalID:            nil,
	ids.AlertmanagerRecentProposalID:              nil,
	ids.AlertmanagerLowDiskSpaceWarningID:         nil,
	ids.AlertmanagerLowDiskSpaceCriticalID:        nil,
	ids.AlertmanagerOSUpdatesAvailableID:          nil,
	ids.AlertmanagerRPUpdatesAvailableID:          nil,
	ids.AlertmanagerFeeRecipientChangedID:         nil,
	ids.AlertmanagerMinipoolBondReducedID:         nil,
	ids.AlertmanagerMinipoolBalanceDistributedID:  nil,
	ids.AlertmanagerMinipoolPromotedID:            nil,
	ids.AlertmanagerMinipoolStakedID:              nil,
	ids.AlertmanagerExecutionClientSyncCompleteID: nil,
	ids.AlertmanagerBeaconClientSyncCompleteID:    nil,
}

// The page wrapper for the alerting config
type AlertingConfigPage struct {
	mainDisplay         *mainDisplay
	homePage            *page
	page                *page
	layout              *standardLayout
	masterConfig        *config.SmartNodeConfig
	alertingEnabledItem parameterizedFormItem
	otherItems          []*parameterizedFormItem
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
	configPage.layout.createForm(&configPage.masterConfig.Network, "Alerting Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.mainDisplay, configPage.homePage)

	// Set up the UI components
	allItems := createParameterizedFormItems(configPage.masterConfig.Alertmanager.GetParameters(), configPage.layout.descriptionBox)

	// Map the config parameters to the UI form items:
	configPage.layout.mapParameterizedFormItems(allItems...)

	var enableAlertingBox *tview.Checkbox = nil
	for _, item := range allItems {
		id := item.parameter.GetCommon().ID
		if id == ids.AlertmanagerEnableAlertingID {
			configPage.alertingEnabledItem = *item
			enableAlertingBox = item.item.(*tview.Checkbox)
			continue
		}
		_, isNativeParameter := alertingParametersNativeMode[id]
		_, isDockerParameter := alertingParametersDockerMode[id]
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
		fmt.Printf("Error: %s checkbox not found in alertmanagerItems\n", ids.AlertmanagerEnableAlertingID)
	}

	// Do the initial draw
	configPage.handleLayoutChanged()
}

// Handle all of the form changes when the Enable Metrics box has changed
func (configPage *AlertingConfigPage) handleLayoutChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.addFormItems([]*parameterizedFormItem{&configPage.alertingEnabledItem})
	if configPage.masterConfig.Alertmanager.EnableAlerting.Value {
		configPage.layout.addFormItems(configPage.otherItems)
	}
	configPage.layout.refresh()
}
