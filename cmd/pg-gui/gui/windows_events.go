package gui

type ExtendedWindowEvent = string

const (
	CloseSettingsWindow  ExtendedWindowEvent = "settings-window:close"
	SettingsWindowClosed ExtendedWindowEvent = "settings-window:closed"
)
