// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package gui

type ExtendedWindowEvent = string

const (
	CloseSettingsWindow  ExtendedWindowEvent = "settings-window:close"
	SettingsWindowClosed ExtendedWindowEvent = "settings-window:closed"
)
