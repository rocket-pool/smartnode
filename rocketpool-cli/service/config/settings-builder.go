package config

import (
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

type BuilderConfigPage struct {
	epbsPage *EpbsPage
	page     *page
	layout   *standardLayout
	items    map[*cfgtypes.Parameter]*parameterizedFormItem
}

func NewBuilderConfigPage(epbsPage *EpbsPage) *BuilderConfigPage {
	configPage := &BuilderConfigPage{
		epbsPage: epbsPage,
		layout:   newStandardLayout(),
		items:    map[*cfgtypes.Parameter]*parameterizedFormItem{},
	}
	cfg := epbsPage.home.md.Config
	configPage.layout.createForm(&cfg.Smartnode.Network, "Builder Settings")
	configPage.layout.setupEscapeReturnHomeHandler(epbsPage.home.md, epbsPage.page)
	items := createParameterizedFormItems(cfg.Epbs.GetParameters(), configPage.layout)
	configPage.layout.mapParameterizedFormItems(items...)
	for _, item := range items {
		configPage.items[item.parameter] = item
	}
	configPage.page = newPage(
		epbsPage.page,
		"settings-builder",
		"Builder Settings",
		"Set the minimum bid, maximum execution payment, and builder boost factor, and select builders for the ePBS builder list. These settings are shared by all consensus clients.",
		configPage.layout.grid,
	)
	configPage.handleLayoutChanged()
	return configPage
}

func (configPage *BuilderConfigPage) getPage() *page {
	return configPage.page
}

func (configPage *BuilderConfigPage) handleLayoutChanged() {
	cfg := configPage.epbsPage.home.md.Config.Epbs
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.items[&cfg.MinBid].item)
	configPage.layout.form.AddFormItem(configPage.items[&cfg.MaxExecutionPayment].item)
	configPage.layout.form.AddFormItem(configPage.items[&cfg.BuilderBoostFactor].item)
	for _, param := range cfg.GetAvailableBuilderParameters() {
		configPage.layout.form.AddFormItem(configPage.items[param].item)
	}
	configPage.layout.refresh()
}
