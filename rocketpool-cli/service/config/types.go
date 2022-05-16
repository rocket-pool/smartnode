package config

type settingsPage interface {
	handleLayoutChanged()
	getPage() *page
}

type wizardStep interface {
	show()
}
