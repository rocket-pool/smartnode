package config

import (
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The page wrapper for the API config
type ApiConfigPage struct {
	mainDisplay  *MainDisplay
	homePage     *page
	page         *page
	layout       *standardLayout
	masterConfig *config.RocketPoolConfig
}

func NewApiConfigPage(home *settingsHome) *ApiConfigPage {
	configPage := &ApiConfigPage{
		mainDisplay:  home.md,
		homePage:     home.homePage,
		masterConfig: home.md.Config,
	}
	configPage.createContent()
	configPage.initPage(false)
	return configPage
}

func NewApiConfigPageForNative(home *settingsNativeHome) *ApiConfigPage {
	configPage := &ApiConfigPage{
		mainDisplay:  home.md,
		homePage:     home.homePage,
		masterConfig: home.md.Config,
	}
	configPage.createContent()
	configPage.initPage(true)
	return configPage
}

func (configPage *ApiConfigPage) initPage(isNative bool) {
	id := "settings-api"
	if isNative {
		id = "settings-api-native"
	}
	configPage.page = newPage(
		configPage.homePage,
		id,
		"API",
		"Select this to configure the Smart Node HTTP API, including the listen port, how it is exposed, the bearer token, unauthenticated reads, and the request rate limit.",
		configPage.layout.grid,
	)
}

func (configPage *ApiConfigPage) getPage() *page {
	return configPage.page
}

func (configPage *ApiConfigPage) createContent() {
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "API Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.mainDisplay, configPage.homePage)

	configPage.masterConfig.IsCLI = true
	_ = configPage.masterConfig.SyncAPITokenFromDisk()

	items := createParameterizedFormItems(configPage.masterConfig.Api.GetParameters(), configPage.layout)
	configPage.layout.mapParameterizedFormItems(items...)
	configPage.layout.addFormItems(items)
	configPage.layout.refresh()
}

func (configPage *ApiConfigPage) handleLayoutChanged() {
	configPage.layout.refresh()
}
