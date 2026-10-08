package config

import (
	"slices"

	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

type PerformanceConfigPage struct {
	page   *page
	layout *standardLayout
}

func NewPerformanceConfigPage(home *settingsHome) *PerformanceConfigPage {
	return newPerformanceConfigPage(home.md, home.homePage, "settings-performance")
}

func NewPerformanceConfigPageForNative(home *settingsNativeHome) *PerformanceConfigPage {
	return newPerformanceConfigPage(home.md, home.homePage, "settings-native-performance")
}

func newPerformanceConfigPage(md *MainDisplay, homePage *page, id string) *PerformanceConfigPage {
	layout := newStandardLayout()
	layout.createForm(&md.Config.Smartnode.Network, "Monitoring / Performance Settings")
	layout.setupEscapeReturnHomeHandler(md, homePage)
	items := createParameterizedFormItems(performanceWatcherParameters(md.Config.Smartnode), layout)
	layout.mapParameterizedFormItems(items...)
	layout.addFormItems(items)
	layout.refresh()

	return &PerformanceConfigPage{
		layout: layout,
		page: newPage(
			homePage,
			id,
			"Monitoring / Performance",
			"Select this to configure finalized Rocket Pool validator performance monitoring, including live collection, historical backfill, and retention before Saturn 2.",
			layout.grid,
		),
	}
}

func (configPage *PerformanceConfigPage) getPage() *page {
	return configPage.page
}

func (configPage *PerformanceConfigPage) handleLayoutChanged() {
	configPage.layout.refresh()
}

func performanceWatcherParameters(cfg *config.SmartnodeConfig) []*cfgtypes.Parameter {
	return []*cfgtypes.Parameter{
		&cfg.RunPerformanceWatcher,
		&cfg.PerformanceWatcherBackfill,
		&cfg.PerformanceWatcherBootstrapRetentionDays,
	}
}

// Group these settings in the UI while preserving their existing config keys.
func smartnodeSettingsParameters(cfg *config.SmartnodeConfig) []*cfgtypes.Parameter {
	performance := performanceWatcherParameters(cfg)
	return slices.DeleteFunc(cfg.GetParameters(), func(param *cfgtypes.Parameter) bool {
		return slices.Contains(performance, param)
	})
}
