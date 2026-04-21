package config

type settingsPage interface {
	handleLayoutChanged()
	getPage() *page
}
